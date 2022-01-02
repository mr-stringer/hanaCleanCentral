package main

import (
	"database/sql"
	"log"
)

//HanaVersion function returns the version string of the database
func HanaVersion(hdb *sql.DB) (string, error) {
	var version string

	r1 := hdb.QueryRow(QUERY_GetVersion)
	err := r1.Scan(&version)
	if err != nil {
		log.Print("THROWING ERROR")
		return "", err // allow calling function to handle the error
	}

	return version, nil
}

//TruncateTraceFiles function removes closed trace files that are older than the number of days
//specified in the 'TrncDaysOlder' argument.  The function will log all activity.  The function will also return
//an error.  If no errors are found nil is returned.
func TruncateTraceFiles(hdb *sql.DB, TrncDaysOlder uint) error {

	//Get the list of tracefiles where the M time days is greater than

	return nil
}
