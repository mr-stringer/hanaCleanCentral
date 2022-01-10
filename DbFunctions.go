package main

import (
	"database/sql"
	"fmt"
	"log"
)

//HanaVersion function returns the version string of the database
func HanaVersion(hdb *sql.DB) (string, error) {
	var version string

	r1 := hdb.QueryRow(QUERY_GetVersion)
	err := r1.Scan(&version)
	if err != nil {
		return "", err // allow calling function to handle the error
	}

	return version, nil
}

//TruncateTraceFiles function removes closed trace files that are older than the number of days
//specified in the 'TrncDaysOlder' argument.  The function will log all activity.  The function will also return
//an error.  If no errors are found nil is returned.
//In some cases it may not be possible to remove a trace file, these incidents are logged but will not cause the function to error.
func TruncateTraceFiles(lc chan<- LogMessage, name string, hdb *sql.DB, TrncDaysOlder uint, dryrun bool) error {
	lc <- LogMessage{name, "Attempting tracefile truncation", true}

	if TrncDaysOlder == 0 {
		lc <- LogMessage{name, "TrncDaysOlder is set to zero, nothing to do", false}
		return nil
	}

	/*slilc <- LogMessage{name, "Attempting tracefile truncation", true}
	ce to hold results*/
	TraceFiles := make([]TraceFile, 0)

	/*Get the list of candidate tracefiles where the M time days is greater than the TrncDaysOlder arguments*/
	//fmt.Printf("%s", GetTraceFileQuery(TrncDaysOlder))
	lc <- LogMessage{name, fmt.Sprintf("Attempting Query:'%s'", GetTraceFileQuery(TrncDaysOlder)), true}

	rows, err := hdb.Query(GetTraceFileQuery(TrncDaysOlder))
	if err != nil {
		/*allow calling function to deal with error*/
		return err
	}
	defer rows.Close()

	for rows.Next() {
		tf := TraceFile{}
		err := rows.Scan(&tf.Hostname, &tf.TraceFile, &tf.SizeBytes, &tf.LastModified)
		if err != nil {
			/*allow calling function to deal with the error*/
			return err
		}
		TraceFiles = append(TraceFiles, tf)
	}

	if len(TraceFiles) == 0 {
		lc <- LogMessage{name, "No tracefiles meet criteria for removal", false}
		return nil
	}

	var count uint = 0
	var saved uint64 = 0
	/*Try and remove the files one by one to increase clarity in the logs*/
	for _, v := range TraceFiles {

		lc <- LogMessage{name, fmt.Sprintf("Attempting Query'%s'", GetRemoveTrace(v.Hostname, v.TraceFile)), true}
		/*do nothing destructive if dryrun enabled*/
		if !dryrun {
			_, err := hdb.Exec(GetRemoveTrace(v.Hostname, v.TraceFile))
			if err != nil {
				lc <- LogMessage{name, fmt.Sprintf("The tracefile '%s' on host '%s' could not be removed, it may be open!  This will be retried next time.", v.TraceFile, v.Hostname), false}
				continue
			}

			count += 1
			saved += v.SizeBytes
		}
	}

	if count > 0 {
		lc <- LogMessage{name, fmt.Sprintf("Removed %d old tracefiles saving %.2f MiB", count, float64(saved/1024/1024)), false}
	}
	return nil
}

