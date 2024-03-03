package main

import (
	"fmt"
	"io"
	"net"
)

func writeHandshake(conn net.Conn, infoHash []byte) error {
	hs := make([]byte, 0, 68)
	hs = append(hs, byte(19))
	hs = append(hs, []byte("BitTorrent protocol")...)
	hs = append(hs, make([]byte, 8)...)
	hs = append(hs, infoHash...)
	hs = append(hs, []byte(PeerId)...)

	bytesWritten, err := conn.Write(hs)
	if err == nil && bytesWritten < len(hs) {
		err = fmt.Errorf("handshake: wrote %d bytes instead of %d", bytesWritten, len(hs))
	}

	return err
}

func readHandshake(conn net.Conn) ([]byte, error) {
	handshakeResp := make([]byte, 68)
	n, err := io.ReadFull(conn, handshakeResp)
	if err != nil && err != io.EOF {
		return nil, err
	}
	if n < 68 {
		err = fmt.Errorf("Expected handshake, but read only %d bytes!", n)
		return nil, err
	}
	if handshakeResp[0] != byte(19) || string(handshakeResp[1:20]) != "BitTorrent protocol" {
		err = fmt.Errorf("Malformed handshake response!")
		return nil, err
	}
	peerId := handshakeResp[48:68]

	return peerId, nil
}
