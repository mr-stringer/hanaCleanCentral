package main

import (
	"testing"
)

func TestConfig_CheckForDupeNames(t *testing.T) {

	tests := []struct {
		name    string
		c       *Config
		wantErr bool
	}{
		{"Good_SingleDB", &Config{true, 60, true, 60, true, true, 60, true, true, 60, true, []DbConfig{{"systemdb_TST", "hanadb.mydomain.int", 30015, "hccuser", "ReallyCoolPassw0rd", true, 60, true, 60, true, true, 60, true, true, 60, true, nil, CleanResults{}}}}, false},
		{"Good_TwoDBs", &Config{true, 60, true, 60, true, true, 60, true, true, 60, true, []DbConfig{{"systemdb_TST", "hanadb.mydomain.int", 30015, "hccuser", "ReallyCoolPassw0rd", true, 60, true, 60, true, true, 60, true, true, 60, true, nil, CleanResults{}}, {"ten1_TST", "hanadb.mydomain.int", 30041, "hccuser", "ReallyCoolPassw0rd", false, 0, false, 0, true, true, 90, false, true, 30, true, nil, CleanResults{}}}}, false},
		{"Err_IdenticalNames", &Config{true, 60, true, 60, true, true, 60, true, true, 60, true, []DbConfig{{"database", "hanadb.mydomain.int", 30015, "hccuser", "ReallyCoolPassw0rd", true, 60, true, 60, true, true, 60, true, true, 60, true, nil, CleanResults{}}, {"database", "hanadb.mydomain.int", 30041, "hccuser", "ReallyCoolPassw0rd", false, 0, false, 0, true, true, 90, false, true, 30, true, nil, CleanResults{}}}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.c.CheckForDupeNames(); (err != nil) != tt.wantErr {
				t.Errorf("Config.CheckForDupeNames() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfig_PrintConfig(t *testing.T) {
	/*Sort of hard to test for faults here as we need valid json to work with in the first place*/
	tests := []struct {
		name    string
		c       *Config
		wantErr bool
	}{
		{"GoodSingleDB", &Config{true, 60, true, 60, true, true, 60, true, true, 60, true, []DbConfig{{"systemdb_TST", "hanadb.mydomain.int", 30015, "hccuser", "ReallyCoolPassw0rd", true, 60, true, 60, true, true, 60, true, true, 60, true, nil, CleanResults{}}}}, false},
		{"GoodTwoDBs", &Config{true, 60, true, 60, true, true, 60, true, true, 60, true, []DbConfig{{"systemdb_TST", "hanadb.mydomain.int", 30015, "hccuser", "ReallyCoolPassw0rd", true, 60, true, 60, true, true, 60, true, true, 60, true, nil, CleanResults{}}, {"ten1_TST", "hanadb.mydomain.int", 30041, "hccuser", "ReallyCoolPassw0rd", false, 0, false, 0, true, true, 90, false, true, 30, true, nil, CleanResults{}}}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.c.PrintConfig(); (err != nil) != tt.wantErr {
				t.Errorf("Config.PrintConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
