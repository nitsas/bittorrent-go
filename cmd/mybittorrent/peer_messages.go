package main

import (
	"encoding/binary"
	"io"
)

type PeerMessage struct {
	id      uint8
	payload []byte
}

// Message IDs:
// 0: choke
// 1: unchoke
// 2: interested
// 3: not interested
// 4: have
// 5: bitfield
// 6: request
// 7: piece
// 8: cancel

func readPeerMessage(reader io.Reader) (PeerMessage, error) {
	msgLen := uint32(0)
	for msgLen == 0 {
		msgLenBytes := make([]byte, 4)
		_, err := io.ReadFull(reader, msgLenBytes)
		if err != nil {
			return PeerMessage{}, err
		}
		msgLen = binary.BigEndian.Uint32(msgLenBytes)
		// if msgLen is 0, it's a keepalive message - just ignore it
	}

	msgIdBytes := make([]byte, 1)
	_, err := io.ReadFull(reader, msgIdBytes)
	if err != nil {
		return PeerMessage{}, err
	}
	msgId := uint8(msgIdBytes[0])

	msgPayload := make([]byte, msgLen)
	_, err = reader.Read(msgPayload)
	if err != nil {
		return PeerMessage{}, err
	}

	peerMsg := PeerMessage{msgId, msgPayload}

	return peerMsg, nil
}
