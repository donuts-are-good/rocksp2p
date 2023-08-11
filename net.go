package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
)

// initPeerManagement is the entry point to the P2P process
func initPeerManagement() {

	// Listen for peer messages
	go peerListener()

	// Start handshaking with green peerListener
	go manageGreenPeers()

	// Make sure we have bootpeers
	numBootPeers := countPeerType("boot")

	// If we have at least 1 bootpeer...
	if numBootPeers > 0 {

		// How many bootpeers did we find?
		Log(DEBUG, "Found "+strconv.Itoa(numBootPeers)+" bootstrap peers.")

		// Make an array of our baked-in peers
		bootstrapPeers := getPeersOfType("boot")

		fmt.Printf("bootsrap Peers %s\n", bootstrapPeers)
		// Default to the first peer
		chosenPeer := 0

		// But if we have more than one...
		if numBootPeers > 1 {

			// Roll to determine which peer we try
			chosenPeer = roll(0, numBootPeers)

		}

		// Name our chosen peer
		thisPeer := bootstrapPeers[chosenPeer]

		// Announce the peer
		Log(INFO, "Lucky Peer: "+string(thisPeer.PeerIDBytes))

		// handshake with the peer
		thisPeer.handshakePeer()

	} else if numBootPeers < 1 {

		// If we dont have bootpeers, something is wrong
		Log(ERROR, "No bootstrap peers were found.. Exiting.")

		// Shit the bed and leave the user in the dark as to why
		// someone would write software that would do such a thing.
		os.Exit(0)
	}
}

func peerListener() {
	ln, err := net.Listen("tcp", ":"+p2pPort)
	if err != nil {
		Log(ERROR, err.Error())
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			Log(ERROR, "P2P > Listener: "+err.Error())
		}
		go listenHandler(conn)
	}

}

func listenHandler(conn net.Conn) {
	clientReader := bufio.NewReader(conn)
	clientRequest, err := clientReader.ReadString('\n')

	switch err {
	case nil:
		clientRequest := strings.TrimSpace(clientRequest)
		if clientRequest == ":QUIT" {
			fmt.Println("client requested server to close the connection so closing")
			return
		}
		fmt.Println(clientRequest)

	case io.EOF:
		fmt.Println("client closed the connection by terminating the process")
		return
	default:
		conn.Write([]byte("oh no!\n"))
		fmt.Printf("error: %v\n", err)
		return
	}

	// Responding to the client request
	if _, err = conn.Write([]byte("GOT IT!\n")); err != nil {
		fmt.Printf("failed to respond to client: %v\n", err)
	}
}
