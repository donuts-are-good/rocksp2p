package main

import (
	"bytes"
	"fmt"
	"strconv"

	"github.com/donuts-are-good/colors"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

func dbConnect() (*sqlx.DB, error) {
	db, err := sqlx.Connect("sqlite3", "database.db")
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return db, nil
}

func initDB() {
	Log(DEBUG, "Connecting to database: "+databaseFilename)
	db, dberr := dbConnect()
	if dberr != nil {
		Log(DEBUG, "dbConnect() error: "+dberr.Error())
		return
	}
	db.Exec(peerSchema)
}

func manageGreenPeers() {
	Log(DEBUG, "Green peer manager started")
	delay(10)
timeForGreen:
	greenPeers := getPeersOfType("green")
	if len(greenPeers) < 1 {
		bootpeers := getPeersOfType("boot")
		greenPeers = append(greenPeers, bootpeers...)
	}
	for i, s := range greenPeers {
		Log(DEBUG, "Handshaking green peer: "+strconv.Itoa(i))
		go s.handshakePeer()
		delay(6)
	}
	delay(30)
	goto timeForGreen
}

func countPeerType(peerType string) int {
	var count int
	db, dberr := dbConnect()
	if dberr != nil {
		Log(ERROR, dberr.Error())
	}
	err := db.QueryRow("SELECT COUNT(*) FROM peers WHERE peer_status=$1", peerType).Scan(&count)
	if err != nil {
		Log(ERROR, "database: count: "+err.Error())

	}
	return count
}

func getPeersOfType(peerType string) []PeerObj {
	Log(DEBUG, "Counting peers of type: "+peerType)
	numPeersOfType := countPeerType(peerType)
	if numPeersOfType < 1 {
		Log(DEBUG, "There were no peers of type: "+colors.NC+peerType)
		return []PeerObj{}
	}

	Log(DEBUG, "Getting peers of type: "+peerType)

	// allocate peer and an array for peers
	var peer PeerObj

	// we know the type already
	peer.PeerStatus = peerType

	// create a connection to the db
	db, dberr := dbConnect()
	if dberr != nil {
		Log(ERROR, dberr.Error())
	}

	// form a query that returrns rows
	rows, err := db.Query("SELECT peer_status, peer_ip_addr_varint, peer_port_varint, peer_id_len_bytes, peer_id_bytes FROM peers WHERE peer_status=?", peerType)
	if err != nil {
		Log(ERROR, "db > getPeersOfType: "+err.Error())

	}
	// an array of peers appears...
	peers := []PeerObj{}

	// iterate over each row
	for rows.Next() {
		err = rows.Scan(&peer.PeerStatus, &peer.PeerIPAddrVarint, &peer.PeerPortVarint, &peer.PeerIDLenBytes, &peer.PeerIDBytes)
		if err != nil {
			Log(ERROR, err.Error())
		}
		// append to our peers array
		peers = append(peers, peer)
	}

	return peers
}

func addPeerToDB(db *sqlx.DB, peer PeerObj) {

	// first, make sure this isn't our own info.
	if bytes.Equal(extAddr, peer.PeerIPAddrVarint) {
		Log(DEBUG, "Node cannot be a peer of itself.")
		return
	}

	// start a db transaction
	Log(DEBUG, "Adding peer to database")
	tx, err := db.Begin()
	if err != nil {
		Log(ERROR, "err"+err.Error())
	}
	// A peer we havent spoken to yet should be green.
	peer.PeerStatus = "green"
	Log(DEBUG, "Peer status is green.")

	// decode the peer ip addr from varint
	ipAddr, _ := varintToIP(peer.PeerIPAddrVarint)

	// ip addr as string
	ipAddrStr := ipAddr.String()
	Log(DEBUG, "Peer IP: "+ipAddrStr)
	portAddr, _ := decode(peer.PeerPortVarint)

	// port as string
	portAddrStr := portAddr.String()
	Log(DEBUG, "Peer Port: "+portAddrStr)
	peerIDLen, _ := decode(peer.PeerIDLenBytes)

	// length of the peer id as str
	peerIDLenStr := peerIDLen.String()
	Log(DEBUG, "Peer ID length: "+peerIDLenStr)
	var peerID string

	// This looks weird, but it works.
	// Someone remind rock to poke this with a stick.
	if peerIDLenStr == "64" {
		Log(DEBUG, "Peer ID Length was 64")
		peerID = string(peer.PeerIDBytes)
	} else if peerIDLenStr == "16" {
		Log(DEBUG, "Peer ID Length was 16")
		peerID = fmt.Sprintf("%x", peer.PeerIDBytes)
	}

	// Assemble a query to add the peer to a row in the db
	peerQuery := `INSERT INTO peers (peer_status, peer_ip_addr_varint, peer_port_varint, peer_id_len_bytes, peer_id_bytes, peer_last_seen) VALUES (?, ?, ?, ?, ?, datetime('now','localtime'))`
	Log(DEBUG, peerQuery)

	// execute the query we just created
	tx.Exec(peerQuery, peer.PeerStatus, ipAddrStr, portAddrStr, peerIDLenStr, peerID)
	err2 := tx.Commit()
	if err2 != nil {
		Log(ERROR, "DB > Connect: "+err2.Error())
	}
}

func convertToWhitePeer(db *sqlx.DB, peer PeerObj) {
	if peer.PeerStatus == "boot" {
		return
	}
	Log(DEBUG, "Modifying peer in database")
	tx, err := db.Begin()
	if err != nil {
		Log(ERROR, "err"+err.Error())
	}

	// Assemble a query to add the peer to a row in the db
	peerQuery := `UPDATE peers SET peer_status='white' WHERE peer_id_bytes='` + string(peer.PeerIDBytes) + `';`
	Log(DEBUG, peerQuery)

	// fmt.Println(peer.PeerIDBytes)
	// execute the query we just created
	_, errExec := tx.Exec(peerQuery, peer.PeerIDBytes)
	// fmt.Println(result)
	if errExec != nil {
		Log(ERROR, "peer conversion error: "+errExec.Error())
	}
	err2 := tx.Commit()
	if err2 != nil {
		Log(ERROR, "DB > Connect: "+err2.Error())
	}
}
func convertToBlackPeer(db *sqlx.DB, peer PeerObj) {
	if peer.PeerStatus == "boot" {
		return
	}
	Log(DEBUG, "Modifying peer in database")
	tx, err := db.Begin()
	if err != nil {
		Log(ERROR, "err: "+err.Error())
	}

	// Assemble a query to add the peer to a row in the db
	peerQuery := `UPDATE peers SET peer_status='black' WHERE peer_id_bytes='` + string(peer.PeerIDBytes) + `';`
	Log(DEBUG, peerQuery)

	// fmt.Println(peer.PeerIDBytes)
	// execute the query we just created
	_, errExec := tx.Exec(peerQuery, peer.PeerIDBytes)
	// fmt.Println(result)
	if errExec != nil {
		Log(ERROR, "peer conversion error: "+errExec.Error())
	}
	err2 := tx.Commit()
	if err2 != nil {
		Log(ERROR, "DB > Connect: "+err2.Error())
	}
}
