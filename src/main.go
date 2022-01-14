package main

import (
	"log"

	_ "github.com/SAP/go-hdb/driver"
)

func main() {

	ac := ProcessFlags()

	log.Printf("HanaCleanCentral initalising\n")
	log.Printf("Configuration file = %s\n", ac.ConfigFile)
	log.Printf("Verbose mode = %t\n", ac.Verbose)
	log.Printf("Dryrun mode = %t\n", ac.DryRun)
	log.Printf("Getting Config")

	cnf, err := GetConfigFromFile(ac.ConfigFile)
	if err != nil {
		log.Fatal(err)
	}

	/*check config for duplicates*/
	err = cnf.CheckForDupeNames()
	if err != nil {
		log.Printf("%s\n", err.Error())
		return
	}

	log.Printf("Found a valid config for %d databases\n", len(cnf.Databases))

	/*Set up the logger*/
	lc := make(chan LogMessage)
	quit := make(chan bool)

	defer close(lc)
	defer close(quit)

	go Logger(ac, lc, quit)

	//basic ranging over DBs found
	for _, dbc := range cnf.Databases {

		/*Get password from environment if password not set*/
		err = dbc.GetPasswordFromEnv()
		if err != nil {
			lc <- LogMessage{dbc.Name, "No password for DB in environment or config file, skipping this DB", false}
			continue
		}

		db, err := dbc.NewDb()
		if err != nil {
			log.Printf("%s:Could not connect to configured database\n", dbc.Name)
			log.Printf("%s:Check the configuration details and try again.  Full error message:", dbc.Name)
			log.Print(err.Error())
			log.Printf("%s:Cannot process any tasks for this databases\n", dbc.Name)
			continue
		}

		v, err := HanaVersion(db)
		if err != nil {
			log.Printf("%s:Could not get HANA version of configured database\n", dbc.Name)
			log.Printf("%s:Full error message:", dbc.Name)
			log.Print(err.Error())
			log.Printf("%s:Will not process any tasks for this databases\n", dbc.Name)
			continue
		}
		log.Printf("%s:Hana Version found %s\n", dbc.Name, v)

		err = TruncateTraceFiles(lc, dbc.Name, db, dbc.TraceRetentionDays, ac.DryRun)
		if err != nil {
			log.Printf("%s:Error occured whilst trying to remove old tracesfiles", dbc.Name)
			log.Printf("%s:Full error message:", dbc.Name)
			log.Print(err.Error())
			log.Printf("%s:Will not process any tasks for this databases\n", dbc.Name)
			continue
		}

		err = TruncateBackupCatalog(lc, dbc.Name, db, dbc.BackupCatalogRetentionDays, dbc.DeleteOldBackups, ac.DryRun)
		if err != nil {
			log.Printf("%s:Backup catalog truncation failed", dbc.Name)
		}

		if dbc.ClearAlerts {
			err = ClearAlert(lc, dbc.Name, db, dbc.AlertsOlderDeleteDays, ac.DryRun)
			if err != nil {
				lc <- LogMessage{dbc.Name, "Failed to clear old alerts", false}
			}
		} else {
			lc <- LogMessage{dbc.Name, "Skipping alert clearing", false}
		}
	}

	/*flush and quit the logger*/
	quit <- true
	close(quit)
	close(lc)

}
