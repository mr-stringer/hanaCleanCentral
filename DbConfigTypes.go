package main

import (
	"database/sql"
	"fmt"
)

//Struct for holding database configuration
type DbConfig struct {
	Name                       string
	Hostname                   string
	Port                       uint
	Username                   string
	Password                   string
	RemoveTraces               bool
	TraceRetentionDays         uint
	TruncateBackupCatalog      bool
	BackupCatalogRetentionDays uint
	DeleteOldBackups           bool
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
