package main

import "log"

type LogMessage struct {
	Name    string
	Message string
	Verbose bool
}

//The logger function is responsible for the vast majority of logging.
//The function is aware of the application configuration for verbositiy so

func Logger(ac AppConfig, ch <-chan LogMessage, quit <-chan bool) {

	//forever loop
	for {
		select {
		case <-quit:
			return
		case lm := <-ch:
			if lm.Verbose && ac.Verbose {
				log.Printf("%s:%s\n", lm.Name, lm.Message)
			} else if !lm.Verbose {
				log.Printf("%s:%s\n", lm.Name, lm.Message)
			}
		}
	}
}
