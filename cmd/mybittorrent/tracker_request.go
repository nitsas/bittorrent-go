package main

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
)

func TrackerRequest(filename string, peerId string) (*http.Response, error) {
	t, infoHash, err := ParseTorrent(os.Args[2])
	if err != nil {
		return nil, err
	}

	targetUrl, err := url.Parse(t.announce)
	if err != nil {
		return nil, err
	}

	params := url.Values{}
	params.Add("info_hash", string(infoHash))
	params.Add("peer_id", peerId)
	params.Add("port", "6881")
	params.Add("uploaded", "0")
	params.Add("downloaded", "0")
	params.Add("left", fmt.Sprintf("%d", t.info.length))
	params.Add("compact", "1")

	targetUrl.RawQuery = params.Encode()

	resp, err := http.Get(targetUrl.String())
	if err != nil {
		return nil, err
	}

	return resp, err
}
