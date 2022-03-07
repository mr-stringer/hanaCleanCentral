package main

import (
	"fmt"
	"os"

	"github.com/Jeffail/gabs/v2"
)

//Function takes a channel to the logger and a path to the config file.  Returns a pointer to the
//configuration and an error.
//All root level fields must be set, if any are not set the function will return an error.
//DBs must have the following fields set:	'Name', 'Hostname', 'Port' and 'Username'.  If any of these parameters are not set, an error is returned.
//If the DB paramerter 'Password' the application will attempt to source it form the environemt.
//If an error occurs the function will return a pointer to an default configuration (all false and 0) and the error
func GetConfigFromFile(lc chan<- LogMessage, path string) (*Config, error) {

	/* The original method - oh if it were that simple :/
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&cnf)
	if err != nil {
		log.Printf("Failed to decode!")
		return cnf, fmt.Errorf("failed to decode")
	}
	*/

	var cnf Config
	var mt Config /*empty struct to return if error*/
	ba1, err := os.ReadFile(path)
	if err != nil {
		lc <- LogMessage{"HccConfig", "Cannot read the given configuration file", false}
		return &mt, err
	}

	jp, err := gabs.ParseJSON(ba1)
	if err != nil {
		lc <- LogMessage{"HccConfig", "Cannot parse configuration file", false}
		lc <- LogMessage{"HccConfig", err.Error(), true}
		return &mt, err
	}

	var ok bool
	var tf float64
	/*Read the root configuration*/
	cnf.CleanTrace, ok = jp.Path("CleanTrace").Data().(bool)
	if !ok {
		lc <- LogMessage{"HccConfig", "Could not parse 'CleanTrace', all root parameters must be set.  Cannot continue", false}
		return &mt, fmt.Errorf("config error")
	}

	tf, ok = jp.Path("RetainTraceDays").Data().(float64)
	if !ok {
		lc <- LogMessage{"HccConfig", "Could not parse 'RetainTraceDays', all root parameters must be set.  Cannot continue", false}
		return &mt, fmt.Errorf("config error")
	}
	/*Check that number is 0 or greater*/
	if tf < 0 {
		lc <- LogMessage{"HccConfig", "Parameter 'RetainTraceDays' must be 0 or higher.  Cannot continue", false}
		return &mt, fmt.Errorf("config error")
	}
	cnf.RetainTraceDays = uint(tf)

	cnf.CleanBackupCatalog, ok = jp.Path("CleanBackupCatalog").Data().(bool)
	if !ok {
		lc <- LogMessage{"HccConfig", "Could not parse 'CleanBackupCatalog', all root parameters must be set.  Cannot continue", false}
		return &mt, fmt.Errorf("config error")
	}

	tf, ok = jp.Path("RetainBackupCatalogDays").Data().(float64)
	if !ok {
		lc <- LogMessage{"HccConfig", "Could not parse 'RetainBackupCatalogDays', all root parameters must be set.  Cannot continue", false}
		return &mt, fmt.Errorf("config error")
	}
	if tf < 0 {
		lc <- LogMessage{"HccConfig", "Parameter 'RetainBackupCatalogDays' must be 0 or higher.  Cannot continue", false}
		return &mt, fmt.Errorf("config error")
	}
	cnf.RetainBackupCatalogDays = uint(tf)

	cnf.DeleteOldBackups, ok = jp.Path("DeleteOldBackups").Data().(bool)
	if !ok {
		lc <- LogMessage{"HccConfig", "Could not parse 'DeleteOldBackups', all root parameters must be set.  Cannot continue", false}
		return &mt, fmt.Errorf("config error")
	}

	cnf.CleanAlerts, ok = jp.Path("CleanAlerts").Data().(bool)
	if !ok {
		lc <- LogMessage{"HccConfig", "Could not parse 'CleanAlerts', all root parameters must be set.  Cannot continue", false}
		return &mt, fmt.Errorf("config error")
	}

	tf, ok = jp.Path("RetainAlertsDays").Data().(float64)
	if !ok {
		lc <- LogMessage{"HccConfig", "Could not parse 'RetainAlertsDays', all root parameters must be set.  Cannot continue", false}
		return &mt, fmt.Errorf("config error")
	}
	if tf < 0 {
		lc <- LogMessage{"HccConfig", "Parameter 'RetainAlertsDays' must be 0 or higher.  Cannot continue", false}
		return &mt, fmt.Errorf("config error")
	}
	cnf.RetainAlertsDays = uint(tf)

	cnf.CleanLogVolume, ok = jp.Path("CleanLogVolume").Data().(bool)
	if !ok {
		lc <- LogMessage{"HccConfig", "Could not parse 'CleanLogVolume', all root parameters must be set.  Cannot continue", false}
		return &mt, fmt.Errorf("config error")
	}

	cnf.CleanAudit, ok = jp.Path("CleanAudit").Data().(bool)
	if !ok {
		lc <- LogMessage{"HccConfig", "Could not parse 'CleanAudit', all root parameters must be set.  Cannot continue", false}
		return &mt, fmt.Errorf("config error")
	}

	tf, ok = jp.Path("RetainAuditDays").Data().(float64)
	if !ok {
		lc <- LogMessage{"HccConfig", "Could not parse 'RetainAuditDays', all root parameters must be set.  Cannot continue", false}
		return &mt, fmt.Errorf("config error")
	}
	if tf < 0 {
		lc <- LogMessage{"HccConfig", "Parameter 'RetainAuditDays' must be 0 or higher.  Cannot continue", false}
		return &mt, fmt.Errorf("config error")
	}
	cnf.RetainAuditDays = uint(tf)

	cnf.CleanDataVolume, ok = jp.Path("CleanDataVolume").Data().(bool)
	if !ok {
		lc <- LogMessage{"HccConfig", "Could not parse 'CleanDataVolume', all root parameters must be set.  Cannot continue", false}
		return &mt, fmt.Errorf("config error")
	}

	/*Now iterate over DBs*/
	for k, child := range jp.S("Databases").Children() {
		//Create an struct instance
		db := DbConfig{}
		var tf float64

		db.Name, ok = child.Path("Name").Data().(string)
		if !ok {
			lc <- LogMessage{"Hcc_Config", fmt.Sprintf("Cannot parse 'Name' for DB config %d", k), false}
			lc <- LogMessage{"HccConfig", "'Name' must be set for all DB configs.  Cannot continue", false}
			return &mt, fmt.Errorf("config error")
		}

		db.Hostname, ok = child.Path("Hostname").Data().(string)
		if !ok {
			lc <- LogMessage{"HccConfig", fmt.Sprintf("Cannot parse 'Hostname' for DB config %d", k), false}
			lc <- LogMessage{"HccConfig", "'Hostname' must be set for all DB configs.  Cannot continue", false}
			return &mt, fmt.Errorf("config error")
		}

		tf, ok = child.Path("Port").Data().(float64)
		if !ok {
			lc <- LogMessage{"HccConfig", fmt.Sprintf("Cannot parse 'Port' for DB config %d", k), false}
			lc <- LogMessage{"HccConfig", "'Port' must be set for all DB configs.  Cannot continue", false}
			return &mt, fmt.Errorf("config error")
		}
		if tf < 0 {
			lc <- LogMessage{"HccConfig", "Parameter 'Port' must be 0 or higher.  Cannot continue", false}
			return &mt, fmt.Errorf("config error")
		}
		db.Port = uint(tf)

		db.Username, ok = child.Path("Username").Data().(string)
		if !ok {
			lc <- LogMessage{"HccConfig", fmt.Sprintf("Cannot parse 'Username' for DB config %d", k), false}
			lc <- LogMessage{"HccConfig", "'Username' must be set for all DB configs.  Cannot continue", false}
			return &mt, fmt.Errorf("config error")
		}

		db.password, ok = child.Path("Password").Data().(string)
		if !ok {
			lc <- LogMessage{"HccConfig", fmt.Sprintf("Cannot parse 'Password' for DB config %d\n", k), false}
			lc <- LogMessage{"HccConfig", fmt.Sprintf("The password will be sourced from the environmental variable 'HCC_%s'", db.Name), false}
		}

		db.CleanTrace, ok = child.Path("CleanTrace").Data().(bool)
		if !ok {
			lc <- LogMessage{"HccConfig", fmt.Sprintf("Cannot parse 'CleanTrace' for DB config %d.  Will inherit from %v from root config", k, cnf.CleanTrace), true}
			db.CleanTrace = cnf.CleanTrace
		}

		tf, ok = child.Path("RetainTraceDays").Data().(float64)
		if !ok {
			lc <- LogMessage{"HccConfig", fmt.Sprintf("Cannot parse 'RetainTraceDays' for DB config %d.  Will inherit from %d from root config", k, cnf.RetainTraceDays), true}
			db.RetainTraceDays = cnf.RetainTraceDays
		} else if tf < 0 {
			lc <- LogMessage{"HccConfig", fmt.Sprintf("Parameter 'RetainTraceDays' for DB %d must be 0 or higher.  Cannot continue", k), false}
			return &mt, fmt.Errorf("config error")
		} else {
			db.RetainTraceDays = uint(tf)
		}

		db.CleanBackupCatalog, ok = child.Path("CleanBackupCatalog").Data().(bool)
		if !ok {
			lc <- LogMessage{"HccConfig", fmt.Sprintf("Cannot parse 'CleanBackupCatalog' for DB config %d.  Will inherit from %v from root config", k, cnf.CleanBackupCatalog), true}
			db.CleanBackupCatalog = cnf.CleanBackupCatalog
		}

		tf, ok = child.Path("RetainBackupCatalogDays").Data().(float64)
		if !ok {
			lc <- LogMessage{"HccConfig", fmt.Sprintf("Cannot parse 'RetainBackupCatalogDays' for DB config %d.  Will inherit from %d from root config", k, cnf.RetainBackupCatalogDays), true}
			db.RetainBackupCatalogDays = cnf.RetainBackupCatalogDays
		} else if tf < 0 {
			lc <- LogMessage{"HccConfig", fmt.Sprintf("Parameter 'RetainBackupCatalogDays' for DB %d must be 0 or higher.  Cannot continue", k), false}
			return &mt, fmt.Errorf("config error")
		} else {
			db.RetainBackupCatalogDays = uint(tf)
		}

		db.DeleteOldBackups, ok = child.Path("DeleteOldBackups").Data().(bool)
		if !ok {
			lc <- LogMessage{"HccConfig", fmt.Sprintf("Cannot parse 'DeleteOldBackups' for DB config %d.  Will inherit from %v from root config", k, cnf.DeleteOldBackups), true}
			db.DeleteOldBackups = cnf.DeleteOldBackups
		}

		db.CleanAlerts, ok = child.Path("CleanAlerts").Data().(bool)
		if !ok {
			lc <- LogMessage{"HccConfig", fmt.Sprintf("Cannot parse 'CleanAlerts' for DB config %d.  Will inherit from %v from root config", k, cnf.CleanAlerts), true}
			db.CleanAlerts = cnf.CleanAlerts
		}

		tf, ok = child.Path("RetainAlertsDays").Data().(float64)
		if !ok {
			lc <- LogMessage{"HccConfig", fmt.Sprintf("Cannot parse 'RetainAlertsDays' for DB config %d.  Will inherit from %d from root config", k, cnf.RetainAlertsDays), true}
			db.RetainAlertsDays = cnf.RetainAlertsDays
		} else if tf < 0 {
			lc <- LogMessage{"HccConfig", fmt.Sprintf("Parameter 'RetainAlertsDays' for DB %d must be 0 or higher.  Cannot continue", k), false}
			return &mt, fmt.Errorf("config error")
		} else {
			db.RetainAlertsDays = uint(tf)
		}

		db.CleanLogVolume, ok = child.Path("CleanLogVolume").Data().(bool)
		if !ok {
			lc <- LogMessage{"HccConfig", fmt.Sprintf("Cannot parse 'CleanLogVolume' for DB config %d.  Will inherit from %v from root config", k, cnf.CleanLogVolume), true}
			db.CleanLogVolume = cnf.CleanLogVolume
		}

		db.CleanAudit, ok = child.Path("CleanAudit").Data().(bool)
		if !ok {
			lc <- LogMessage{"HccConfig", fmt.Sprintf("Cannot parse 'CleanAudit' for DB config %d.  Will inherit from %v from root config", k, cnf.CleanAudit), true}
			db.CleanAudit = cnf.CleanAudit
		}

		tf, ok = child.Path("RetainAuditDays").Data().(float64)
		if !ok {
			lc <- LogMessage{"HccConfig", fmt.Sprintf("Cannot parse 'RetainAuditDays' for DB config %d.  Will inherit from %d from root config", k, cnf.RetainAuditDays), true}
			db.RetainAuditDays = cnf.RetainAuditDays
		} else if tf < 0 {
			lc <- LogMessage{"HccConfig", fmt.Sprintf("Parameter 'RetainAuditDays' for DB %d must be 0 or higher.  Cannot continue", k), false}
			return &mt, fmt.Errorf("config error")
		} else {
			db.RetainAuditDays = uint(tf)
		}

		db.CleanDataVolume, ok = child.Path("CleanDataVolume").Data().(bool)
		if !ok {
			lc <- LogMessage{"HccConfig", fmt.Sprintf("Cannot parse 'CleanDataVolume' for DB config %d.  Will inherit from %v from root config", k, cnf.CleanAudit), true}
			db.CleanDataVolume = cnf.CleanDataVolume
		}

		//append to slice
		cnf.Databases = append(cnf.Databases, db)
	}

	return &cnf, nil
}
