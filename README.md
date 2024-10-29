# Bittorrent Go

Building a bittorent client with Go (v1.19) step by step, following the [instructions on codecrafters.io](https://app.codecrafters.io/courses/bittorrent/overview).

In this project, we build a BitTorrent client that's capable of parsing a
.torrent file and downloading a file from a peer. Along the way, we learn
about how torrent files are structured, HTTP trackers, BitTorrent’s Peer
Protocol, pipelining and more.

## How to try this out

### Decode bencoded strings

```sh
./your_bittorrent.sh decode 5:hello
```

Expected output:
```
"hello"
```

### Decode bencoded integers

```sh
./your_bittorrent.sh decode i52e
```

Expected output:
```
52
```

### Decode bencoded lists

```sh
./your_bittorrent.sh decode l5:helloi52ee
```

Expected output
```
[“hello”,52]
```

### Decode bencoded dictionaries

```sh
./your_bittorrent.sh decode d3:foo3:bar5:helloi52ee
```

Expected output:
```
{"foo":"bar","hello":52}
```

### Parse torrent file, calculate info hash, and get piece hashes

```sh
./your_bittorrent.sh info sample.torrent
```

Expected output:
```
Tracker URL: http://bittorrent-test-tracker.codecrafters.io/announce
Length: 92063
Info Hash: d69f91e6b2ae4c542468d1073a71d4ea13879a7f
Piece Length: 32768
Piece Hashes:
e876f67a2a8886e8f36b136726c30fa29703022d
6e2275e604a0766656736e81ff10b55204ad8d35
f00d937a0213df1982bc8d097227ad9e909acc17
```

### Discover peers

```sh
./your_bittorrent.sh peers sample.torrent
```

Expected output:
(exact values may be different)
```
167.71.141.80:51470
167.71.141.82:51482
134.209.186.165:51417
```

### Peer handshake

```sh
./your_bittorrent.sh handshake sample.torrent 178.62.82.89:51470
```

Expected output:
(exact value will be different as it is randomly generated)
```
Peer ID: 0102030405060708090a0b0c0d0e0f1011121314
```

### Download a piece

```sh
./your_bittorrent.sh download_piece -o /tmp/test-piece-0 sample.torrent 0
```

Expected output:
```
Piece 0 downloaded to /tmp/test-piece-0.
```

### Download a whole torrent

```sh
./your_bittorrent.sh download -o /tmp/sample.txt sample.torrent
```

Expected output:
```
Downloaded sample.torrent to /tmp/sample.txt.
```

### Parse magnet link

```sh
./your_bittorrent.sh magnet_parse 'magnet:?xt=urn:btih:ad42ce8109f54c99613ce38f9b4d87e70f24a165&dn=magnet1.gif&tr=http%3A%2F%2Fbittorrent-test-tracker.codecrafters.io%2Fannounce'
```

Expected output:
```
Tracker URL: http://bittorrent-test-tracker.codecrafters.io/announce
Info Hash: d69f91e6b2ae4c542468d1073a71d4ea13879a7f
```

### Announce extension support

```sh
./your_bittorrent.sh magnet_handshake 'magnet:?xt=urn:btih:ad42ce8109f54c99613ce38f9b4d87e70f24a165&dn=magnet1.gif&tr=http%3A%2F%2Fbittorrent-test-tracker.codecrafters.io%2Fannounce'
```

Expected output:
```
Peer ID: 0102030405060708090a0b0c0d0e0f1011121314
```
