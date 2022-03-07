package main

import "flag"

func ProcessFlags() AppConfig {
	var config string
	var verbose bool
	var dryrun bool
	var printconfig bool

	flag.StringVar(&config, "f", "config.json", "The location of the configuration file")
	flag.BoolVar(&verbose, "v", false, "Verbose - When true, verbose logging is enabled.")
	flag.BoolVar(&dryrun, "d", false, "Dry Run - When true, no changes will be made the database/s")
	flag.BoolVar(&printconfig, "p", false, "Print Effective Config - When true, the application configuration is printed to screen and the application quits\nPasswords will not be printed!")

	flag.Parse()

	return AppConfig{config, verbose, dryrun, printconfig}
}
