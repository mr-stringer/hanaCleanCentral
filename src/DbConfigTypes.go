package main

import (
	"database/sql"
	"fmt"
)

//Struct for holding database configuration
type DbConfig struct {
	Name                       string // Friendly name of the DB.  <Tenant>@<SID> is a good option here
	Hostname                   string // Hostname or IP address of the primary HANA node
	Port                       uint   // Port of the HANA DB
	Username                   string // HANA DB user name to use
	Password                   string // Password for HANA DB user
	RemoveTraces               bool   // If true, trace file management will be enabled - Defaults to false
	TraceRetentionDays         uint   // Specifies the number of days of trace files to retain
	TruncateBackupCatalog      bool   // If true, backup catalog truncation will be enabled - Defaults to false
	BackupCatalogRetentionDays uint   // Specifies the number of days of entries to retain
	DeleteOldBackups           bool   // If true, truncated files will be physically removed, if false entries are removed from the database only - Defaults to false
	ClearAlerts                bool   // If true, old alerts are removed from the embedded statistics server - Defaults to false
	AlertsOlderDeleteDays      uint   // Specifies the number of days of alerts to retain
}

func (hdb DbConfig) Dsn() string {
	return fmt.Sprintf("hdb://%s:%s@%s:%d", hdb.Username, hdb.Password, hdb.Hostname, hdb.Port)
}

//helper function that populates
func (hdb *DbConfig) NewDb() (*sql.DB, error) {
	db, err := sql.Open("hdb", hdb.Dsn())
	if err != nil {
		return db, err
	}

	err = db.Ping()
	if err != nil {
		return db, err
	}

	return db, nil
}
