package main

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"net"
	"os"
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
		conn, bitfield, err := ConnectToPeer(peer, infoHash)
		defer conn.Close()
		panicIf(err)

		if !bitfield.HasPiece(pieceIndex) {
			panic(fmt.Sprintf("Peer %#v does NOT have piece %d\n", peer, pieceIndex))
		}

		torrLength := torr.info.length
		pieceLength := torr.info.pieceLength
		if pieceIndex == numPieces-1 {
			// last piece may be shorter than the others
			pieceLength = torrLength % pieceLength
		}

		piece, err := DownloadPiece(conn, pieceIndex, pieceLength)
		panicIf(err)

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
