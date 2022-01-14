package main

import "testing"

func TestConfig_CheckForDupeNames(t *testing.T) {
	tests := []struct {
		name    string
		c       *Config
		wantErr bool
	}{
		{"Good_SingleDB", &Config{[]DbConfig{{"systemdb@TST", "hanadb.mydomain.int", 30015, "sstringer", "ReallzyKoolPassw0rd", true, 14, false, 0, false, true, 30}}}, false},
		{"Good_TwoDBs", &Config{[]DbConfig{{"systemdb@TST", "hanadb.mydomain.int", 30015, "sstringer", "ReallzyKoolPassw0rd", true, 14, false, 0, false, true, 30}, {"ten@TST", "hanadb.mydomain.int", 30041, "hcc_admin", "ReallzyKoolPassw0rd", true, 14, false, 0, false, true, 30}}}, false},
		{"Err_IdenticalNames", &Config{[]DbConfig{{"ten_DB", "hanadb.mydomain.int", 30015, "sstringer", "ReallzyKoolPassw0rd", true, 14, false, 0, false, true, 30}, {"ten_DB", "hanadb.mydomain.int", 30041, "hcc_admin", "ReallzyKoolPassw0rd", true, 14, false, 0, false, true, 30}}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.c.CheckForDupeNames(); (err != nil) != tt.wantErr {
				t.Errorf("Config.CheckForDupeNames() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
