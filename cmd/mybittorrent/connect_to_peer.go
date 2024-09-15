package main

import (
	"fmt"
	"net"
)

func ConnectToPeer(peer Peer, infoHash []byte) (net.Conn, Bitfield, error) {
	peerIpPort := fmt.Sprintf("%s:%d", peer.Ip, peer.Port)

	conn, err := net.Dial("tcp", peerIpPort)
	if err != nil {
		return nil, nil, err
	}

	_, err = handshake(conn, infoHash)
	if err != nil {
		return nil, nil, err
	}

	fmt.Printf("Waiting for 'bitfield' msg from the peer...\n")
	peerMsg, err := readPeerMessage(conn)
	if err != nil {
		return nil, nil, err
	}
	fmt.Printf("Read peer msg with id %d, payload: %x\n", peerMsg.id, peerMsg.payload)

	var bitfield Bitfield
	if peerMsg.id == pmidBitfield {
		bitfield = peerMsg.payload
	}

	return conn, bitfield, nil
}
