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
	log.Printf("Found a valid config for %d databases\n", len(cnf.Databases))

	/*Set up the logger*/
	lc := make(chan LogMessage)
	quit := make(chan bool)
	go Logger(ac, lc, quit)

	//basic ranging over DBs found
	for _, dbc := range cnf.Databases {

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
		log.Printf("%s:Finished tasks", dbc.Name)

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
	}
	/*flush and quit the logger*/
	quit <- true
}
