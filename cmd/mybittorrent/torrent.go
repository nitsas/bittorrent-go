package main

type torrent struct {
	announce string
	info     torrentInfo
}

type torrentInfo struct {
	length      int
	name        string
	pieceLength int
	pieces      []string
}
