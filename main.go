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
	log.Printf("Drymode mode = %t\n", ac.DryRun)
	log.Printf("Getting Config")

	cnf, err := GetConfigFromFile(ac.ConfigFile)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Found a vailid config for %d databases\n", len(cnf.Databases))

}
