package main

import (
	"database/sql"
	"fmt"
	"strings"
)

//HanaVersion function returns the version string of the database
func (dbc *DbConfig) HanaVersionFunc(lc chan<- LogMessage) (string, error) {
	fname := fmt.Sprintf("%s:%s", dbc.Name, "HanaVersion")
	var version string
	lc <- LogMessage{fname, "Starting", true}
	lc <- LogMessage{fname, fmt.Sprintf("Performing query: %s", QUERY_GetVersion), true}
	r1 := dbc.db.QueryRow(QUERY_GetVersion)
	err := r1.Scan(&version)
	if err != nil {
		lc <- LogMessage{fname, "Query failed", false}
		lc <- LogMessage{fname, err.Error(), false}
		return "", err // allow calling function to handle the error
	}
	lc <- LogMessage{fname, "OK", true}
	return version, nil
}

//CleanTraceFiles function removes closed trace files that are older than the number of days
//specified in the 'CleanDaysOlder' argument.  The function will log all activity.  The function will also return
//an error.  If no errors are found nil is returned.
//In some cases it may not be possible to remove a trace file, these incidents are logged but will not cause the function to error.
func (dbc *DbConfig) CleanTraceFilesFunc(lc chan<- LogMessage, CleanDaysOlder uint, dryrun bool) error {
	fname := fmt.Sprintf("%s:%s", dbc.Name, "CleanTraceFiles")
	lc <- LogMessage{fname, "Starting", true}
	if dryrun {
		lc <- LogMessage{fname, "Dry run enabled, no changes will be made", false}
	}

	/*This doesn't look right, you could set it to zero and have all history prior to the last full back removed*/
	if CleanDaysOlder == 0 {
		lc <- LogMessage{fname, "CleanDaysOlder is set to zero, nothing to do", false}
		return nil
	}

	/*slice <- LogMessage{name, "Attempting tracefile truncation", true}
	ce to hold results*/
	TraceFiles := make([]TraceFile, 0)

	/*Get the list of candidate tracefiles where the M time days is greater than the TrncDaysOlder arguments*/
	//fmt.Printf("%s", GetTraceFileQuery(TrncDaysOlder))
	lc <- LogMessage{fname, fmt.Sprintf("Performing query:'%s'", GetTraceFileQuery(CleanDaysOlder)), true}

	rows, err := dbc.db.Query(GetTraceFileQuery(CleanDaysOlder))
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
			_, err := dbc.db.Exec(GetRemoveTrace(v.Hostname, v.TraceFile))
			if err != nil {
				lc <- LogMessage{fname, fmt.Sprintf("The tracefile '%s' on host '%s' could not be removed, it may be open!  This will be retried next time.", v.TraceFile, v.Hostname), true}
				lc <- LogMessage{fname, err.Error(), true}
				continue

			}
			//Check if the trace file was actually deleted
			var tracePresent uint = 0

			lc <- LogMessage{fname, "Checking if tracefile was removed", true}
			lc <- LogMessage{fname, fmt.Sprint("Performing Query:", v.TraceFile), true}
			err = dbc.db.QueryRow(GetCheckTracePresent(v.TraceFile)).Scan(&tracePresent)
			switch {
			case err == sql.ErrNoRows:
				lc <- LogMessage{fname, "No rows returned", true}
				lc <- LogMessage{fname, err.Error(), true}
				lc <- LogMessage{fname, fmt.Sprintf("Failed to remove tracefile %s", v.TraceFile), false}
				continue /*try the next one perhaps we should check for how many files couldn't be removed*/
			case err != nil:
				lc <- LogMessage{fname, "DB failed to query failed", true}
				lc <- LogMessage{fname, err.Error(), true}
				lc <- LogMessage{fname, fmt.Sprintf("Failed to remove tracefile %s", v.TraceFile), false}
				continue /*try the next one - don't throw error - perhaps we should check for how many files couldn't be removed*/
			default:
				if tracePresent == 0 {
					lc <- LogMessage{fname, fmt.Sprintf("Successfully removed trace file %s", v.TraceFile), true}
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
func (dbc *DbConfig) CleanBackupFunc(lc chan<- LogMessage, CleanDaysOlder uint, delete bool, dryrun bool) error {
	fname := fmt.Sprintf("%s:%s", dbc.Name, "CleanBackupCatalog")
	lc <- LogMessage{fname, "Starting", true}
	if dryrun {
		lc <- LogMessage{fname, "Dry run enabled, no changes will be made", false}
	}

	/*Find the backup ID of the latest full backup that matches the */
	var backupID string
	lc <- LogMessage{fname, fmt.Sprintf("Performing Query: %s", GetLatestFullBackupID(CleanDaysOlder)), true}
	err := dbc.db.QueryRow(GetLatestFullBackupID(CleanDaysOlder)).Scan(&backupID)
	switch {
	case err == sql.ErrNoRows:
		lc <- LogMessage{fname, "No backupID found which matches the criteria", false}
		return nil
	case err != nil:
		lc <- LogMessage{fname, "An error occurred querying the database", false}
		lc <- LogMessage{fname, err.Error(), true}
		return fmt.Errorf("query error")

	default:
		lc <- LogMessage{fname, fmt.Sprintf("Found recent backupID (%s) that meets the search criteria", backupID), true}
	}

	/*Count how many backups will be deleted*/
	bfs := []BackupFiles{}
	lc <- LogMessage{fname, fmt.Sprintf("Performing Query: %s", GetBackupFileData(backupID)), true}
	rows, err := dbc.db.Query(GetBackupFileData(backupID))
	if err != nil {
		lc <- LogMessage{fname, "An error occurred querying the database", false}
		lc <- LogMessage{fname, err.Error(), true}
		return fmt.Errorf("failed to retrieve data on backup catalog entries to remove")
	}
	defer rows.Close()
	for rows.Next() {
		bf := BackupFiles{}
		err := rows.Scan(&bf.EntryType, &bf.FileCount, &bf.Bytes) //need unit test here!
		if err != nil {
			lc <- LogMessage{fname, "An error occurred querying the database", false}
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
		_, err = dbc.db.Exec(query)
		if err != nil {
			lc <- LogMessage{fname, "Query failed", false}
			lc <- LogMessage{fname, err.Error(), true}
			return fmt.Errorf("couldn't clean backup catalog")
		}

		lc <- LogMessage{fname, "Backup catalog successfully cleaned", false}
	}
	return nil
}

//This function deletes alerts from the table _SYS_STATISTICS.STATISTICS_ALERTS_BASE.  Alerts are deleted if they are older than
//the given number of days in the CleanDaysOlder argument.  No changes are made to the database if the dryrun argument is set to true
func (dbc *DbConfig) CleanAlertFunc(lc chan<- LogMessage, CleanDaysOlder uint, dryrun bool) error {
	fname := fmt.Sprintf("%s:%s", dbc.Name, "CleanAlert")
	lc <- LogMessage{fname, "Starting", true}
	if dryrun {
		lc <- LogMessage{fname, "Dry run enabled, no changes will be made", false}
	}
	/*Find how many alerts there are that match the deletion criteria*/
	var ac uint
	lc <- LogMessage{fname, fmt.Sprintf("Performing query:%s", GetAlertCount(CleanDaysOlder)), true}
	err := dbc.db.QueryRow(GetAlertCount(CleanDaysOlder)).Scan(&ac)
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
		_, err = dbc.db.Exec(GetAlertDelete(CleanDaysOlder))
		if err != nil {
			lc <- LogMessage{fname, "Query to remove alerts failed", false}
			lc <- LogMessage{fname, err.Error(), true}
			return err
		}
		lc <- LogMessage{fname, fmt.Sprintf("Successfully deleted %d alerts", ac), false}
	}
	return nil
}

//This function deletes free logsegments from the log volume.  Performing this task will reduce the disk space used in the log volume
//but may also cause a minor IO penalty when new new log segments need to be created.  It is more important to run this function is an MDC
//environemt than a non-MDC one.
func (dbc *DbConfig) CleanLogFunc(lc chan<- LogMessage, dryrun bool) error {
	fname := fmt.Sprintf("%s:%s", dbc.Name, "CleanLog")
	lc <- LogMessage{fname, "Starting", true}
	if dryrun {
		lc <- LogMessage{fname, "Dry run enabled, no changes will be made", false}
	}
	var count uint
	var bytes uint64
	lc <- LogMessage{fname, fmt.Sprintf("Performing Query:%s", QUERY_GetFeeLogSegments), true}
	err := dbc.db.QueryRow(QUERY_GetFeeLogSegments).Scan(&count, &bytes)
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
		lc <- LogMessage{fname, fmt.Sprintf("Performing Query:%s", QUERY_ReclaimLog), true}
		_, err = dbc.db.Exec(QUERY_ReclaimLog)
		if err != nil {
			lc <- LogMessage{fname, "Query produced a database error", false}
			lc <- LogMessage{fname, err.Error(), true}
			return fmt.Errorf("db error")
		}

		lc <- LogMessage{fname, "Log Volume Cleaning was successful.", false}
	}
	return nil
}

//Function that removes old audit events from the audit table.
func (dbc *DbConfig) CleanAuditFunc(lc chan<- LogMessage, CleanDaysOlder uint, dryrun bool) error {

	fname := fmt.Sprintf("%s:%s", dbc.Name, "CleanAuditLog")
	lc <- LogMessage{fname, "Starting", true}
	if dryrun {
		lc <- LogMessage{fname, "Dry run enabled, no changes will be made", false}
	}

	//Get the number of items to be removed
	lc <- LogMessage{fname, fmt.Sprintf("Performing query:%s", GetAuditCount(CleanDaysOlder)), true}
	var auditCount uint
	err := dbc.db.QueryRow(GetAuditCount(CleanDaysOlder)).Scan(&auditCount)
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
	Therefore we need to pass the time argument in a string.  We don't want to use the local time as the DB could be different
	So we'll run a query the DB for the time and feed it back in.*/
	lc <- LogMessage{fname, fmt.Sprintf("Performing Query:%s", GetDatetime(CleanDaysOlder)), true}
	var dateString string
	err = dbc.db.QueryRow(GetDatetime(CleanDaysOlder)).Scan(&dateString)
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
		lc <- LogMessage{fname, fmt.Sprintf("Performing Query:%s", GetTruncateAuditLog(dateParts[0])), true}
		_, err = dbc.db.Exec(GetTruncateAuditLog(dateParts[0]))
		if err != nil {
			lc <- LogMessage{fname, "Clean audit log query failed", false}
			lc <- LogMessage{fname, err.Error(), true}
			return fmt.Errorf("db error")
		}
	}

	return nil
}

func (dbc *DbConfig) CleanDataVolumeFunc(lc chan<- LogMessage, dryrun bool) error {
	fname := fmt.Sprintf("%s:%s", dbc.Name, "CleanDataVolume")
	lc <- LogMessage{fname, "Starting", true}
	if dryrun {
		lc <- LogMessage{fname, "Dry run enabled, no changes will be made", false}
	}

	/*Get the information about each datavolume*/
	rows, err := dbc.db.Query(QUERY_GetDataVolume)
	if err != nil {
		lc <- LogMessage{fname, "Query Failed", true}
		lc <- LogMessage{fname, err.Error(), true}
		return err
	}
	defer rows.Close()

	dvs := make([]DataVolume, 0)

	for rows.Next() {
		dv := DataVolume{}
		err := rows.Scan(&dv.Host, &dv.Port, &dv.UsedSizeBytes, &dv.TotalSizeBytes)
		if err != nil {
			lc <- LogMessage{fname, "Scan Error", true}
			lc <- LogMessage{fname, err.Error(), true}
			/*allow calling function to deal with the error*/
			return err
		}
		dvs = append(dvs, dv)
	}

	count := len(dvs)
	switch {
	case count == 0:
		lc <- LogMessage{fname, "No data volumes found", true}
		return nil /*Should this be an error?  It may be possible for some form of tenant to exist without data volumes*/
	case count == 1:
		lc <- LogMessage{fname, "1 data volume found", true}
	default:
		lc <- LogMessage{fname, fmt.Sprintf("%d data volumes found", count), true}
	}

	var failures = 0

	for k, v := range dvs {
		lc <- LogMessage{fname, fmt.Sprintf("Processing data volume %d of %d", k+1, len(dvs)), true}
		if !v.CleanNeeded() {
			lc <- LogMessage{fname, "Cleaning not required, data volume is less than 50% whitespace", true}
			continue
		} else {
			if dryrun {
				lc <- LogMessage{fname, "Cleaning required, but skipping due to dry run mode", true}
				continue
			} else {
				lc <- LogMessage{fname, "Cleaning required, data volume is more than 50% whitespace", true}
				_, err = dbc.db.Exec(GetCleanDataVolume(v.Host, v.Port))
				if err != nil {
					lc <- LogMessage{fname, "Failed to clean data volume", false}
					lc <- LogMessage{fname, err.Error(), true}
					failures += 1
				} else {
					lc <- LogMessage{fname, "Clean data volume OK", true}
				}
			}
		}
	}

	/*choose an exit*/
	switch {
	case failures == 0:
		lc <- LogMessage{fname, "Clean data volume finished with no errors", false}
		return nil
	case failures == 1:
		lc <- LogMessage{fname, "Clean data volume finished with one error", false}
		return fmt.Errorf("one data volume clean error recorded")
	default:
		lc <- LogMessage{fname, fmt.Sprintf("Clean data volume finished with %d errors", failures), false}
		return fmt.Errorf("%d data volume clean errors recorded", failures)
	}
}
