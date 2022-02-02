package main

import (
	"database/sql"
	"fmt"
	"strings"
)

//HanaVersion function returns the version string of the database
func HanaVersion(name string, lc chan<- LogMessage, hdb *sql.DB) (string, error) {
	fname := fmt.Sprintf("%s:%s", name, "HanaVersion")
	var version string
	lc <- LogMessage{fname, "Starting", true}
	lc <- LogMessage{fname, fmt.Sprintf("Performing query: %s", QUERY_GetVersion), true}
	r1 := hdb.QueryRow(QUERY_GetVersion)
	err := r1.Scan(&version)
	if err != nil {
		lc <- LogMessage{fname, "Query failed", false}
		lc <- LogMessage{fname, err.Error(), false}
		return "", err // allow calling function to handle the error
	}
	lc <- LogMessage{fname, "OK", true}
	return version, nil
}

//TruncateTraceFiles function removes closed trace files that are older than the number of days
//specified in the 'TrncDaysOlder' argument.  The function will log all activity.  The function will also return
//an error.  If no errors are found nil is returned.
//In some cases it may not be possible to remove a trace file, these incidents are logged but will not cause the function to error.
func TruncateTraceFiles(lc chan<- LogMessage, name string, hdb *sql.DB, TrncDaysOlder uint, dryrun bool) error {
	fname := fmt.Sprintf("%s:%s", name, "TruncateTraceFiles")
	lc <- LogMessage{fname, "Starting", true}
	if dryrun {
		lc <- LogMessage{fname, "Dry run enabled, no changes will be made", false}
	}

	/*This doesn't look right, you could set it to zero and have all history prior to the last full back removed*/
	if TrncDaysOlder == 0 {
		lc <- LogMessage{fname, "TrncDaysOlder is set to zero, nothing to do", false}
		return nil
	}

	/*slilc <- LogMessage{name, "Attempting tracefile truncation", true}
	ce to hold results*/
	TraceFiles := make([]TraceFile, 0)

	/*Get the list of candidate tracefiles where the M time days is greater than the TrncDaysOlder arguments*/
	//fmt.Printf("%s", GetTraceFileQuery(TrncDaysOlder))
	lc <- LogMessage{fname, fmt.Sprintf("Performing query:'%s'", GetTraceFileQuery(TrncDaysOlder)), true}

	rows, err := hdb.Query(GetTraceFileQuery(TrncDaysOlder))
	if err != nil {
		lc <- LogMessage{fname, "Query Failed", true}
		lc <- LogMessage{fname, err.Error(), true}
		/*allow calling function to deal with error*/
		return err
	}
	defer rows.Close()

	for rows.Next() {
		tf := TraceFile{}
		err := rows.Scan(&tf.Hostname, &tf.TraceFile, &tf.SizeBytes, &tf.LastModified)
		if err != nil {
			lc <- LogMessage{fname, "Scan Error", true}
			lc <- LogMessage{fname, err.Error(), true}
			/*allow calling function to deal with the error*/
			return err
		}
		TraceFiles = append(TraceFiles, tf)
	}

	if len(TraceFiles) == 0 {
		lc <- LogMessage{fname, "No tracefiles meet criteria for removal", false}
		return nil
	}

	var count uint = 0
	var saved uint64 = 0
	/*Try and remove the files one by one to increase clarity in the logs*/
	for _, v := range TraceFiles {

		lc <- LogMessage{fname, fmt.Sprintf("Performing Query'%s'", GetRemoveTrace(v.Hostname, v.TraceFile)), true}
		/*do nothing destructive if dryrun enabled*/
		if !dryrun {
			_, err := hdb.Exec(GetRemoveTrace(v.Hostname, v.TraceFile))
			if err != nil {
				lc <- LogMessage{fname, fmt.Sprintf("The tracefile '%s' on host '%s' could not be removed, it may be open!  This will be retried next time.", v.TraceFile, v.Hostname), true}
				lc <- LogMessage{fname, err.Error(), true}
				continue

			}
			//Check if the trace file was actually deleted
			var tracePresent uint = 0

			lc <- LogMessage{fname, "Checking if tracefile was removed", true}
			lc <- LogMessage{fname, fmt.Sprint("Performing Query:", v.TraceFile), true}
			err = hdb.QueryRow(GetCheckTracePresent(v.TraceFile)).Scan(&tracePresent)
			switch {
			case err == sql.ErrNoRows:
				lc <- LogMessage{fname, "No rows returned", true}
				lc <- LogMessage{fname, err.Error(), true}
				lc <- LogMessage{fname, fmt.Sprintf("Failed to remove tracefile %s", v.TraceFile), false}
				continue /*try the next one*/
			case err != nil:
				lc <- LogMessage{fname, "DB failed to query failed", true}
				lc <- LogMessage{fname, err.Error(), true}
				lc <- LogMessage{fname, fmt.Sprintf("Failed to remove tracefile %s", v.TraceFile), false}
				continue /*try the next one*/
			default:
				if tracePresent == 0 {
					lc <- LogMessage{fname, fmt.Sprintf("Sucessfully removed trace file %s", v.TraceFile), true}
				} else { /*for trace files we should only ever see 0 or 1*/
					lc <- LogMessage{fname, fmt.Sprintf("Tracefile %s was not removed", v.TraceFile), true}
					continue
				}
			}
			count += 1
			saved += v.SizeBytes

		}
	}

	if count > 0 {
		lc <- LogMessage{fname, fmt.Sprintf("Removed %d old tracefiles saving %.2f MiB", count, float64(saved/1024/1024)), false}
	} else {
		lc <- LogMessage{fname, "Nothing was removed", false}
	}
	return nil
}

