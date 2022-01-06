package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

func GetConfigFromFile(path string) (Config, error) {
	var cnf Config
	file, err := os.Open(path)
	if err != nil {
		return cnf, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&cnf)
	if err != nil {
		log.Printf("Failed to decode!")
		return cnf, fmt.Errorf("failed to decode")
	}

	return cnf, nil
}
