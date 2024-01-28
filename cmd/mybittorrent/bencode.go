package main

import "fmt"

func bencode(obj interface{}) (result string) {
	switch t := obj.(type) {
	case int:
		result = fmt.Sprintf("i%de", t)
	case string:
		result = fmt.Sprintf("%d:%s", len(t), t)
	case map[string]interface{}:
		result = "d"
		for key, val := range t {
			result += bencode(key) + bencode(val)
		}
		result += "e"
	case []interface{}:
		result = "l"
		for _, elem := range t {
			result += bencode(elem)
		}
		result += "e"
	}

	return result
}