//This function truncates the backup catalog of the HANA database with the option of also deleting the underlying database backup files.
//The options in the configuration file that control this are:
//TruncateBackupCatalog - bool.  If true, the function will be triggered.
//BackupCatalogRetentionDays - uint, used to decide which backup catalog entries to retain
//DeleteOldBackups - bool, if false only the catalog entries will be removed, if true the removed backup catalog entries will be deleted from the file system or BACKINT, use with caution
func TruncateBackupCatalog(lc chan<- LogMessage, name string, hdb *sql.DB, TrncDaysOlder uint, delete bool, dryrun bool) error {
	fname := fmt.Sprintf("%s:%s", name, "TruncateBackupCatalog")
	lc <- LogMessage{fname, "Starting", true}
	if dryrun {
		lc <- LogMessage{fname, "Dry run enabled, no changes will be made", false}
	}

	/*Find the backup ID of the latest full backup that matches the */
	var backupID string
	lc <- LogMessage{fname, fmt.Sprintf("Performing Query: %s", GetLatestFullBackupID(TrncDaysOlder)), true}
	err := hdb.QueryRow(GetLatestFullBackupID(TrncDaysOlder)).Scan(&backupID)
	switch {
	case err == sql.ErrNoRows:
		lc <- LogMessage{fname, "No backupID found which matches the criteria", false}
		return nil
	case err != nil:
		lc <- LogMessage{fname, "An error occured querying the database", false}
		lc <- LogMessage{fname, err.Error(), true}
		return fmt.Errorf("query error")

	default:
		lc <- LogMessage{fname, fmt.Sprintf("Found recent backupID (%s) that meets the serach criteria", backupID), true}
	}

	/*Count how many backups will be deleted*/
	bfs := []BackupFiles{}
	lc <- LogMessage{fname, fmt.Sprintf("Performing Query: %s", GetBackupFileData(backupID)), true}
	rows, err := hdb.Query(GetBackupFileData(backupID))
	if err != nil {
		lc <- LogMessage{fname, "An error occured querying the database", false}
		lc <- LogMessage{fname, err.Error(), true}
		return fmt.Errorf("failed to retrieve data on backup catalog entries to remove")
	}
	defer rows.Close()
	for rows.Next() {
		bf := BackupFiles{}
		err := rows.Scan(&bf.EntryType, &bf.FileCount, &bf.Bytes) //need unit test here!
		if err != nil {
			lc <- LogMessage{fname, "An error occured querying the database", false}
			lc <- LogMessage{fname, err.Error(), true}
		}
		bfs = append(bfs, bf)
	}

	if len(bfs) == 0 {
		/*Looks like we found a backup, but it is the oldest backup in the catalog so we have nothing to delete*/
		/*bail out here gracefully*/
		lc <- LogMessage{fname, fmt.Sprintf("Nothing to delete older than %s", backupID), false}

		return nil
	}

	/*print some info in the log*/
	for _, v := range bfs {
		if delete {
			lc <- LogMessage{fname, fmt.Sprintf("Will remove %d %s files from the catalog, this will free %.2f MiB of space", v.FileCount, v.EntryType, float32(v.Bytes/1024/1024)), false}
		} else {
			lc <- LogMessage{fname, fmt.Sprintf("Will remove %d %s files from the catalog", v.FileCount, v.EntryType), false}

		}
	}

	/*do the truncation*/
	var query string
	if delete {
		query = GetBackupDeleteComplete(backupID)
	} else {
		query = GetBackupDelete(backupID)
	}
	lc <- LogMessage{fname, fmt.Sprintf("Performing query: %s", query), true}

	if !dryrun {
		_, err = hdb.Exec(query)
		if err != nil {
			lc <- LogMessage{fname, "Query failed", false}
			lc <- LogMessage{fname, err.Error(), true}
			return fmt.Errorf("couldn't clean backup catalog")
		}

		lc <- LogMessage{fname, "Backup catalog sucessfully cleaned", false}
	}
	return nil
}

