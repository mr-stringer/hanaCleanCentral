package main

import (
	"fmt"
	"log"

	_ "github.com/SAP/go-hdb/driver"
)

func main() {

	ac := ProcessFlags()

	/*Set up the logger*/
	lc := make(chan LogMessage)
	quit := make(chan bool)
	defer close(lc)
	defer close(quit)
	go Logger(ac, lc, quit)

	lc <- LogMessage{"HCC", "HanaCleanCentral initalising", false}
	lc <- LogMessage{"HCC", fmt.Sprintf("Configuration file = %s", ac.ConfigFile), false}
	lc <- LogMessage{"HCC", fmt.Sprintf("Verbose mode = %t", ac.Verbose), false}
	lc <- LogMessage{"HCC", fmt.Sprintf("Dryrun mode = %t", ac.DryRun), false}
	lc <- LogMessage{"HCC", "Getting Config", false}

	cnf, err := GetConfigFromFile(lc, ac.ConfigFile)
	if err != nil {
		lc <- LogMessage{"HCC", fmt.Sprintf("%s\n", err.Error()), false}
		quit <- true
		return
	}

	/*check config for duplicates*/
	err = cnf.CheckForDupeNames()
	if err != nil {
		lc <- LogMessage{"HCC", fmt.Sprintf("%s\n", err.Error()), false}
		quit <- true
		return
	}

	if ac.PrintConfig {
		fmt.Printf("Printing application configuration\n")
		err := cnf.PrintConfig()
		if err != nil {
			fmt.Printf("Printing application configuration failed.")
			fmt.Println(err.Error())
			return
		}
		return
	}

	log.Printf("Found a valid config for %d databases\n", len(cnf.Databases))

	//basic ranging over DBs found - requires changes to support parallel execution
	for index := range cnf.Databases {

		/*Get password from environment if password not set*/
		if cnf.Databases[index].password == "" {
			err = cnf.Databases[index].GetPasswordFromEnv()
			if err != nil {
				lc <- LogMessage{cnf.Databases[index].Name, "No password for DB in environment or config file, skipping this DB", false}
				continue
			}
		}

		/*Initialise and test connection*/
		err := cnf.Databases[index].NewDb()
		if err != nil {
			lc <- LogMessage{cnf.Databases[index].Name, fmt.Sprintln("Could not connect to configured database"), false}
			lc <- LogMessage{cnf.Databases[index].Name, fmt.Sprintln("Check the configuration details and try again.  Full error message:"), false}
			lc <- LogMessage{cnf.Databases[index].Name, fmt.Sprintf("%s", err.Error()), false}
			lc <- LogMessage{cnf.Databases[index].Name, fmt.Sprintln("Cannot process any tasks for this databases"), false}
			continue
		}

		/*Get and print version - this may be used in later versions to test compatability*/
		v, err := cnf.Databases[index].HanaVersionFunc(lc)
		if err != nil {
			lc <- LogMessage{cnf.Databases[index].Name, fmt.Sprintln("Could not get HANA version of configured database"), false}
			lc <- LogMessage{cnf.Databases[index].Name, fmt.Sprintln("Full error message:"), false}
			lc <- LogMessage{cnf.Databases[index].Name, fmt.Sprintf("%s", err.Error()), false}
			lc <- LogMessage{cnf.Databases[index].Name, fmt.Sprintln("Will not process any tasks for this databases"), false}
			continue
		}
		lc <- LogMessage{cnf.Databases[index].Name, fmt.Sprintf("Hana Version found %s", v), false}

		err = cnf.Databases[index].CheckPrivileges(lc)
		if err != nil {
			lc <- LogMessage{cnf.Databases[index].Name, fmt.Sprint("There was a problem checking privileges for this database"), false}
			lc <- LogMessage{cnf.Databases[index].Name, fmt.Sprint("Full error message:"), false}
			lc <- LogMessage{cnf.Databases[index].Name, fmt.Sprintf("%s", err.Error()), false}
			continue
		}

		/*Clean trace files*/
		if cnf.Databases[index].CleanTrace {
			err = cnf.Databases[index].CleanTraceFilesFunc(lc, cnf.Databases[index].RetainTraceDays, ac.DryRun)
			if err != nil {
				lc <- LogMessage{cnf.Databases[index].Name, fmt.Sprintln("An error occurred trying to clean trace files"), false}
				lc <- LogMessage{cnf.Databases[index].Name, fmt.Sprintln("Full error message:"), false}
				lc <- LogMessage{cnf.Databases[index].Name, fmt.Sprintf("%s", err.Error()), false}
			}
		} else {
			lc <- LogMessage{cnf.Databases[index].Name, fmt.Sprintln("CleanTrace not enabled for this database"), false}
		}

		/*Clean backup catalog*/
		if cnf.Databases[index].CleanBackupCatalog {
			err = cnf.Databases[index].CleanBackupFunc(lc, cnf.Databases[index].RetainBackupCatalogDays, cnf.Databases[index].DeleteOldBackups, ac.DryRun)
			if err != nil {
				lc <- LogMessage{cnf.Databases[index].Name, fmt.Sprintln("An error occurred trying clean backup catalog"), false}
				lc <- LogMessage{cnf.Databases[index].Name, fmt.Sprintln("Full error message:"), false}
				lc <- LogMessage{cnf.Databases[index].Name, fmt.Sprintf("%s", err.Error()), false}
			}
		} else {
			lc <- LogMessage{cnf.Databases[index].Name, fmt.Sprintln("CleanBackupCatalog not enabled for this database"), false}
		}

		/*Clean Alerts*/
		if cnf.Databases[index].CleanAlerts {
			err = cnf.Databases[index].CleanAlertFunc(lc, cnf.Databases[index].RetainAlertsDays, ac.DryRun)
			if err != nil {
				lc <- LogMessage{cnf.Databases[index].Name, fmt.Sprintln("An error occurred trying clean alerts"), false}
				lc <- LogMessage{cnf.Databases[index].Name, fmt.Sprintln("Full error message:"), false}
				lc <- LogMessage{cnf.Databases[index].Name, fmt.Sprintf("%s", err.Error()), false}
			}
		} else {
			lc <- LogMessage{cnf.Databases[index].Name, fmt.Sprintln("CleanAlerts not enabled for this database"), false}
		}

		/*Clean Log Volume*/
		if cnf.Databases[index].CleanLogVolume {
			err = cnf.Databases[index].CleanLogFunc(lc, ac.DryRun)
			if err != nil {
				lc <- LogMessage{cnf.Databases[index].Name, fmt.Sprintln("An error occurred trying clean log volume"), false}
				lc <- LogMessage{cnf.Databases[index].Name, fmt.Sprintln("Full error message:"), false}
				lc <- LogMessage{cnf.Databases[index].Name, fmt.Sprintf("%s", err.Error()), false}
			}
		} else {
			lc <- LogMessage{cnf.Databases[index].Name, fmt.Sprintln("CleanLogVolume not enabled for this database"), false}
		}

		/*Clean Log Volume*/
		if cnf.Databases[index].CleanAudit {
			err = cnf.Databases[index].CleanAuditFunc(lc, cnf.Databases[index].RetainAuditDays, ac.DryRun)
			if err != nil {
				lc <- LogMessage{cnf.Databases[index].Name, fmt.Sprintln("An error occurred trying clean audit log"), false}
				lc <- LogMessage{cnf.Databases[index].Name, fmt.Sprintln("Full error message:"), false}
				lc <- LogMessage{cnf.Databases[index].Name, fmt.Sprintf("%s", err.Error()), false}
			}
		} else {
			lc <- LogMessage{cnf.Databases[index].Name, fmt.Sprintln("CleanAudit not enabled for this database"), false}
		}

		/*Clean Data Volume*/
		if cnf.Databases[index].CleanDataVolume {
			err = cnf.Databases[index].CleanDataVolumeFunc(lc, ac.DryRun)
			if err != nil {
				lc <- LogMessage{cnf.Databases[index].Name, fmt.Sprintln("An error occurred trying clean data volume log"), false}
				lc <- LogMessage{cnf.Databases[index].Name, fmt.Sprintln("Full error message:"), false}
				lc <- LogMessage{cnf.Databases[index].Name, fmt.Sprintf("%s", err.Error()), false}
			}
		} else {
			lc <- LogMessage{cnf.Databases[index].Name, fmt.Sprintln("CleanDataVolume not enabled for this database"), false}
		}

	}

	/*Print Results*/
	if !ac.DryRun {
		for _, dbc := range cnf.Databases { /*Can use range here*/
			dbc.PrintResults()
		}
	} else {
		fmt.Printf("Dry run enabled: No report data\n")
	}

	quit <- true
}
