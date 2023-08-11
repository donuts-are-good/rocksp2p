package main

// // Directories
var (
	databaseFilename string = "./config/p2p/RocksP2P.db"
)

// Network
var (
	extAddr    = []byte("")
	extAddrStr = ""
	p2pPort    = "33445"
)

// Database vars
var (
	peerSchema = `

	CREATE TABLE IF NOT EXISTS peers (
    peer_status text,
    peer_ip_addr_varint text,
    peer_port_varint text,
    peer_id_len_bytes text,
		peer_id_bytes text,
		peer_last_seen DATE DEFAULT (datetime('now','localtime')),
		unique (peer_ip_addr_varint, peer_id_bytes));

		INSERT INTO "peers" ("peer_status", "peer_ip_addr_varint", "peer_port_varint", "peer_id_len_bytes", "peer_id_bytes") VALUES ('boot', '24.199.103.63', '33445', '32', '1e9a6130a71af1652a262ee9b92d0eb7ba6e948b8391a6d5b58d68c0bc282a46');

		CREATE TRIGGER [update_last_seen]
			AFTER UPDATE
			ON peers
			FOR EACH ROW
			BEGIN
			UPDATE peers SET peer_last_seen = CURRENT_TIMESTAMP WHERE peer_last_seen = old.peer_last_seen;
			END

		`
)
