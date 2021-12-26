package main

import (
	"log"

	_ "github.com/SAP/go-hdb/driver"
)

func main() {
	log.Printf("HanaCleanCentral initalising")
	log.Printf("Getting Config")

	cnf, err := GetConfigFromFile("testfiles/configtest01.json")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Found a vailid config for %d databases\n", len(cnf.Databases))

}
