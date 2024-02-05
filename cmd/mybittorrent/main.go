package main

import (
	"encoding/binary"
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
		resp, err := TorrentRequest(os.Args[2], PeerId)
		panicIf(err)
		defer resp.Body.Close()

		respBodyBytes, err := io.ReadAll(resp.Body)
		panicIf(err)

		respString := string(respBodyBytes)
		respDecoded, err := DecodeBencode(respString)
		panicIf(err)
		respMap, ok := respDecoded.(map[string]interface{})
		if !ok {
			panic("Unexpected type in decoded response.")
		}
		// interval := respMap["interval"].(int)
		// if !ok {
		// 	panic("Unexpected type in key 'interval'.")
		// }
		peersStr := respMap["peers"].(string)
		if !ok {
			panic("Unexpected type in key 'peers'.")
		}
		peersBytes := []byte(peersStr)
		for i := 0; i < len(peersBytes); i += 6 {
			ip := net.IP(peersBytes[i : i+4])
			port := binary.BigEndian.Uint16(peersBytes[i+4 : i+6])
			fmt.Printf("%s:%d\n", ip, port)
		}
	case "handshake":
		// ipPort := strings.Split(os.Args[2], ":")
		// ip := ipPort[0]
		// port := ipPort[1]
		// fmt.Printf("%s:%s\n", ip, port)

		torrName := os.Args[2]
		peerIpPort := os.Args[3]

		_, infoHash, err := ParseTorrent(torrName)
		panicIf(err)

		handshake := make([]byte, 0, 68)
		handshake = append(handshake, byte(19))
		handshake = append(handshake, []byte("BitTorrent protocol")...)
		handshake = append(handshake, make([]byte, 8)...)
		handshake = append(handshake, infoHash...)
		handshake = append(handshake, []byte(PeerId)...)

		conn, err := net.Dial("tcp", peerIpPort)
		panicIf(err)
		defer conn.Close()

		_, err = conn.Write(handshake)
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