//This function truncates the backup catalog of the HANA database with the option of also deleting the underlying database backup files.
//The options in the configuration file that control this are:
//TruncateBackupCatalog - bool.  If true, the function will be triggered.
//BackupCatalogRetentionDays - uint, used to decide which backup catalog entries to retain
//DeleteOldBackups - bool, if false only the catalog entries will be removed, if true the removed backup catalog entries will be deleted from the file system or BACKINT, use with caution
func TruncateBackupCatalog(lc chan<- LogMessage, name string, hdb *sql.DB, TrncDaysOlder uint, delete bool, dryrun bool) error {

	lc <- LogMessage{name, "Attempting to truncate the backup catalog", true}

	/*Find the backup ID of the latest full backup that matches the */
	var backupID string
	err := hdb.QueryRow(GetLatestFullBackupID(TrncDaysOlder)).Scan(&backupID)
	switch {
	case err == sql.ErrNoRows:
		lc <- LogMessage{name, "No backupID found which matches the criteria", false}
		return nil
	case err != nil:
		lc <- LogMessage{name, "An error occured querying the database", false}
		lc <- LogMessage{name, err.Error(), false}
		return fmt.Errorf("query error")

	default:
		lc <- LogMessage{name, fmt.Sprintf("Found recent backupID (%s) that meets the serach criteria", backupID), false}
	}

	/*Count how many backups will be deleted*/
	bfs := []BackupFiles{}
	rows, err := hdb.Query(GetBackupFileData(backupID))
	if err != nil {
		lc <- LogMessage{name, "An error occured querying the database", false}
		lc <- LogMessage{name, err.Error(), false}
		return fmt.Errorf("failed to retrieve data on backup catalog entries to remove")
	}
	defer rows.Close()
	for rows.Next() {
		bf := BackupFiles{}
		err := rows.Scan(&bf.EntryType, &bf.FileCount, &bf.Bytes)
		if err != nil {
			lc <- LogMessage{name, "An error occured querying the database", false}
		}
		bfs = append(bfs, bf)
	}

	if len(bfs) == 0 {
		/*Looks like we found a backup, but it is the oldest backup in the catalog so we have nothing to delete*/
		/*bail out here gracefully*/
		lc <- LogMessage{name, fmt.Sprintf("Nothing to delete older than %s", backupID), false}

		return nil
	}

	/*print some info in the log*/
	for _, v := range bfs {
		if delete {
			lc <- LogMessage{name, fmt.Sprintf("Will remove %d %s files from the catalog, this will free %.2f MiB of space", v.FileCount, v.EntryType, float32(v.Bytes/1024/1024)), false}
		} else {
			lc <- LogMessage{name, fmt.Sprintf("Will remove %d %s files from the catalog", v.FileCount, v.EntryType), false}

		}
	}

	/*do the truncation*/
	var query string
	if delete {
		query = GetBackupDeleteComplete(backupID)
	} else {
		query = GetBackupDelete(backupID)
	}
	lc <- LogMessage{name, fmt.Sprintf("Attempting query: %s", query), true}
	log.Printf("%s:Attempting query %s", name, query)

	if !dryrun {
		_, err = hdb.Exec(query)
		if err != nil {
			lc <- LogMessage{name, fmt.Sprintf("Query to truncate backup catalog failed: %s", err.Error()), false}
			return fmt.Errorf("couldn't truncate database")
		}

		lc <- LogMessage{name, "Backup catalog sucessfully truncated", false}
	}
	return nil
}

//This function deletes alerts from the table _SYS_STATISTICS.STATISTICS_ALERTS_BASE.  Alerts are deleted if they are older than
//the given number of days in the DeleteOlderDays argument.  No changes are made to the database if the dryrun argument is set to true
func ClearAlert(lc chan<- LogMessage, name string, hdb *sql.DB, DeleteOlderDays uint, dryrun bool) error {
	lc <- LogMessage{name, "Attempting to Clear Alerts", true}

	/*Find how many alerts there are that match the deltion criteria*/
	var ac uint
	lc <- LogMessage{name, "Attempting query", true}
	lc <- LogMessage{name, GetAlertCount(DeleteOlderDays), true}
	err := hdb.QueryRow(GetAlertCount(DeleteOlderDays)).Scan(&ac)
	switch {
	case err == sql.ErrNoRows:
		lc <- LogMessage{name, "DB failed to count rows", false}
		return err
	case err != nil:
		lc <- LogMessage{name, "DB failed to query failed", false}
		return err
	default:
		lc <- LogMessage{name, fmt.Sprintf("Found %d alerts to delete ", ac), false}
	}

	if ac == 0 {
		lc <- LogMessage{name, "Nothing to delete", false}
		return nil
	}

	/*Attempt to delete the records*/
	if !dryrun {
		_, err = hdb.Exec(GetAlertDelete(DeleteOlderDays))
		if err != nil {
			lc <- LogMessage{name, "Query to remove alerts failed", false}
			lc <- LogMessage{name, err.Error(), false}
			return err
		}
		lc <- LogMessage{name, fmt.Sprintf("Sucessfully deleted %d alerts", ac), false}
	}
	return nil
}
