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
//In some cases it may not be possible to remove a trace file, these incidents are logged but will not cause the function to error.
func TruncateTraceFiles(hdb *sql.DB, TrncDaysOlder uint) error {

	if TrncDaysOlder == 0 {
		log.Printf("TrncDaysOlder is set to zero, nothing to do")
		return nil
	}

	/*slice to hold results*/
	TraceFiles := make([]TraceFile, 0)

	/*Get the list of candidate tracefiles where the M time days is greater than the TrncDaysOlder arguments*/
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

	var count uint = 0
	var saved uint64 = 0
	/*Try and remove the files one by one to increase clarity in the logs*/
	for _, v := range TraceFiles {
		_, err := hdb.Exec(GetRemoveTrace(v.Hostname, v.TraceFile))
		if err != nil {
			log.Println(err.Error())
			log.Printf("The tracefile '%s' on host '%s' could not be removed, it may be open!  This will be retried next time.\n", v.TraceFile, v.Hostname)
			continue
		}

		count += 1
		saved += v.SizeBytes

	}

	if count > 0 {
		log.Printf("Removed %d old tracefiles saving %.2f MiB", count, float64(saved/1024/1024))
	}
	return nil
}
