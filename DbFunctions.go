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
func TruncateTraceFiles(lc chan<- LogMessage, name string, hdb *sql.DB, TrncDaysOlder uint) error {
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
	lc <- LogMessage{name, fmt.Sprintf("Attempting Query:'%s'\n", GetTraceFileQuery(TrncDaysOlder)), true}

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

		lc <- LogMessage{name, fmt.Sprintf("Attempting Query'%s'\n", GetRemoveTrace(v.Hostname, v.TraceFile)), true}
		/*do nothing destructive if dryrun enabled*/
		if !ac.DryRun {
			_, err := hdb.Exec(GetRemoveTrace(v.Hostname, v.TraceFile))
			if err != nil {
				log.Println(err.Error())
				log.Printf("%s:The tracefile '%s' on host '%s' could not be removed, it may be open!  This will be retried next time.\n", name, v.TraceFile, v.Hostname)
				continue
			}

			count += 1
			saved += v.SizeBytes
		}
	}

	if count > 0 {
		log.Printf("%s:Removed %d old tracefiles saving %.2f MiB", name, count, float64(saved/1024/1024))
	}
	return nil
}

//This function truncates the backup catalog of the HANA database with the option of also deleting the underlying database backup files.
//The options in the configuration file that control this are:
//TruncateBackupCatalog - bool.  If true, the function will be triggered.
//BackupCatalogRetentionDays - uint, used to decide which backup catalog entries to retain
//DeleteOldBackups - bool, if false only the catalog entries will be removed, if true the removed backup catalog entries will be deleted from the file system or BACKINT, use with caution
func TruncateBackupCatalog(ac AppConfig, name string, hdb *sql.DB, TrncDaysOlder uint, delete bool) error {

	if ac.Verbose {
		log.Printf("%s:Attempting to truncate the backup catalog", name)
	}
	/*Find the backup ID of the latest full backup that matches the */
	var backupID string
	err := hdb.QueryRow(GetLatestFullBackupID(TrncDaysOlder)).Scan(&backupID)
	switch {
	case err == sql.ErrNoRows:
		log.Printf("%s:No backupID found which matches the criteria", name)
		return nil
	case err != nil:
		log.Printf("%s:An error occured querying the database.", name)
		log.Printf("%s:%s", name, err.Error())
		return fmt.Errorf("query error")

	default:
		log.Printf("%s:Found recent backupID (%s) that meets the serach criteria\n", name, backupID)
	}

	/*Count how many backups will be deleted*/
	bfs := []BackupFiles{}
	rows, err := hdb.Query(GetBackupFileData(backupID))
	if err != nil {
		log.Printf("%s:An error occured querying the database.", name)
		log.Printf("%s:%s", name, err.Error())
		return fmt.Errorf("failed to retrieve data on backup catalog entries to remove")
	}
	defer rows.Close()
	for rows.Next() {
		bf := BackupFiles{}
		err := rows.Scan(&bf.EntryType, &bf.FileCount, &bf.Bytes)
		if err != nil {
			log.Printf("%s:An error occured querying the database.", name)
		}
		bfs = append(bfs, bf)
	}

	if len(bfs) == 0 {
		/*Looks like we found a backup, but it is the oldest backup in the catalog so we have nothing to delete*/
		/*bail out here gracefully*/
		log.Printf("%s:Nothing to delete older than %s\n", name, backupID)
		return nil
	}

	/*print some info in the log*/
	for _, v := range bfs {
		if delete {
			log.Printf("%s:Will remove %d %s files from the catalog, this will free %.2f MiB of space", name, v.FileCount, v.EntryType, float32(v.Bytes/1024/1024))
		} else {
			log.Printf("%s:Will remove %d %s files from the catalog", name, v.FileCount, v.EntryType)
		}
	}

	/*do the truncation*/
	var query string
	if delete {
		query = GetBackupDeleteComplete(backupID)
	} else {
		query = GetBackupDelete(backupID)
	}
	if ac.Verbose {
		log.Printf("%s:Attempting query %s\n", name, query)
	}

	/*do nothing destructive if dryrun enabled*/
	if !ac.DryRun {
		_, err = hdb.Exec(query)
		if err != nil {
			log.Printf("%s:Query to truncate backup catalog failed: %s", name, err.Error())
			return fmt.Errorf("couldn't truncate database")
		}

		log.Printf("%s:Backup catalog sucessfully truncated", name)
	}

	return nil
}
