package main

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"os"
)

type torrent struct {
	announce string
	info     torrentInfo
}

type torrentInfo struct {
	length      int
	name        string
}

func parseTorrent(filename string) (*torrent, string, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, "", err
	}
	torrContent := string(data)

	torrDecoded, err := DecodeBencode(torrContent)
	torrDict := torrDecoded.(map[string]interface{})
	infoDict := torrDict["info"].(map[string]interface{})
	infoHashBytes := sha1.Sum([]byte(Bencode(infoDict)))
	infoHash := string(infoHashBytes[:])

	t := torrent{
		announce: torrDict["announce"].(string),
		info: torrentInfo{
			length:      infoDict["length"].(int),
			name:        infoDict["name"].(string),
		},
	}

	return &t, infoHash, nil
}

func panicIf(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	command := os.Args[1]

	if command == "decode" {
		bencodedValue := os.Args[2]

		decoded, err := DecodeBencode(bencodedValue)
		panicIf(err)

		jsonOutput, _ := json.Marshal(decoded)
		fmt.Println(string(jsonOutput))
	} else if command == "info" {
		t, infoHash, err := parseTorrent(os.Args[2])
		panicIf(err)
		fmt.Printf("Tracker URL: %s\n", t.announce)
		fmt.Printf("Length: %d\n", t.info.length)
		fmt.Printf("Info Hash: %x\n", infoHash)
	} else {
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}
}
