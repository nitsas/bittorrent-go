package main

import (
	"fmt"
	"sort"
)

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
