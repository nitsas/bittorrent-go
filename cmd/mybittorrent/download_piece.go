package main

import (
	"encoding/binary"
	"fmt"
	"net"
	"sort"
)

type Piece []byte

func DownloadPiece(conn net.Conn, pieceIndex int, pieceLength int) (Piece, error) {
	interestedMsg := PeerMessage{pmidInterested, []byte{}}
	err := sendPeerMessage(conn, interestedMsg)
	if err != nil {
		return nil, err
	}
	fmt.Printf("Sent 'interested' msg to the peer!\n")

	peerMsg := PeerMessage{pmidChoke, []byte{}}
	for peerMsg.id != pmidUnchoke {
		fmt.Printf("Waiting for 'unchoke' msg from the peer...\n")
		peerMsg, err = readPeerMessage(conn)
		if err != nil {
			return nil, err
		}
		fmt.Printf("Read peer msg with id: %d, payload: %x\n", peerMsg.id, peerMsg.payload)
	}

	const BlockSize = 16 * 1024

	fmt.Printf("Sending 'request' messages to peer:\n")
	numBlocks := 0
	blocks := make(map[int][]byte)
	for blockBegin := 0; blockBegin < pieceLength; blockBegin += BlockSize {
		numBlocks += 1

		blockLength := pieceLength - blockBegin
		if BlockSize < blockLength {
			blockLength = BlockSize
		}

		requestPayload := make([]byte, 12)
		binary.BigEndian.PutUint32(requestPayload[0:4], uint32(pieceIndex))
		binary.BigEndian.PutUint32(requestPayload[4:8], uint32(blockBegin))
		binary.BigEndian.PutUint32(requestPayload[8:12], uint32(blockLength))

		requestMsg := PeerMessage{pmidRequest, requestPayload}
		err = sendPeerMessage(conn, requestMsg)
		if err != nil {
			return nil, err
		}
		fmt.Printf("- Sent 'request' msg to peer: piece: %d, blockBegin: %d, requestPayload: %x\n",
			pieceIndex, blockBegin, requestPayload)
	}

	notInterestedMsg := PeerMessage{pmidNotInterested, []byte{}}
	err = sendPeerMessage(conn, notInterestedMsg)
	if err != nil {
		return nil, err
	}
	fmt.Printf("Sent 'not interested' msg to the peer!\n")

	fmt.Printf("Listening for 'piece' messages from peer:\n")
	for len(blocks) < numBlocks {
		peerMsg, err = readPeerMessage(conn)
		if err != nil {
			return nil, err
		}

		if peerMsg.id != 7 {
			fmt.Printf("- Read (and ignored) peer msg with id %d, payload: %x\n",
				peerMsg.id, peerMsg.payload)
			continue
		}

		pieceIndex := int(binary.BigEndian.Uint32(peerMsg.payload[0:4]))
		blockBegin := int(binary.BigEndian.Uint32(peerMsg.payload[4:8]))
		blockData := peerMsg.payload[8:]
		fmt.Printf("- Read 'piece' peer msg, pieceIndex: %d, blockBegin: %d, len(blockData): %d\n",
			pieceIndex, blockBegin, len(blockData))

		if _, ok := blocks[blockBegin]; !ok {
			blocks[blockBegin] = blockData
		}
	}

	blockBegins := make([]int, 0, len(blocks))
	for bb := range blocks {
		blockBegins = append(blockBegins, bb)
	}
	sort.Ints(blockBegins)

	piece := make([]byte, 0, pieceLength)
	for _, bb := range blockBegins {
		piece = append(piece, blocks[bb]...)
	}

	return piece, nil
}
