package main

// BootPeers is a struct that holds a given number of bootstrap peers
type BootPeers struct {
	Peers []PeerObj
}

// HandshakePacket is a struct for the initial packet sent upon first
// connecting to a bootstrap peer
type HandshakePacket struct {
	TypeVarint             []byte    `json:"type_varint"`
	VersionVarint          []byte    `json:"version_varint"`
	PeerIDLenBytesVarint   []byte    `json:"peer_id_len_bytes_varint"`
	PeerIDBytes            []byte    `json:"peer_id_bytes"`
	PeerPortVarint         []byte    `json:"peer_port_varint"`
	NumPeersProvidedVarint []byte    `json:"num_peers_provided_varint"`
	Peers                  []PeerObj `json:"peers"`
}

// KeepAlivePacket is the bare minimum packet, used to keep a connection open
type KeepAlivePacket struct {
	TypeVarint    []byte `json:"type_varint"`
	VersionVarint []byte `json:"version_varint"`
}

// PeerExchangePacket is a packet to do peer exchange
type PeerExchangePacket struct {
	TypeVarint             []byte    `json:"type_varint"`
	VersionVarint          []byte    `json:"version_varint"`
	NumPeersProvidedVarint []byte    `json:"num_peers_provided_varint"`
	Peers                  []PeerObj `json:"peers"`
}

// PeerObj is the structure of a peer message payload
type PeerObj struct {
	PeerStatus       string `json:"peer_status"`
	PeerIPAddrVarint []byte `json:"peer_ip_addr_varint"`
	PeerPortVarint   []byte `json:"peer_port_varint"`
	PeerIDLenBytes   []byte `json:"peer_id_len_bytes"`
	PeerIDBytes      []byte `json:"peer_id_bytes"`
}

// Packet is the generic packet constructor interface
type Packet interface {
	Handshake() HandshakePacket
	Keepalive() KeepAlivePacket
	PeerExchange() PeerExchangePacket
}
