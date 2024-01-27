package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"unicode"
)

// Examples:
// - "0:" -> ""
// - "5:hello" -> "hello"
// - "i52e" -> 52
func decodeBencode(bencData string) (interface{}, error) {
	token, _, error := decodeNextBencToken(bencData[0:])
	return token, error
}

func decodeNextBencToken(bencData string) (interface{}, uint, error) {
	if unicode.IsDigit(rune(bencData[0])) {
		return decodeBencString(bencData)
	} else if rune(bencData[0]) == 'i' {
		return decodeBencInt(bencData)
	} else {
		return "", 0, fmt.Errorf("Unrecognized type")
	}
}

// Example:
// - "0:" -> ""
// - "5:hello" -> "hello"
// - "10:hello12345" -> "hello12345"
func decodeBencString(bencData string) (string, uint, error) {
	var colonIndex int

	for i, c := range bencData {
		if c == ':' {
			colonIndex = i
			break
		}
	}

	lengthStr := bencData[:colonIndex]
	length, err := strconv.Atoi(lengthStr)
	if err != nil {
		return "", 0, err
	}
	nextIndex := uint(colonIndex + 1 + length)

	return bencData[colonIndex+1 : nextIndex], nextIndex, nil
}

// Examples:
// - "i52e" -> 52
// - "i-42e" -> -42
func decodeBencInt(bencData string) (int, uint, error) {
	nextIndex := len(bencData) + 1
	for i, c := range bencData {
		if c == 'e' {
			nextIndex = i + 1
			break
		}
	}
	if nextIndex > len(bencData) {
		err := fmt.Errorf("Bencoded integer '%s' missing trailing 'e'", bencData)
		return 0, 0, err
	}

	// Here are some extra invalid cases that Atoi can handle sanely:
	// - "i-0e": -0 is invalid
	// - "i03e": leading 0s are invalid, except for "i0e"
	//
	// The maximum number of bits is unspecified. But 64-bit integers can handle very big numbers.

	num, err := strconv.Atoi(bencData[1 : nextIndex-1])
	if err != nil {
		return 0, 0, err
	} else {
		return num, uint(nextIndex), nil
	}
}

func main() {
	command := os.Args[1]

	if command == "decode" {
		bencodedValue := os.Args[2]

		decoded, err := decodeBencode(bencodedValue)
		if err != nil {
			panic(err)
		}

		jsonOutput, _ := json.Marshal(decoded)
		fmt.Println(string(jsonOutput))
	} else {
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}
}
