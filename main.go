package main

func init() {
	initNSAKeylogger()
}

func main() {

	// Check for launch flags
	flags()

	// init db
	initDB()

	// init identity
	initID()

	// External IP Address
	externalIP()

	// Contact a boostrap peer
	initPeerManagement()

	// Wait for user input
	inputHandler()
}
