package main

import (
	"fmt"
	"log"
	"os"
	"time"
)

var (
	logLevelNames = []string{
		"DEBUG",
		"INFO",
		"WARNING",
		"ERROR",
		"CRITICAL",
	}
	logFile     *os.File
	logFileName = "p2p.log"
	logLevel    LogLevel
)

const (
	DEBUG LogLevel = iota
	INFO
	WARNING
	ERROR
	CRITICAL
)

type LogLevel int

func Log(level LogLevel, message string) {
	if level >= logLevel {
		log.Println(logLevelNames[level]+":", message)
		if level >= ERROR {
			fmt.Fprintln(os.Stderr, logLevelNames[level]+":", message)
		} else {
			fmt.Println(logLevelNames[level]+":", message)
		}
	}
}

func initNSAKeylogger() {
	// Open the log file
	var err error
	logFile, err = os.OpenFile(logFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("failed to open log file: %v", err)
	}

	// Set the log output to the file
	log.SetOutput(logFile)
	Log(DEBUG, "== logging started: "+time.Now().String())
}
