package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
)

type TrackerResponse struct {
	Interval int
	Peers    []Peer
}

type Peer struct {
	Ip   net.IP
	Port uint
}

func TrackerRequest(trackerURL string, infoHash []byte, peerId string) (*TrackerResponse, error) {
	targetUrl, err := url.Parse(trackerURL)
	if err != nil {
		return nil, err
	}

	params := url.Values{}
	params.Add("info_hash", string(infoHash))
	params.Add("peer_id", peerId)
	params.Add("port", "6881")
	params.Add("uploaded", "0")
	params.Add("downloaded", "0")
	params.Add("left", fmt.Sprintf("%d", 999))
	params.Add("compact", "1")

	targetUrl.RawQuery = params.Encode()

	resp, err := http.Get(targetUrl.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	trackerResp, err := ParseTrackerResponse(resp)

	return trackerResp, err
}

func ParseTrackerResponse(resp *http.Response) (*TrackerResponse, error) {
	respBodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	respString := string(respBodyBytes)
	respDecoded, err := DecodeBencode(respString)
	if err != nil {
		return nil, err
	}

	respMap, ok := respDecoded.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("Unexpected type in decoded tracker response. Expected map[string]interface{}.")
	}
	interval, ok := respMap["interval"].(int)
	if !ok {
		return nil, fmt.Errorf("Unexpected type in key 'interval'. Expected int.")
	}
	peersStr, ok := respMap["peers"].(string)
	if !ok {
		return nil, fmt.Errorf("Unexpected type in key 'peers'. Expected string.")
	}

	peersBytes := []byte(peersStr)
	peers := make([]Peer, 0, len(peersBytes)/6)
	for i := 0; i < len(peersBytes); i += 6 {
		ip := net.IP(peersBytes[i : i+4])
		port := uint(binary.BigEndian.Uint16(peersBytes[i+4 : i+6]))
		peers = append(peers, Peer{ip, port})
	}

	return &TrackerResponse{interval, peers}, nil
}
