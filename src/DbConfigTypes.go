package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

//Struct for holding database configuration
type DbConfig struct {
	Name                    string // Friendly name of the DB.  <Tenant>@<SID> is a good option here
	Hostname                string // Hostname or IP address of the primary HANA node
	Port                    uint   // Port of the HANA DB
	Username                string // HANA DB user name to use
	password                string // Password for HANA DB user
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
	Results                 CleanResults //Results stored here and printed later
}

func (hdb DbConfig) Dsn() string {
	return fmt.Sprintf("hdb://%s:%s@%s:%d", hdb.Username, hdb.password, hdb.Hostname, hdb.Port)
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
	db.password = os.Getenv(fmt.Sprintf("HCC_%s", db.Name))
	if db.password == "" {
		return fmt.Errorf("was not able to source password from %s for the environnement", db.Name)
	}
	return nil
}

type CleanResults struct {
	TraceFilesRemoved       uint
	BackupFilesRemoved      uint
	BackupFilesBytesRemoved uint
	AlertsRemoved           uint
	LogSegmentsRemoved      uint
	LogSegmentsBytesRemoved uint
	AuditEntriesRemoved     uint
	DataVolumeBytesRemoved  uint
	TotalDiskBytesRemoved   uint
}

func (dbc *DbConfig) PrintResults() {

	/*Could I source this from the env?*/
	p := message.NewPrinter(language.English)

	fmt.Printf("%s:Cleaning Report\n", dbc.Name)

	/*Backup file report*/
	if dbc.CleanBackupCatalog {
		p.Printf("Backup files removed:\t\t%d\n", dbc.Results.BackupFilesRemoved)
		if dbc.DeleteOldBackups {
			p.Printf("Backup data removed:\t\t%.2fMiB\n", float64(dbc.Results.BackupFilesBytesRemoved/1024/1024))
		} else {
			p.Printf("Backup data removed:\t\tNot Enabled\n")
		}
	} else {
		p.Printf("Backup files removed:\t\tNot Enabled\n")
		p.Printf("Backup data removed:\t\tNot Enabled\n")
	}

	/*Alerts report*/
	if dbc.CleanAlerts {
		p.Printf("Alert entries removed:\t\t%d\n", dbc.Results.AlertsRemoved)
	} else {
		p.Printf("Alert entries removed:\t\tNot Enabled\n")
	}

	/*Audit report*/
	if dbc.CleanAudit {
		p.Printf("Audit entries removed:\t\t%d\n", dbc.Results.AuditEntriesRemoved)
	} else {
		p.Printf("Audit entries removed:\t\tNot Enabled\n")
	}

	/*Log Report*/
	if dbc.CleanLogVolume {
		p.Printf("Log Segments removed:\t\t%d\n", dbc.Results.LogSegmentsRemoved)
		p.Printf("Log Segments reduced by:\t%.2fMiB\n", float64(dbc.Results.LogSegmentsBytesRemoved)/1024/1024)
	} else {
		p.Printf("Log Segments removed:\t\tNot Enabled\n")
		p.Printf("Log Segments reduction:\t\tNot Enabled\n")
	}

	/*Data Report*/
	if dbc.CleanDataVolume {
		p.Printf("Data Volume reduced by:\t\t%.2fMiB\n", float64(dbc.Results.LogSegmentsBytesRemoved)/1024/1024)
	} else {
		p.Printf("Data Volume reduction:\t\tNot Enabled\n")
	}

}
