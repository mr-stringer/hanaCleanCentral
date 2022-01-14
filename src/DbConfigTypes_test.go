package main

import (
	"os"
	"testing"
)

func TestDbConfig_Dsn(t *testing.T) {
	tests := []struct {
		name string
		hdb  DbConfig
		want string
	}{
		{"good001", DbConfig{"test", "localhost", 30015, "admin", "password", true, 14, true, 14, true, true, 30}, "hdb://admin:password@localhost:30015"},
		{"good002", DbConfig{"test", "dbserver.int.comp.net", 31013, "admin", "password", true, 14, true, 14, true, true, 30}, "hdb://admin:password@dbserver.int.comp.net:31013"},
		{"good003", DbConfig{"test", "nvfr111", 30015, "hccadmin", "345ertgfdG$", true, 14, true, 14, true, true, 30}, "hdb://hccadmin:345ertgfdG$@nvfr111:30015"},
		{"good004", DbConfig{"test", "localhost", 30015, "admin", "password", true, 14, true, 14, true, true, 30}, "hdb://admin:password@localhost:30015"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.hdb.Dsn(); got != tt.want {
				t.Errorf("DbConfig.GetDsn() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDbConfig_GetPasswordFromEnv(t *testing.T) {
	tests := []struct {
		name    string
		db      *DbConfig
		wantErr bool
	}{
		{"Bad_NoEnvVarSet", &DbConfig{"SystemDB_HN1", "hanadb.mydomain.int", 30015, "sstringer", "", true, 14, false, 0, false, true, 30}, true},
		{"Good_EnvVarSet", &DbConfig{"SystemDB_HN1", "hanadb.mydomain.int", 30015, "sstringer", "", true, 14, false, 0, false, true, 30}, false},
	}
	for _, tt := range tests {
		if tt.name == "Good_EnvVarSet" {
			os.Setenv("HCC_SystemDB_HN1", "!g03g3598fu254g36t")
		}
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.db.GetPasswordFromEnv(); (err != nil) != tt.wantErr {
				t.Errorf("DbConfig.GetPasswordFromEnv() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
