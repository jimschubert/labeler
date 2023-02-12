/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"os"

	"github.com/jimschubert/labeler/cmd/labeler"
	log "github.com/sirupsen/logrus"
)

func main() {
	initLogging()
	labeler.Execute()
}

func initLogging() {
	logLevel, ok := os.LookupEnv("LOG_LEVEL")
	if !ok {
		logLevel = "error"
	}
	ll, err := log.ParseLevel(logLevel)
	if err != nil {
		ll = log.DebugLevel
	}
	log.SetLevel(ll)
	log.SetOutput(os.Stderr)
}
