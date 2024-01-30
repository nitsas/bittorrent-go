package main

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"os"
)

func parseTorrent(filename string) (map[string]interface{}, error) {
	data, err := os.ReadFile(filename)
	panicIf(err)
	torrContent := string(data)
	torrDict, err := DecodeBencode(torrContent)

	return torrDict.(map[string]interface{}), err
}

func cmdInfo(filename string) (string, int, string, error) {
	torrDict, err := parseTorrent(filename)
	if err != nil {
		return "", 0, "", err
	}

	url := torrDict["announce"].(string)
	info := torrDict["info"].(map[string]interface{})
	length := info["length"].(int)
	infoHash := sha1.Sum([]byte(Bencode(info)))

	return url, length, string(infoHash[:]), nil
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
		url, length, infoHash, err := cmdInfo(os.Args[2])
		panicIf(err)
		fmt.Printf("Tracker URL: %s\nLength: %d\nInfo Hash: %x\n", url, length, infoHash)
	} else {
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}
}
