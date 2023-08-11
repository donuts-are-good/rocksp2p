package main

import "flag"

// parseFlags This evaluates the flags used when the program was run
// and assigns the values of those flags according to sane defaults.
func flags() {
	// Define the log level flag
	logLevelFlag := flag.String("loglevel", "INFO", "Set the log level (DEBUG, INFO, WARNING, ERROR, CRITICAL)")

	// Define the p2p port
	flag.StringVar(&p2pPort, "p2pport", "33445", "P2P Port")

	flag.Parse()

	// Set the log level based on the flag value
	for i, name := range logLevelNames {
		if *logLevelFlag == name {
			logLevel = LogLevel(i)
			break
		}
	}
}
