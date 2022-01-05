package main

//Top level configuration for hanaCleanCentral
type Config struct {
	Databases []DbConfig
}

//Application configuration parameters to be shared with functions
type AppConfig struct {
	ConfigFile string //the location of the config file
	Verbose    bool   //used for verbose logging
	DryRun     bool   //used for non-destructive testing
}
