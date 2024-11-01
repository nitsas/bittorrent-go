package main

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
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
		trackerResp, err := TrackerRequest(torr.announce, infoHash, PeerId)
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

		err = writeHandshake(conn, infoHash, false)
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

		trackerResp, err := TrackerRequest(torr.announce, infoHash, PeerId)
		panicIf(err)

		peer := trackerResp.Peers[0]
		peerConn, bitfield, err := ConnectToPeer(peer, infoHash, false)
		defer peerConn.Close()
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

		piece, err := DownloadPiece(&peerConn, pieceIndex, pieceLength)
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

		_, err = outFile.Write(piece)
		panicIf(err)
		fmt.Printf("Piece %d downloaded to %s.\n", pieceIndex, outFilepath)
	case "download":
		usageString := fmt.Sprintf("Usage: %s download -o <output-filepath> <torrent-filepath>", os.Args[0])
		if len(os.Args) < 5 || os.Args[2] != "-o" {
			panic(usageString)
		}

		outFilepath := os.Args[3]
		torrFilepath := os.Args[4]

		torr, infoHash, err := ParseTorrent(torrFilepath)
		panicIf(err)

		trackerResp, err := TrackerRequest(torr.announce, infoHash, PeerId)
		panicIf(err)

		peer := trackerResp.Peers[0]
		peerConn, bitfield, err := ConnectToPeer(peer, infoHash, false)
		defer peerConn.Close()
		panicIf(err)

		numPieces := len(torr.info.pieces)
		pieces := make([]Piece, 0, numPieces)
		for pieceIndex := 0; pieceIndex < numPieces; pieceIndex++ {
			if !bitfield.HasPiece(pieceIndex) {
				panic(fmt.Sprintf("Peer %#v does NOT have piece %d\n", peer, pieceIndex))
			}

			torrLength := torr.info.length
			pieceLength := torr.info.pieceLength
			if pieceIndex == numPieces-1 {
				// last piece may be shorter than the others
				pieceLength = torrLength % pieceLength
			}

			piece, err := DownloadPiece(&peerConn, pieceIndex, pieceLength)
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

			pieces = append(pieces, piece)
		}

		outFile, err := os.Create(outFilepath)
		panicIf(err)
		defer outFile.Close()
		fmt.Printf("Opened file %s to write torrent.\n", outFilepath)

		for pieceIndex, piece := range pieces {
			n, err := outFile.Write(piece)
			panicIf(err)
			fmt.Printf("- Wrote piece %d, %d bytes.\n", pieceIndex, n)
		}

		fmt.Printf("Downloaded %s to %s.\n", torrFilepath, outFilepath)
	case "magnet_parse":
		usageString := fmt.Sprintf("Usage: %s magnet_parse <magnet-uri>", os.Args[0])
		if len(os.Args) < 3 {
			panic(usageString)
		}

		magnetURI := os.Args[2]
		u, err := url.Parse(magnetURI)
		panicIf(err)
		if u.Scheme != "magnet" {
			panic(fmt.Errorf("Error: Expected URI scheme to be 'magnet'. Got %s\n", u.Scheme))
		}

		q, err := url.ParseQuery(u.RawQuery)
		panicIf(err)

		xt := q["xt"]
		if len(xt) < 1 {
			panic("Expected param xt to have at least one value")
		}

		xtVals := strings.Split(xt[0], ":")
		if len(xtVals) < 3 || xtVals[0] != "urn" || xtVals[1] != "btih" {
			panic(fmt.Errorf("Expected param xt to start with 'urn:btih:' - got %s\n", xt[0]))
		}

		tr := q["tr"]
		if len(tr) < 1 {
			panic("Expected param tr to have at least one value")
		}

		infoHash := xtVals[2]
		trackerURL := tr[0]

		fmt.Printf("Tracker URL: %s\nInfo Hash: %s\n", trackerURL, infoHash)
	case "magnet_handshake":
		usageString := fmt.Sprintf("Usage: %s magnet_handshake <magnet-uri>", os.Args[0])
		if len(os.Args) < 3 {
			panic(usageString)
		}

		magnetURI := os.Args[2]
		u, err := url.Parse(magnetURI)
		panicIf(err)
		if u.Scheme != "magnet" {
			panic(fmt.Errorf("Error: Expected URI scheme to be 'magnet'. Got %s\n", u.Scheme))
		}

		q, err := url.ParseQuery(u.RawQuery)
		panicIf(err)

		xt := q["xt"]
		if len(xt) < 1 {
			panic("Expected param xt to have at least one value")
		}

		xtVals := strings.Split(xt[0], ":")
		if len(xtVals) < 3 || xtVals[0] != "urn" || xtVals[1] != "btih" {
			panic(fmt.Errorf("Expected param xt to start with 'urn:btih:' - got %s\n", xt[0]))
		}

		tr := q["tr"]
		if len(tr) < 1 {
			panic("Expected param tr to have at least one value")
		}

		infoHash := xtVals[2]
		trackerURL := tr[0]

		infoHashDecoded, err := hex.DecodeString(infoHash)
		panicIf(err)
		trackerResp, err := TrackerRequest(trackerURL, infoHashDecoded, PeerId)
		panicIf(err)

		if len(trackerResp.Peers) < 1 {
			panic("Expected at least one peer in tracker response")
		}
		peer := trackerResp.Peers[0]

		// for _, peer := range trackerResp.Peers {
		// 	fmt.Printf("%s:%d\n", peer.Ip, peer.Port)
		// }

		peerIpPort := fmt.Sprintf("%s:%d", peer.Ip, peer.Port)
		conn, err := net.Dial("tcp", peerIpPort)
		panicIf(err)
		defer conn.Close()

		peerId, err := handshake(conn, infoHashDecoded, true)
		panicIf(err)

		fmt.Printf("Peer ID: %s\n", hex.EncodeToString(peerId))
	default:
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}
}
