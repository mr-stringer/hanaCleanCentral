package main

import (
	"reflect"
	"testing"
)

func TestGetConfigFromFile(t *testing.T) {

	/*set up logger*/
	lc := make(chan LogMessage)
	quit := make(chan bool)
	defer close(lc)
	defer close(quit)
	go Logger(AppConfig{"file", false, false, false}, lc, quit)

	type args struct {
		lc   chan<- LogMessage
		path string
	}
	tests := []struct {
		name    string
		args    args
		want    *Config
		wantErr bool
	}{
		{"GoodFile01", args{lc, "testFiles/configtest01.json"}, &Config{true, 60, true, 60, true, true, 60, true, true, 60, true, []DbConfig{{"systemdb_TST", "hanadb.mydomain.int", 30015, "sstringer", "ReallyCoolPassw0rd", true, 60, true, 60, true, true, 60, true, true, 60, true, nil, CleanResults{}}}}, false},
		{"GoodFile02", args{lc, "testFiles/configtest02.json"}, &Config{true, 60, true, 60, true, true, 60, true, true, 60, true, []DbConfig{{"systemdb_TST", "hanadb.mydomain.int", 30015, "sstringer", "ReallyCoolPassw0rd", true, 60, true, 60, true, true, 60, true, true, 60, true, nil, CleanResults{}}, {"Ten01_TST", "hanadb.mydomain.int", 30041, "sstringer", "ReallyCoolPassw0rd", true, 60, true, 60, true, true, 60, true, false, 0, true, nil, CleanResults{}}}}, false},
		{"NoRootCleanTrace", args{lc, "testFiles/NoRootCleanTrace.json"}, &Config{}, true},
		{"NoRootRetainTraceDays", args{lc, "testFiles/NoRootRetainTraceDays.json"}, &Config{}, true},
		{"NoRootCleanBackupCatalog", args{lc, "testFiles/NoRootCleanBackupCatalog.json"}, &Config{}, true},
		{"NoRootRetainBackupCatalogDays", args{lc, "testFiles/NoRootRetainBackupCatalogDays.json"}, &Config{}, true},
		{"NoRootDeleteOldBackups", args{lc, "testFiles/NoRootDeleteOldBackups.json"}, &Config{}, true},
		{"NoRootCleanAlerts", args{lc, "testFiles/NoRootCleanAlerts.json"}, &Config{}, true},
		{"NoRootRetainAlertsDays", args{lc, "testFiles/NoRootRetainAlertsDays.json"}, &Config{}, true},
		{"NoRootCleanLogVolume", args{lc, "testFiles/NoRootCleanLogVolume.json"}, &Config{}, true},
		{"NoRootCleanAudit", args{lc, "testFiles/NoRootCleanAudit.json"}, &Config{}, true},
		{"NoRootRetainAuditDays", args{lc, "testFiles/NoRootRetainAuditDays.json"}, &Config{}, true},
		{"NegativeRootRetainTraceDays", args{lc, "testFiles/NegativeRootRetainTraceDays.json"}, &Config{}, true},
		{"NegativeRootRetainBackupCatalogDays", args{lc, "testFiles/NegativeRootRetainBackupCatalogDays.json"}, &Config{}, true},
		{"NegativeRootRetainAlertsDays", args{lc, "testFiles/NegativeRootRetainAlertsDays.json"}, &Config{}, true},
		{"NegativeRootRetainAuditDays", args{lc, "testFiles/NegativeRootRetainAuditDays.json"}, &Config{}, true},
		{"NoDbName", args{lc, "testFiles/NoDbName.json"}, &Config{}, true},
		{"NoDbHostname", args{lc, "testFiles/NoDbHostname.json"}, &Config{}, true},
		{"NoDbPort", args{lc, "testFiles/NoDbPort.json"}, &Config{}, true},
		{"NoDbUsername", args{lc, "testFiles/NoDbUsername.json"}, &Config{}, true},
		{"NoDbPassword", args{lc, "testFiles/NoDbPassword.json"}, &Config{true, 60, true, 60, true, true, 60, true, true, 60, true, []DbConfig{{"systemdb_TST", "hanadb.mydomain.int", 30015, "sstringer", "", true, 60, true, 60, true, true, 60, true, true, 60, true, nil, CleanResults{}}}}, false},
		{"NegativeDbPort", args{lc, "testFiles/NegativeDbPort.json"}, &Config{}, true},
		{"NegativeDbRetainTraceDays", args{lc, "testFiles/NegativeDbRetainTraceDays.json"}, &Config{}, true},
		{"NegativeDbRetainAlertsDays", args{lc, "testFiles/NegativeDbRetainAlertsDays.json"}, &Config{}, true},
		{"NegativeDbRetainBackupCatalogDays", args{lc, "testFiles/NegativeDbRetainBackupCatalogDays.json"}, &Config{}, true},
		{"NegativeDbRetainAuditDays", args{lc, "testFiles/NegativeDbRetainAuditDays.json"}, &Config{}, true},
		{"NoDbUsername", args{lc, "testFiles/NoDbUsername.json"}, &Config{}, true},
		{"DbOveride", args{lc, "testFiles/DbOverride.json"}, &Config{false, 0, false, 0, false, false, 0, false, false, 0, true, []DbConfig{{"systemdb_TST", "hanadb.mydomain.int", 30015, "sstringer", "ReallyCoolPassw0rd", true, 30, true, 30, true, true, 30, true, true, 30, true, nil, CleanResults{}}}}, false},
		{"InvalidJson", args{lc, "testFiles/invalidJson.json"}, &Config{}, true},
		{"InvalidPath", args{lc, "testFiles/NOFILE.json"}, &Config{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetConfigFromFile(tt.args.lc, tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetConfigFromFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetConfigFromFile() = %v, want %v", got, tt.want)
			}
		})
	}
}
