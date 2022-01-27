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
//All root config parameters must be set
type Config struct {
	CleanTrace              bool // If true, trace file management will be enabled
	RetainTraceDays         uint // Specifies the number of days of trace files to retain
	CleanBackupCatalog      bool // If true, backup catalog truncation will be enabled
	RetainBackupCatalogDays uint // Specifies the number of days of entries to retain
	DeleteOldBackups        bool // If true, truncated files will be physically removed, if false entries are removed from the database only
	CleanAlerts             bool // If true, old alerts are removed from the embedded statistics server
	RetainAlertsDays        uint // Specifies the number of days of alerts to retain
	CleanLogVolume          bool // If true, free log segments will be removed from the file system
	CleanAudit              bool // If true, old audit records will be deleted
	RetainAuditDays         uint // Specifes the number of days of audit log to retain
	Databases               []DbConfig
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
