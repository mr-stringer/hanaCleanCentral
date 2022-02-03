package main

import "testing"

func TestConfig_CheckForDupeNames(t *testing.T) {

	tests := []struct {
		name    string
		c       *Config
		wantErr bool
	}{
		{"Good_SingleDB", &Config{true, 60, true, 60, true, true, 60, true, true, 60, true, []DbConfig{{"systemdb_TST", "hanadb.mydomain.int", 30015, "hccuser", "ReallzyKoolPassw0rd", true, 60, true, 60, true, true, 60, true, true, 60, true}}}, false},
		{"Good_TwoDBs", &Config{true, 60, true, 60, true, true, 60, true, true, 60, true, []DbConfig{{"systemdb_TST", "hanadb.mydomain.int", 30015, "hccuser", "ReallzyKoolPassw0rd", true, 60, true, 60, true, true, 60, true, true, 60, true}, {"ten1_TST", "hanadb.mydomain.int", 30041, "hccuser", "ReallzyKoolPassw0rd", false, 0, false, 0, true, true, 90, false, true, 30, true}}}, false},
		{"Err_IndenticalNames", &Config{true, 60, true, 60, true, true, 60, true, true, 60, true, []DbConfig{{"database", "hanadb.mydomain.int", 30015, "hccuser", "ReallzyKoolPassw0rd", true, 60, true, 60, true, true, 60, true, true, 60, true}, {"database", "hanadb.mydomain.int", 30041, "hccuser", "ReallzyKoolPassw0rd", false, 0, false, 0, true, true, 90, false, true, 30, true}}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.c.CheckForDupeNames(); (err != nil) != tt.wantErr {
				t.Errorf("Config.CheckForDupeNames() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
