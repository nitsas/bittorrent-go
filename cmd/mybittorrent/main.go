package main

import (
	"crypto/sha1"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"sort"
	"strconv"
)

const PeerId = "c20e54494e34aa21c2af"

func panicIf(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	command := os.Args[1]

	switch command {
	case "decode":
		bencodedValue := os.Args[2]

		decoded, err := DecodeBencode(bencodedValue)
		panicIf(err)

		jsonOutput, _ := json.Marshal(decoded)
		fmt.Println(string(jsonOutput))
	case "info":
		t, infoHash, err := ParseTorrent(os.Args[2])
		panicIf(err)
		fmt.Printf("Tracker URL: %s\n", t.announce)
		fmt.Printf("Length: %d\n", t.info.length)
		fmt.Printf("Info Hash: %x\n", infoHash)
		fmt.Printf("Piece Length: %d\n", t.info.pieceLength)
		fmt.Println("Piece Hashes:")
		for _, p := range t.info.pieces {
			fmt.Printf("%x\n", p)
		}
	case "peers":
		torrFile := os.Args[2]
		torr, infoHash, err := ParseTorrent(torrFile)
		panicIf(err)
		trackerResp, err := TrackerRequest(torr, infoHash, PeerId)
		panicIf(err)
		for _, peer := range trackerResp.Peers {
			fmt.Printf("%s:%d\n", peer.Ip, peer.Port)
		}
	case "handshake":
		torrName := os.Args[2]
		peerIpPort := os.Args[3]

		_, infoHash, err := ParseTorrent(torrName)
		panicIf(err)

		conn, err := net.Dial("tcp", peerIpPort)
		panicIf(err)
		defer conn.Close()

		err = writeHandshake(conn, infoHash)
		panicIf(err)

		peerId, err := readHandshake(conn)
		panicIf(err)
		fmt.Printf("Peer ID: %x\n", peerId)
	case "download_piece":
		usageString := fmt.Sprintf("Usage: %s download_piece -o <output-filepath> <torrent-filepath> <piece-number>", os.Args[0])
		if len(os.Args) < 6 || os.Args[2] != "-o" {
			panic(usageString)
		}

		outFilepath := os.Args[3]
		torrFilepath := os.Args[4]
		pieceIndex, err := strconv.Atoi(os.Args[5])
		panicIf(err)

		torr, infoHash, err := ParseTorrent(torrFilepath)
		panicIf(err)

		numPieces := len(torr.info.pieces)
		if pieceIndex < 0 || pieceIndex >= numPieces {
			panic(fmt.Sprintf("Torrent %s has %d pieces, so <piece-number> can be between 0 and %d", torrFilepath, numPieces, numPieces-1))
		}

		trackerResp, err := TrackerRequest(torr, infoHash, PeerId)
		panicIf(err)

		peer := trackerResp.Peers[0]
		conn, _, err := ConnectToPeer(peer, infoHash)
		defer conn.Close()
		panicIf(err)

		interestedMsg := PeerMessage{pmidInterested, []byte{}}
		err = sendPeerMessage(conn, interestedMsg)
		panicIf(err)
		fmt.Printf("Sent 'interested' msg to the peer!\n")

		peerMsg := PeerMessage{pmidChoke, []byte{}}
		for peerMsg.id != pmidUnchoke {
			fmt.Printf("Waiting for 'unchoke' msg from the peer...\n")
			peerMsg, err = readPeerMessage(conn)
			panicIf(err)
			fmt.Printf("Read peer msg with id: %d, payload: %x\n", peerMsg.id, peerMsg.payload)
		}

		torrLength := torr.info.length
		pieceLength := torr.info.pieceLength
		if pieceIndex == numPieces-1 {
			// last piece may be shorter than the others
			pieceLength = torrLength % pieceLength
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
			panicIf(err)
			fmt.Printf("- Sent 'request' msg to peer: piece: %d, blockBegin: %d, requestPayload: %x\n",
				pieceIndex, blockBegin, requestPayload)
		}

		fmt.Printf("Listening for 'piece' messages from peer:\n")
		for len(blocks) < numBlocks {
			peerMsg, err = readPeerMessage(conn)
			panicIf(err)

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

		pieceHashArray := sha1.Sum(piece)
		pieceHash := string(pieceHashArray[:])
		expectedHash := torr.info.pieces[pieceIndex]
		if pieceHash[:] != expectedHash {
			panic(fmt.Errorf("Got piece %d with hash %x which differs from expected hash %x!\n",
				pieceIndex, pieceHash, expectedHash))
		} else {
			fmt.Printf("Piece %d matches expected hash %x! :)\n", pieceIndex, expectedHash)
		}

		outFile, err := os.Create(outFilepath)
		panicIf(err)
		defer outFile.Close()
		fmt.Printf("Opened file %s to write piece %d.\n", outFilepath, pieceIndex)

		n, err := outFile.Write(piece)
		panicIf(err)
		fmt.Printf("Wrote piece %d, %d bytes.\n", pieceIndex, n)

		fmt.Printf("Enjoy! :)\n")
	default:
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}
}
