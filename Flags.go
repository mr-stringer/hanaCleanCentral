package main

import "flag"

func ProcessFlags() AppConfig {
	var config string
	var verbose bool
	var dryrun bool

	flag.StringVar(&config, "ConfigFile", "config.json", "The location of the configuration file")
	flag.BoolVar(&verbose, "v", false, "Verbose - When true, verbose logging is enabled.")
	flag.BoolVar(&dryrun, "d", false, "Dry Run - When true, no changes will be made the database/s")

	flag.Parse()

	return AppConfig{config, verbose, dryrun}
}