//This function deletes alerts from the table _SYS_STATISTICS.STATISTICS_ALERTS_BASE.  Alerts are deleted if they are older than
//the given number of days in the DeleteOlderDays argument.  No changes are made to the database if the dryrun argument is set to true
func ClearAlert(lc chan<- LogMessage, name string, hdb *sql.DB, DeleteOlderDays uint, dryrun bool) error {
	fname := fmt.Sprintf("%s:%s", name, "ClearAlert")
	lc <- LogMessage{fname, "Starting", true}
	if dryrun {
		lc <- LogMessage{fname, "Dry run enabled, no changes will be made", false}
	}
	/*Find how many alerts there are that match the deltion criteria*/
	var ac uint
	lc <- LogMessage{fname, fmt.Sprintf("Performing query:%s", GetAlertCount(DeleteOlderDays)), true}
	err := hdb.QueryRow(GetAlertCount(DeleteOlderDays)).Scan(&ac)
	switch {
	case err == sql.ErrNoRows:
		lc <- LogMessage{fname, "DB failed to count rows", false}
		lc <- LogMessage{fname, err.Error(), true}
		return err
	case err != nil:
		lc <- LogMessage{fname, "DB failed to query failed", false}
		lc <- LogMessage{fname, err.Error(), true}

		return err
	default:
		lc <- LogMessage{fname, fmt.Sprintf("Found %d alerts to delete ", ac), true}
	}

	if ac == 0 {
		lc <- LogMessage{fname, "No alerts met the criteria for removal", false}
		return nil
	}

	/*Attempt to delete the records*/
	if !dryrun {
		_, err = hdb.Exec(GetAlertDelete(DeleteOlderDays))
		if err != nil {
			lc <- LogMessage{name, "Query to remove alerts failed", false}
			lc <- LogMessage{name, err.Error(), true}
			return err
		}
		lc <- LogMessage{name, fmt.Sprintf("Sucessfully deleted %d alerts", ac), false}
	}
	return nil
}

//This function deletes free logsegments from the log volume.  Performing this task will reduce the disk space used in the log volume
//but may also cause a minor IO penalty when new new log segments need to be created.  It is more important to run this function is an MDC
//environemt than a non-MDC one.
func ReclaimLog(lc chan<- LogMessage, name string, hdb *sql.DB, dryrun bool) error {
	fname := fmt.Sprintf("%s:%s", name, "ReclaimLog")
	lc <- LogMessage{fname, "Starting", true}
	if dryrun {
		lc <- LogMessage{fname, "Dry run enabled, no changes will be made", false}
	}
	var count uint
	var bytes uint64
	lc <- LogMessage{fname, fmt.Sprintf("Performing Query:%s", QUERY_GetFeeLogSegments), true}
	err := hdb.QueryRow(QUERY_GetFeeLogSegments).Scan(&count, &bytes)
	switch {
	case err == sql.ErrNoRows:
		lc <- LogMessage{fname, "No rows produced by query", false}
		lc <- LogMessage{fname, err.Error(), true}
		return fmt.Errorf("no results")
	case err != nil:
		lc <- LogMessage{fname, "Query produced a database error", false}
		lc <- LogMessage{fname, err.Error(), true}
		return fmt.Errorf("db error")
	}

	lc <- LogMessage{fname, fmt.Sprintf("Attempting to clear %d log segments saving %.2f MiB of disk space", count, float32(bytes/1024/1024)), true}

	if !dryrun {
		lc <- LogMessage{fname, fmt.Sprintf("Performing Query:%s", QUERY_RecalimLog), true}
		_, err = hdb.Exec(QUERY_RecalimLog)
		if err != nil {
			lc <- LogMessage{fname, "Query produced a database error", false}
			lc <- LogMessage{fname, err.Error(), true}
			return fmt.Errorf("db error")
		}

		lc <- LogMessage{fname, "Log Reclaim was sucessful.", false}
	}

	return nil
}

