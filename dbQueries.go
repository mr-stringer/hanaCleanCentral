package main

import "fmt"

/*In order to ensure that the exact same queries are used in the main code and in testing,
all of the DB queries are listed here and called from here*/

/*Queries that are static are available as constant strings whereas queries that are variable are returned from functions*/

const QUERY_GetVersion string = "SELECT VERSION FROM \"SYS\".\"M_DATABASE\""

func GetTraceFileQuery(days uint) string {
	return fmt.Sprintf("SELECT HOST, FILE_NAME, FILE_SIZE, FILE_MTIME FROM \"PUBLIC\".\"M_TRACEFILES\" WHERE FILE_MTIME > %d", days)
}
