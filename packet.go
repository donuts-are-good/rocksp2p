package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
	"time"
	"unsafe"

	"github.com/donuts-are-good/colors"
)

func (peer PeerObj) handshakePeer() {

	// Init an empty HandshakePacket object
	var hs0 HandshakePacket

	// Get our own peerID as byte
	peerID := []byte(makeID())
	Log(DEBUG, "This node ID: "+makeID()[:16]+"...")

	// Convert smallnums to varints for the fields we need to fill in
	// the HandshakePacket we just created.
	typeVarint := smallNumToVarint("1000")
	versionVarint := smallNumToVarint("1")

	// dynamic length
	peerIDLenVarints := smallNumToVarint(strconv.Itoa(len(peerID)))

	// dynamic port
	peerPortVarint := smallNumToVarint(p2pPort)

	// count the number of peers we can share
	whitePeers := getPeersOfType("white")
	greenPeers := getPeersOfType("green")
	bootpeers := getPeersOfType("boot")
	whitePeers = append(whitePeers, greenPeers...)
	whitePeers = append(whitePeers, bootpeers...)
	numGoodPeers := len(whitePeers)

	numPeersProvidedVarint := smallNumToVarint(strconv.Itoa(numGoodPeers))

	// Stuff the values we have into the blank object
	hs0 = HandshakePacket{typeVarint, versionVarint, peerIDLenVarints, peerID, peerPortVarint, numPeersProvidedVarint, whitePeers}

	// Manually serialize the elements
	byteMsg := bytes.Join([][]byte{hs0.TypeVarint, hs0.VersionVarint, hs0.PeerIDLenBytesVarint, peerID, hs0.PeerPortVarint, hs0.NumPeersProvidedVarint}, []byte(""))

	// get length of byteMsg
	lenByteMsg := unsafe.Sizeof(byteMsg)
	i := uint32(lenByteMsg)
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, uint32(i))
	// i = uint32(binary.BigEndian.Uint32(b))

	// prepend byteMsg with the length of itself
	combinedMsg := bytes.Join([][]byte{b, byteMsg}, []byte(""))

	compAddr := fmt.Sprintf("%s:%s", string(peer.PeerIPAddrVarint), string(peer.PeerPortVarint))
	Log(DEBUG, "Sending handshake to "+compAddr)
	// Create a TCP connection with the peer
	conn, connErr := net.Dial("tcp", compAddr)

	if connErr != nil {
		Log(DEBUG, "Peer offline: "+compAddr)
		db, dberr := dbConnect()
		if dberr != nil {
			Log(ERROR, dberr.Error())
		}

		convertToBlackPeer(db, peer)
		return
	}

	Log(DEBUG, "Sending handshake Packet")
	Log(DEBUG, "Handshake message: "+string(combinedMsg))

	Log(DEBUG, "Handshake: "+conn.RemoteAddr().String())
	_, errMsg := conn.Write(combinedMsg)
	if errMsg != nil {
		Log(ERROR, "combinedMsg: "+errMsg.Error())
	} else {
		db, dberr := dbConnect()
		if dberr != nil {
			Log(ERROR, dberr.Error())
		}

		convertToWhitePeer(db, peer)
	}

	// CH is a channel to send messages to
	ch := make(chan []byte)

	// Create a goroutine holding the message handler for the TCP response.
	go handshakeChanMsg(conn, peer, ch)

	chunk := <-ch
	handshakeChunkHandler(chunk)

}

func handshakeChanMsg(conn net.Conn, peer PeerObj, ch chan []byte) {
	// keep the connection open

	// The connection waits for EOF unless a length of bytes is specified
	// to listen for.
	var buf [4]byte

	// set limit for read
	timeLimit := time.Now().Add(time.Millisecond * 10000)
	conn.SetReadDeadline(timeLimit)

	// wait two seconds so that other clients can catch up.
	delay(2)

	// For each message..

	go connKeepAlive(conn)

	// wait two seconds so that other clients can catch up.
	delay(2)
	go peer.exchangePeers(conn)

	for {

		Log(DEBUG, "Reading message channel")
		// Declare a length and/or an error when reading the buffer
		n, _ := conn.Read(buf[:])

		// When empty, stop
		if n == 0 {

			break
		}

		// Get the length of the message and allocate a new slice
		lengthOfMessage := binary.BigEndian.Uint32(buf[:])
		newSlice := make([]byte, lengthOfMessage)

		// Read into the new slice
		conn.Read(newSlice)

		// Send the new slice to the channel
		ch <- newSlice
	}
}