//Function that removes old audit events from the audit table.
func TruncateAuditLog(lc chan<- LogMessage, name string, hdb *sql.DB, daysToKeep uint, dryrun bool) error {

	fname := fmt.Sprintf("%s:%s", name, "TruncateAuditLog")
	lc <- LogMessage{fname, "Starting", true}
	if dryrun {
		lc <- LogMessage{fname, "Dry run enabled, no changes will be made", false}
	}

	//Get the number of items to be removed
	lc <- LogMessage{fname, fmt.Sprintf("Performing query:%s", GetAuditCount(daysToKeep)), true}
	var auditCount uint
	err := hdb.QueryRow(GetAuditCount(daysToKeep)).Scan(&auditCount)
	switch {
	case err == sql.ErrNoRows:
		lc <- LogMessage{fname, "No rows produced by query", false}
		return fmt.Errorf("no results")
	case err != nil:
		lc <- LogMessage{fname, "Query produced a database error", false}
		lc <- LogMessage{fname, err.Error(), true}
		return fmt.Errorf("db error")
	}

	switch {
	case auditCount == 0:
		lc <- LogMessage{fname, "No audit records found that meet deletion criteria", false}
		return nil
	case auditCount == 1:
		lc <- LogMessage{fname, "1 audit record found that meet deletion criteria", true}
	case auditCount > 1:
		lc <- LogMessage{fname, fmt.Sprintf("%d audit records found that meet deletion criteria", auditCount), true}
	}

	/*For whatever reason, HANA 2.0 SPS5 doesn't like taking a subquery as the timestamp argument
	in the ALTER SYSTEM CLEAR AUDIT LOG UNIT .... command.
	Therefore we need to pass the time argument in a string.  We don't want to use the local time as the DB could be differnt
	So we'll run a query the DB for the time and feed it back in.*/
	lc <- LogMessage{fname, fmt.Sprintf("Performing Query:%s", GetDatetime(daysToKeep)), true}
	var dateString string
	err = hdb.QueryRow(GetDatetime(daysToKeep)).Scan(&dateString)
	switch {
	case err == sql.ErrNoRows:
		lc <- LogMessage{fname, "No rows produced by query", false}
		return fmt.Errorf("no results")
	case err != nil:
		lc <- LogMessage{fname, "Scan error or query produced a database error", false}
		lc <- LogMessage{fname, err.Error(), true}
		return fmt.Errorf("db error")
	}

	//Get rid of subseconds everything after and including the period
	dateParts := strings.Split(dateString, ".")
	/*ensure we have at least two elements in the array*/
	if len(dateParts) != 2 {
		lc <- LogMessage{fname, fmt.Sprintf("The date string %s retrieved from the database couldn't be split to 2 parts.  Expect 2, got %d", dateString, len(dateParts)), false}
		return fmt.Errorf("couldn't split string")
	}

	if !dryrun {
		lc <- LogMessage{name, fmt.Sprintf("Performing Query:%s", GetTruncateAuditLog(dateParts[0])), true}
		_, err = hdb.Exec(GetTruncateAuditLog(dateParts[0]))
		if err != nil {
			lc <- LogMessage{name, "Clean audit log query failed", false}
			lc <- LogMessage{name, err.Error(), true}
			return fmt.Errorf("db error")
		}
	}

	return nil
}
