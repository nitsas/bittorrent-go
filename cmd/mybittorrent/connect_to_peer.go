package main

import (
	"fmt"
	"net"
)

type PeerConn struct {
	Conn       net.Conn
	Interested bool
	Choked     bool
}

func (peerConn *PeerConn) Close() {
	peerConn.Conn.Close()
}

func NewPeerConn(conn net.Conn) PeerConn {
	return PeerConn{conn, false, true}
}

func ConnectToPeer(peer Peer, infoHash []byte) (PeerConn, Bitfield, error) {
	peerIpPort := fmt.Sprintf("%s:%d", peer.Ip, peer.Port)

	conn, err := net.Dial("tcp", peerIpPort)
	if err != nil {
		return PeerConn{}, nil, err
	}

	_, err = handshake(conn, infoHash)
	if err != nil {
		return PeerConn{}, nil, err
	}

	fmt.Printf("Waiting for 'bitfield' msg from the peer...\n")
	peerMsg, err := readPeerMessage(conn)
	if err != nil {
		return PeerConn{}, nil, err
	}
	fmt.Printf("Read peer msg with id %d, payload: %x\n", peerMsg.id, peerMsg.payload)

	var bitfield Bitfield
	if peerMsg.id == pmidBitfield {
		bitfield = peerMsg.payload
	}

	return NewPeerConn(conn), bitfield, nil
}
