package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
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

		handshakeResp := make([]byte, 1000)
		_, err = conn.Read(handshakeResp)
		if err != nil && err != io.EOF {
			panic(err)
		}
		fmt.Printf("Peer ID: %x\n", handshakeResp[48:68])
	default:
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}
}
