package main

import "fmt"

/*In order to ensure that the exact same queries are used in the main code and in testing,
all of the DB queries are listed here and called from here*/

/*Queries that are static are available as constant strings whereas queries that are variable are returned from functions*/

//Query to get HANA version
//Requires no additional privleges
const QUERY_GetVersion string = "SELECT VERSION FROM \"SYS\".\"M_DATABASE\""

//Query to get the number of free log segments and their total size in bytes
//Requires role MONITORING role
const QUERY_GetFeeLogSegments string = "SELECT COUNT(STATE) AS COUNT, COALESCE(SUM(TOTAL_SIZE),0) AS BYTES FROM SYS.M_LOG_SEGMENTS WHERE STATE = 'Free'"

//Query to delete freed log segments
//Requires LOG ADMIN privilege
const QUERY_ReclaimLog string = "ALTER SYSTEM RECLAIM LOG"

//Query to get the used and total space from each data volume.
//Requires MONITORING role
const QUERY_GetDataVolume string = "SELECT HOST, PORT, USED_SIZE, TOTAL_SIZE FROM SYS.M_VOLUME_FILES WHERE FILE_TYPE = 'DATA'"

//The function returns a string which is used to query the HANA database. The function takes the argument days, this argument is used in the query to define the age of tracefiles
//that should be returned.  If the days is set to one, only tracefiles that have not been modified in the last 24 hours will be returned.  In addition, they query will only return
//tracefiles that end "trc" or "gz".  This may change in the future but right now "gz" and "trc" are considered safe for housekeeping.
//Requires MONITORING role
func GetTraceFileQuery(days uint) string {
	return fmt.Sprintf("SELECT HOST, FILE_NAME, FILE_SIZE, FILE_MTIME FROM \"SYS\".\"M_TRACEFILES\" WHERE FILE_MTIME < (SELECT ADD_DAYS(NOW(), -%d) FROM DUMMY) AND RIGHT(FILE_NAME, 3) = 'trc' OR FILE_MTIME < (SELECT ADD_DAYS(NOW(), -%d) FROM DUMMY) AND RIGHT(FILE_NAME, 2) = 'gz'", days, days)
}

//Requires MONITORING role
func GetCheckTracePresent(filename string) string {
	/*trace file names should always be unique as they contain hostnames, rotation numbers etc*/
	return fmt.Sprintf("SELECT COUNT(FILE_NAME) AS TRACE FROM \"SYS\".\"M_TRACEFILES\" WHERE FILE_NAME = '%s'", filename)
}

//Returns a string query that is used to attempt to remove the identified tracefile
//Require TRACE ADMIN priv
func GetRemoveTrace(hostname, filename string) string {
	return fmt.Sprintf("ALTER SYSTEM REMOVE TRACES('%s', '%s')", hostname, filename)
}

//Returns a string query that is used to find the backup ID of the most recent full backup that is older than the days given in the argument
func GetLatestFullBackupID(days uint) string {
	return fmt.Sprintf("SELECT BACKUP_ID FROM \"SYS\".\"M_BACKUP_CATALOG\" WHERE STATE_NAME = 'successful' AND ENTRY_TYPE_NAME = 'complete data backup' AND SYS_END_TIME < (SELECT ADD_DAYS(NOW(),-%d) FROM DUMMY) ORDER BY SYS_END_TIME DESC LIMIT 1", days)
}

func GetBackupFileData(backupid string) string {
	return fmt.Sprintf("SELECT "+
		"B.ENTRY_TYPE_NAME AS ENTRY, "+
		"COUNT(B.BACKUP_ID) AS COUNT, "+
		"SUM(F.BACKUP_SIZE) AS BYTES "+
		"FROM \"SYS\".\"M_BACKUP_CATALOG\" AS B "+
		"LEFT JOIN \"SYS\".\"M_BACKUP_CATALOG_FILES\" AS F ON B.BACKUP_ID = F.BACKUP_ID "+
		"WHERE B.BACKUP_ID < %s "+
		"GROUP BY B.ENTRY_TYPE_NAME", backupid)
}

func GetBackupDelete(backupid string) string {
	return fmt.Sprintf("BACKUP CATALOG DELETE ALL BEFORE BACKUP_ID %s", backupid)
}

func GetBackupDeleteComplete(backupid string) string {
	return fmt.Sprintf("BACKUP CATALOG DELETE ALL BEFORE BACKUP_ID %s COMPLETE", backupid)
}

func GetAlertCount(days uint) string {
	return fmt.Sprintf("SELECT COUNT(SNAPSHOT_ID) AS COUNT FROM \"_SYS_STATISTICS\".\"STATISTICS_ALERTS_BASE\" WHERE ALERT_TIMESTAMP < ADD_DAYS(NOW(), -%d) LIMIT 1", days)
}

func GetAlertDelete(days uint) string {
	return fmt.Sprintf("DELETE FROM \"_SYS_STATISTICS\".\"STATISTICS_ALERTS_BASE\" WHERE ALERT_TIMESTAMP < ADD_DAYS(NOW(), -%d)", days)
}

func GetAuditCount(days uint) string {
	return fmt.Sprintf("SELECT COUNT(TIMESTAMP) AS COUNT FROM \"SYS\".\"AUDIT_LOG\" WHERE TIMESTAMP < (SELECT ADD_DAYS(NOW(), -%d) FROM DUMMY)", days)
}

//Function a string query the will return the historic datetime for now minus the number of days given in the argument
//Be aware, the returned string need to have subsecond element removed
func GetDatetime(days uint) string {
	return fmt.Sprintf("SELECT ADD_DAYS(NOW(), -%d) AS NOW FROM DUMMY", days)

}

//Function that returns a query that can is used to clear the audit log to the given datetime
func GetTruncateAuditLog(datetime string) string {
	return fmt.Sprintf("ALTER SYSTEM CLEAR AUDIT LOG UNTIL '%s'", datetime)
}

//Function that returns a query that is used to clean (defragment) HANA data volumes.  Must specify hostname and port
func GetCleanDataVolume(host string, port uint) string {
	return fmt.Sprintf("ALTER SYSTEM RECLAIM DATAVOLUME '%s:%d' 120 DEFRAGMENT", host, port)
}
