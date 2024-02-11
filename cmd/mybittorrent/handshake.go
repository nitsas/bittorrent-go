package main

import (
	"fmt"
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