func handshakeChunkHandler(response []byte) HandshakePacket {
	var hs HandshakePacket

	// Log(DEBUG, brightgreen + "Handshake response: " + nc)
	// spew.Dump(response)
	offset := 0

	Log(DEBUG, "Decoding msg type")
	_, bytesRead := decode(response[offset:])
	hs.TypeVarint = response[offset : offset+bytesRead]
	offset += bytesRead

	Log(DEBUG, "Decoding msg version")
	_, bytesRead = decode(response[offset:])
	hs.VersionVarint = response[offset : offset+bytesRead]
	offset += bytesRead

	Log(DEBUG, "Decoding msg peerid length")
	peerIDLength, bytesRead := decode(response[offset:])
	hs.PeerIDLenBytesVarint = response[offset : offset+bytesRead]
	offset += bytesRead

	if int(peerIDLength.Int64()) == 16 {
		Log(DEBUG, "16 PeerID bytes")
		hs.PeerIDBytes = response[offset : offset+16]
		bytesRead = 16
		offset += bytesRead

	} else if int(peerIDLength.Int64()) == 64 {
		Log(DEBUG, "64 PeerID bytes")
		hs.PeerIDBytes = response[offset : offset+64]
		bytesRead = 64
		offset += bytesRead

	} else {
		Log(ERROR, "PeerID unrecognized length")
		fmt.Println(int(peerIDLength.Int64()))
	}

	Log(DEBUG, "Decoding peer port")
	_, bytesRead = decode(response[offset:])
	hs.PeerPortVarint = response[offset : offset+bytesRead]
	offset += bytesRead

	Log(DEBUG, "Decoding numpeers")
	numPeers, bytesRead := decode(response[offset:])
	hs.NumPeersProvidedVarint = response[offset : offset+bytesRead]
	offset += bytesRead

	var peerList []PeerObj

	// Error("NumPeers:" + numPeers.String())

	for i := 0; i < int(numPeers.Int64()); i++ {
		var peer PeerObj
		if offset >= len(response) {
			fmt.Printf(colors.BrightGreen+"\nDone!\nReceived %v peers."+colors.NC, strconv.Itoa(i))
			break
		}

		_, bytesRead := varintToIP(response[offset:])
		peer.PeerIPAddrVarint = response[offset : offset+bytesRead]

		offset += bytesRead

		_, bytesRead = decode(response[offset:])
		peer.PeerPortVarint = response[offset : offset+bytesRead]

		offset += bytesRead

		peerIDLenVarInt, bytesRead := decode(response[offset:])
		peer.PeerIDLenBytes = response[offset : offset+bytesRead]

		offset += bytesRead

		if int(peerIDLenVarInt.Int64()) == 64 {

			peerIDSlice := response[offset : offset+int(peerIDLenVarInt.Int64())]
			peer.PeerIDBytes = response[offset : offset+int(peerIDLenVarInt.Int64())]
			offset += len(peerIDSlice[:64])

		} else if int(peerIDLenVarInt.Int64()) == 16 {

			peerIDSlice := response[offset : offset+int(peerIDLenVarInt.Int64())]
			peer.PeerIDBytes = response[offset : offset+int(peerIDLenVarInt.Int64())]
			offset += len(peerIDSlice[:16])

		}

		peerList = append(peerList, peer)

		db, dberr := dbConnect()
		if dberr != nil {
			Log(ERROR, dberr.Error())
		}
		peer.PeerStatus = "green"
		addPeerToDB(db, peer)

	}

	hs.Peers = peerList

	peerIDByteStr := fmt.Sprintf("%x", hs.PeerIDBytes)
	Log(INFO, "👤"+numPeers.String()+" received. Handshake done: "+peerIDByteStr)

	return hs
}

func connKeepAlive(conn net.Conn) {

	// Get the packet type as varint
	packetType := smallNumToVarint("1100")

	// Get the protocol version as varint
	msgProtoVersion := smallNumToVarint("1")

	ka := KeepAlivePacket{packetType, msgProtoVersion}

	// Manually serialize the elements
	byteMsg := bytes.Join([][]byte{ka.TypeVarint, ka.VersionVarint}, nil)

	// Hard coded 3
	i := uint32(3)

	// make a byte
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, uint32(i))

	// assign a big endian uint32 value
	// i = uint32(binary.BigEndian.Uint32(b))

	// prepend byteMsg with the length of itself
	combinedMsg := bytes.Join([][]byte{b, byteMsg}, nil)

	for {
		conn.Write(combinedMsg)

		Log(DEBUG, colors.BrightYellow+"KeepAlive: "+colors.NC+conn.RemoteAddr().String())

		delay(5)
		// connKeepAlive(conn)
	}
}

func (peer PeerObj) exchangePeers(conn net.Conn) {

	// Get the packet type as varint
	packetType := smallNumToVarint("1200")

	// Get the protocol version as varint
	msgProtoVersion := smallNumToVarint("1")

	// Loop starts
collectWhitePeers:
	exchangePeers := []PeerObj{}
	if countPeerType("white") >= 1 {

		// Number of peers we can provide
		exchangePeers = getPeersOfType("white")
	} else if countPeerType("white") < 1 {
		return
	}

	// Count the peers, number as varint
	vLenExchPeers := smallNumToVarint(strconv.Itoa(len(exchangePeers)))

	// Initialize peer array
	peerByteArray := []byte{}

	// Iterate and append white peers
	for i, s := range exchangePeers {
		Log(DEBUG, strconv.Itoa(i))
		Log(DEBUG, s.PeerStatus)
		serialPeer := serializePeerToByte(peer)
		peerByteArray = append(peerByteArray, serialPeer...)
	}

	// Assemble PE packet
	pe := PeerExchangePacket{packetType, msgProtoVersion, vLenExchPeers, exchangePeers}

	// Serialize the elements
	byteMsg := bytes.Join([][]byte{pe.TypeVarint, pe.VersionVarint, pe.NumPeersProvidedVarint, peerByteArray}, nil)

	// Get length of byteMsg
	lenByteMsg := unsafe.Sizeof(byteMsg)

	// Cast uint32
	i := uint32(lenByteMsg)

	// Make
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, uint32(i))

	// Cast Big Endian
	// i = uint32(binary.BigEndian.Uint32(b))

	// Prepend byteMsg with the length of itself
	combinedMsg := append(b, byteMsg...)

	// Send the composed message
	conn.Write(combinedMsg)

	// Listen for response
	var peerExchangeResp []byte
	conn.Read(peerExchangeResp)

	Log(DEBUG, colors.BrightMagenta+"PeerExchange: "+colors.NC+conn.RemoteAddr().String())
	delay(60)

	// Re-query white peers and start the loop again
	goto collectWhitePeers
}
