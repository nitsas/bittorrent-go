package main

type Bitfield []byte

func (bitfield Bitfield) HasPiece(pieceIndex int) bool {
	targetByte := uint8(bitfield[pieceIndex/8])
	indexAsByteBitfield := uint8(1 << (7 - (pieceIndex % 8)))
	return (targetByte & indexAsByteBitfield) > 0
}
