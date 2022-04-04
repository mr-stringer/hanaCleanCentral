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
	for _, dbc := range cnf.Databases {

		/*Get password from environment if password not set*/
		if dbc.password == "" {
			err = dbc.GetPasswordFromEnv()
			if err != nil {
				lc <- LogMessage{dbc.Name, "No password for DB in environment or config file, skipping this DB", false}
				continue
			}
		}

		/*Initialise and test connection*/
		err := dbc.NewDb()
		if err != nil {
			lc <- LogMessage{dbc.Name, fmt.Sprintln("Could not connect to configured database"), false}
			lc <- LogMessage{dbc.Name, fmt.Sprintln("Check the configuration details and try again.  Full error message:"), false}
			lc <- LogMessage{dbc.Name, fmt.Sprintf("%s", err.Error()), false}
			lc <- LogMessage{dbc.Name, fmt.Sprintln("Cannot process any tasks for this databases"), false}
			continue
		}

		/*Get and print version - this may be used in later versions to test compatability*/
		v, err := dbc.HanaVersionFunc(lc)
		if err != nil {
			lc <- LogMessage{dbc.Name, fmt.Sprintln("Could not get HANA version of configured database"), false}
			lc <- LogMessage{dbc.Name, fmt.Sprintln("Full error message:"), false}
			lc <- LogMessage{dbc.Name, fmt.Sprintf("%s", err.Error()), false}
			lc <- LogMessage{dbc.Name, fmt.Sprintln("Will not process any tasks for this databases"), false}
			continue
		}
		lc <- LogMessage{dbc.Name, fmt.Sprintf("Hana Version found %s", v), false}

		err = dbc.CheckPrivileges(lc)
		if err != nil {
			lc <- LogMessage{dbc.Name, fmt.Sprint("There was a problem checking privileges for this database"), false}
			lc <- LogMessage{dbc.Name, fmt.Sprint("Full error message:"), false}
			lc <- LogMessage{dbc.Name, fmt.Sprintf("%s", err.Error()), false}
			continue
		}

		/*Clean trace files*/
		if dbc.CleanTrace {
			err = dbc.CleanTraceFilesFunc(lc, dbc.RetainTraceDays, ac.DryRun)
			if err != nil {
				lc <- LogMessage{dbc.Name, fmt.Sprintln("An error occurred trying to clean trace files"), false}
				lc <- LogMessage{dbc.Name, fmt.Sprintln("Full error message:"), false}
				lc <- LogMessage{dbc.Name, fmt.Sprintf("%s", err.Error()), false}
			}
		} else {
			lc <- LogMessage{dbc.Name, fmt.Sprintln("CleanTrace not enabled for this database"), false}
		}

		/*Clean backup catalog*/
		if dbc.CleanBackupCatalog {
			err = dbc.CleanBackupFunc(lc, dbc.RetainBackupCatalogDays, dbc.DeleteOldBackups, ac.DryRun)
			if err != nil {
				lc <- LogMessage{dbc.Name, fmt.Sprintln("An error occurred trying clean backup catalog"), false}
				lc <- LogMessage{dbc.Name, fmt.Sprintln("Full error message:"), false}
				lc <- LogMessage{dbc.Name, fmt.Sprintf("%s", err.Error()), false}
			}
		} else {
			lc <- LogMessage{dbc.Name, fmt.Sprintln("CleanBackupCatalog not enabled for this database"), false}
		}

		/*Clean Alerts*/
		if dbc.CleanAlerts {
			err = dbc.CleanAlertFunc(lc, dbc.RetainAlertsDays, ac.DryRun)
			if err != nil {
				lc <- LogMessage{dbc.Name, fmt.Sprintln("An error occurred trying clean alerts"), false}
				lc <- LogMessage{dbc.Name, fmt.Sprintln("Full error message:"), false}
				lc <- LogMessage{dbc.Name, fmt.Sprintf("%s", err.Error()), false}
			}
		} else {
			lc <- LogMessage{dbc.Name, fmt.Sprintln("CleanAlerts not enabled for this database"), false}
		}

		/*Clean Log Volume*/
		if dbc.CleanLogVolume {
			err = dbc.CleanLogFunc(lc, ac.DryRun)
			if err != nil {
				lc <- LogMessage{dbc.Name, fmt.Sprintln("An error occurred trying clean log volume"), false}
				lc <- LogMessage{dbc.Name, fmt.Sprintln("Full error message:"), false}
				lc <- LogMessage{dbc.Name, fmt.Sprintf("%s", err.Error()), false}
			}
		} else {
			lc <- LogMessage{dbc.Name, fmt.Sprintln("CleanLogVolume not enabled for this database"), false}
		}

		/*Clean Log Volume*/
		if dbc.CleanAudit {
			err = dbc.CleanAuditFunc(lc, dbc.RetainAuditDays, ac.DryRun)
			if err != nil {
				lc <- LogMessage{dbc.Name, fmt.Sprintln("An error occurred trying clean audit log"), false}
				lc <- LogMessage{dbc.Name, fmt.Sprintln("Full error message:"), false}
				lc <- LogMessage{dbc.Name, fmt.Sprintf("%s", err.Error()), false}
			}
		} else {
			lc <- LogMessage{dbc.Name, fmt.Sprintln("CleanAudit not enabled for this database"), false}
		}

		/*Clean Data Volume*/
		if dbc.CleanDataVolume {
			err = dbc.CleanDataVolumeFunc(lc, ac.DryRun)
			if err != nil {
				lc <- LogMessage{dbc.Name, fmt.Sprintln("An error occurred trying clean data volume log"), false}
				lc <- LogMessage{dbc.Name, fmt.Sprintln("Full error message:"), false}
				lc <- LogMessage{dbc.Name, fmt.Sprintf("%s", err.Error()), false}
			}
		} else {
			lc <- LogMessage{dbc.Name, fmt.Sprintln("CleanDataVolume not enabled for this database"), false}
		}

	}

	/*Print Results*/
	for _, dbc := range cnf.Databases {
		dbc.PrintResults()
	}

	quit <- true
}
