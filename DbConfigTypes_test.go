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
		{"good001", DbConfig{"test", "localhost", 30015, "admin", "password"}, "hdb://admin:password@localhost:30015"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.hdb.Dsn(); got != tt.want {
				t.Errorf("DbConfig.GetDsn() = %v, want %v", got, tt.want)
			}
		})
	}
}
