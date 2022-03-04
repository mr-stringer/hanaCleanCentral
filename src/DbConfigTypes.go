package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
)

//Struct for holding database configuration
type DbConfig struct {
	Name                    string // Friendly name of the DB.  <Tenant>@<SID> is a good option here
	Hostname                string // Hostname or IP address of the primary HANA node
	Port                    uint   // Port of the HANA DB
	Username                string // HANA DB user name to use
	Password                string // Password for HANA DB user
	CleanTrace              bool   // If true, trace file management will be enabled - Defaults to false
	RetainTraceDays         uint   // Specifies the number of days of trace files to retain
	CleanBackupCatalog      bool   // If true, backup catalog truncation will be enabled - Defaults to false
	RetainBackupCatalogDays uint   // Specifies the number of days of entries to retain
	DeleteOldBackups        bool   // If true, truncated files will be physically removed, if false entries are removed from the database only - Defaults to false
	CleanAlerts             bool   // If true, old alerts are removed from the embedded statistics server - Defaults to false
	RetainAlertsDays        uint   // Specifies the number of days of alerts to retain
	CleanLogVolume          bool   // If true, free log segments will be removed from the file system
	CleanAudit              bool   // If true, old audit records will be deleted
	RetainAuditDays         uint   // Specifies the number of days of audit log to retain
	CleanDataVolume         bool   // If true, the data volume will be defragemented, currently uses default size of 120
	db                      *sql.DB
}

func (hdb DbConfig) Dsn() string {
	return fmt.Sprintf("hdb://%s:%s@%s:%d", hdb.Username, hdb.Password, hdb.Hostname, hdb.Port)
}

//helper function that connects to the target database and populates the DbConfig.db struct.
func (hdb *DbConfig) NewDb() error {
	var err error
	hdb.db, err = sql.Open("hdb", hdb.Dsn())
	if err != nil {
		return err
	}

	err = hdb.db.Ping()
	if err != nil {
		return err
	}

	return nil
}

//Gets the database password from the environment.
//The code search for the variable HCC_<dbConfig.Name>
func (db *DbConfig) GetPasswordFromEnv() error {

	log.Printf("Searching for password for %s from environment variable", db.Name)
	db.Password = os.Getenv(fmt.Sprintf("HCC_%s", db.Name))
	if db.Password == "" {
		return fmt.Errorf("was not able to source password from %s for the environnement", db.Name)
	}
	return nil
}
