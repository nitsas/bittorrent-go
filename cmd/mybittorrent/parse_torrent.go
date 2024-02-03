package main

import (
	"crypto/sha1"
	"os"
)

func ParseTorrent(filename string) (*torrent, string, error) {
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
	piecesString := infoDict["pieces"].(string)
	pieces := make([]string, len(piecesString)/20)
	for i := 0; i < len(piecesString); i += 20 {
		pieces[i/20] = piecesString[i : i+20]
	}

	t := torrent{
		announce: torrDict["announce"].(string),
		info: torrentInfo{
			length:      infoDict["length"].(int),
			name:        infoDict["name"].(string),
			pieceLength: infoDict["piece length"].(int),
			pieces:      pieces,
		},
	}

	return &t, infoHash, nil
}
