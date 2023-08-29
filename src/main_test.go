package main

import (
	"io"
	"log"
	"os"
	"testing"
)

/*Disable logging during testing*/
func TestMain(m *testing.M) {
	log.SetOutput(io.Discard)
	os.Exit(m.Run())
}
