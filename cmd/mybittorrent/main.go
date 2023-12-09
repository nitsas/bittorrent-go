package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"unicode"
	// bencode "github.com/jackpal/bencode-go" // Available if you need it!
)

// Examples:
// - "5:hello" -> "hello"
// - "i52e" -> 52
func decodeBencode(bencodedString string) (interface{}, error) {
	if unicode.IsDigit(rune(bencodedString[0])) {
		return decodeBencString(bencodedString)
	} else if rune(bencodedString[0]) == 'i' {
		return decodeBencInt(bencodedString)
	} else {
		return "", fmt.Errorf("Only strings and integers are supported at the moment")
	}
}

// Example:
// - "5:hello" -> "hello"
// - "10:hello12345" -> "hello12345"
func decodeBencString(bencodedString string) (string, error) {
	var firstColonIndex int

	for i := 0; i < len(bencodedString); i++ {
		if bencodedString[i] == ':' {
			firstColonIndex = i
			break
		}
	}

	lengthStr := bencodedString[:firstColonIndex]

	length, err := strconv.Atoi(lengthStr)
	if err != nil {
		return "", err
	}

	return bencodedString[firstColonIndex+1 : firstColonIndex+1+length], nil
}

// Examples:
// - "i52e" -> 52
// - "i-42e" -> -42
func decodeBencInt(bencodedString string) (int64, error) {
	lastRune := rune(bencodedString[len(bencodedString)-1])
	if lastRune != 'e' {
		return 0, fmt.Errorf("Bencoded integer not terminated properly: %q instead of 'e'", lastRune)
	}
	// Here are some extra invalid cases that Atoi can handle sanely:
	// "i-0e": -0 is invalid
	// "i03e": leading 0s are invalid, except for "i0e"

	// The maximum number of bits is unspecified. But 64-bit integers can handle very big numbers.

	return strconv.ParseInt(bencodedString[1:len(bencodedString)-1], 10, 64)
}

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	// fmt.Println("Logs from your program will appear here!")

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
