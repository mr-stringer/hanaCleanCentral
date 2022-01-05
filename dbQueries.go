package main

import "fmt"

/*In order to ensure that the exact same queries are used in the main code and in testing,
all of the DB queries are listed here and called from here*/

/*Queries that are static are available as constant strings whereas queries that are variable are returned from functions*/

const QUERY_GetVersion string = "SELECT VERSION FROM \"SYS\".\"M_DATABASE\""

//The function returns a string which is used to query the HANA dataase. The function takes the argument days, this argument is used in the query to define the age of tracefiles
//that should be returned.  If the days is set to one, only tracefiles that have not been modified in the last 24 hours will be retutned.  In addition, they query will only return
//tracefiles that end "trc" or "gz".  This may change in the future but right now "gz" and "trc" are considered safe for housekeeping.
func GetTraceFileQuery(days uint) string {
	return fmt.Sprintf("SELECT HOST, FILE_NAME, FILE_SIZE, FILE_MTIME FROM \"SYS\".\"M_TRACEFILES\" WHERE FILE_MTIME < (SELECT ADD_DAYS(NOW(), -%d) FROM DUMMY) AND RIGHT(FILE_NAME, 3) = 'trc' OR FILE_MTIME < (SELECT ADD_DAYS(NOW(), -%d) FROM DUMMY) AND RIGHT(FILE_NAME, 2) = 'gz'", days, days)
}

//Test query
//SELECT * FROM "SYS"."M_TRACEFILES" WHERE FILE_MTIME < (SELECT ADD_DAYS(NOW(), -14) FROM DUMMY) AND RIGHT(FILE_NAME, 3) = 'trc' OR RIGHT(FILE_NAME, 2) = 'gz'

func GetRemoveTrace(hostname, filename string) string {
	return fmt.Sprintf("ALTER SYSTEM REMOVE TRACES('%s', '%s'", hostname, filename)
}
