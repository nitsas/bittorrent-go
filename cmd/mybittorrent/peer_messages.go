package main

import (
	"encoding/binary"
	"fmt"
	"io"
)

type pmid uint8

type PeerMessage struct {
	id      pmid
	payload []byte
}

const (
	pmidChoke         pmid = 0
	pmidUnchoke       pmid = 1
	pmidInterested    pmid = 2
	pmidNotInterested pmid = 3
	pmidHave          pmid = 4
	pmidBitfield      pmid = 5
	pmidRequest       pmid = 6
	pmidPiece         pmid = 7
	pmidCancel        pmid = 8
)

func readPeerMessage(reader io.Reader) (PeerMessage, error) {
	msgLen := uint32(0)

	for msgLen == 0 {
		msgLenBytes := make([]byte, 4)
		_, err := io.ReadFull(reader, msgLenBytes)
		if err != nil {
			return PeerMessage{}, err
		}
		msgLen = binary.BigEndian.Uint32(msgLenBytes)
		// If msgLen is 0, it's a keepalive message, and we should ignore it.
	}

	msgPayloadBytes := make([]byte, msgLen)
	_, err := io.ReadFull(reader, msgPayloadBytes)
	if err != nil {
		return PeerMessage{}, err
	}
	msgId := pmid(msgPayloadBytes[0])

	peerMsg := PeerMessage{msgId, msgPayloadBytes[1:]}

	return peerMsg, nil
}

func sendPeerMessage(writer io.Writer, msg PeerMessage) error {
	msgLenBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(msgLenBytes, uint32(1+len(msg.payload)))

	bytesWritten, err := writer.Write(msgLenBytes)
	if err == nil && bytesWritten < len(msgLenBytes) {
		err = fmt.Errorf("sendPeerMessage: wrote %d bytes instead of %d", bytesWritten, len(msgLenBytes))
	}

	bytesWritten, err = writer.Write([]byte{byte(msg.id)})
	if err == nil && bytesWritten < 1 {
		err = fmt.Errorf("sendPeerMessage: wrote %d bytes instead of %d", bytesWritten, 1)
	}

	bytesWritten, err = writer.Write(msg.payload)
	if err == nil && bytesWritten < len(msg.payload) {
		err = fmt.Errorf("sendPeerMessage: wrote %d bytes instead of %d", bytesWritten, len(msg.payload))
	}

	return err
}
