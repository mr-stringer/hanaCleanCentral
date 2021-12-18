package main

import "fmt"

//Struct for holding database configuration
type DbConfig struct {
	Name     string
	Hostname string
	Port     uint16
	Username string
	Password string
}

func (hdb DbConfig) GetDsn() string {
	return fmt.Sprintf("hdb://%s:%s@%s:%d", hdb.Username, hdb.Password, hdb.Hostname, hdb.Port)
}
