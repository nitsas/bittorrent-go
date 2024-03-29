package main

import (
	"fmt"
	"sort"
	"strconv"
	"unicode"
)

// --- Encoding ---

func Bencode(obj interface{}) (result string) {
	switch tobj := obj.(type) {
	case int:
		result = fmt.Sprintf("i%de", tobj)
	case string:
		result = fmt.Sprintf("%d:%s", len(tobj), tobj)
	case map[string]interface{}:
		result = "d"
		mapKeys := make([]string, 0, len(tobj))
		for key := range tobj {
			mapKeys = append(mapKeys, key)
		}
		sort.Strings(mapKeys)
		for _, key := range mapKeys {
			result += Bencode(key) + Bencode(tobj[key])
		}
		result += "e"
	case []interface{}:
		result = "l"
		for _, elem := range tobj {
			result += Bencode(elem)
		}
		result += "e"
	default:
		err := fmt.Errorf("Unrecognized type (%T) of %#v\n", tobj, tobj)
		panic(err)
	}

	return result
}

// --- Decoding ---

// Examples:
// - "0:" -> ""
// - "5:hello" -> "hello"
// - "i52e" -> 52
// - "l5:helloi52ee" -> ["hello", 52]
func DecodeBencode(bencData string) (interface{}, error) {
	token, _, error := decodeNextBencToken(bencData[0:])
	return token, error
}

func decodeNextBencToken(bencData string) (interface{}, uint, error) {
	if unicode.IsDigit(rune(bencData[0])) {
		return decodeBencString(bencData)
	} else if rune(bencData[0]) == 'i' {
		return decodeBencInt(bencData)
	} else if rune(bencData[0]) == 'l' {
		return decodeBencList(bencData)
	} else if rune(bencData[0]) == 'd' {
		return decodeBencDict(bencData)
	} else {
		return "", 0, fmt.Errorf("Unrecognized type")
	}
}

// Examples:
// - "de" -> {}
// - "d3:foo3:bar5:helloi52ee" -> {"foo": "bar", "hello": 52}
func decodeBencDict(bencData string) (map[string]interface{}, uint, error) {
	result := make(map[string]interface{})
	i := 1
	for bencData[i] != 'e' && i < len(bencData) {
		key, skipIndex, err := decodeBencString(bencData[i:])
		if err != nil {
			return map[string]interface{}{}, 0, err
		}
		i += int(skipIndex)

		val, skipIndex, err := decodeNextBencToken(bencData[i:])
		if err != nil {
			return map[string]interface{}{}, 0, err
		}

		i += int(skipIndex)
		result[key] = val
	}

	if i >= len(bencData) {
		err := fmt.Errorf("Bencoded dict '%s' missing trailing 'e'", bencData)
		return map[string]interface{}{}, 0, err
	}

	return result, uint(i + 1), nil
}

// Examples:
// - "le" -> []
// - "l5:helloi52ee" -> ["hello", 52]
// - "lli4eei5ee" -> [[4], 5]
func decodeBencList(bencData string) ([]interface{}, uint, error) {
	result := make([]interface{}, 0)
	i := uint(1)
	for bencData[i] != 'e' {
		token, nextIndex, err := decodeNextBencToken(bencData[i:])
		if err != nil {
			return []interface{}{}, 0, err
		}
		i += nextIndex
		result = append(result, token)
	}
	if bencData[i] != 'e' {
		err := fmt.Errorf("Bencoded list '%s' missing trailing 'e'", bencData)
		return []interface{}{}, 0, err
	}

	return result, i + 1, nil
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
