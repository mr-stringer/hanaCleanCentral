package main

import (
	"fmt"
	"log"
)

//Application configuration parameters to be shared with functions
type AppConfig struct {
	ConfigFile string //the location of the config file
	Verbose    bool   //used for verbose logging
	DryRun     bool   //used for non-destructive testing
}

//Top level configuration for hanaCleanCentral
type Config struct {
	Databases []DbConfig
}

//Duplicate DB names are confusing at best and make it impossible to set
//password names from environment variables.  This function checks for duplicate names and
//returns an error if duplicate names are found
func (c *Config) CheckForDupeNames() error {
	keys := make(map[string]bool)

	//Use the map to check all db names are unique
	for _, v := range c.Databases {
		if _, value := keys[v.Name]; !value {
			keys[v.Name] = true
		} else {
			log.Printf("The database name %s occurs more than once is the configuration\n", v.Name)
			log.Printf("Each database name must be unique.\n")
			return fmt.Errorf("duplicate database name in configuration")
		}
	}
	return nil
}
