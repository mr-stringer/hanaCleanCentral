package main

import (
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
