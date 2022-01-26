// Copyright (c) 2021 Jing-Ying Chen. Subject to the MIT License.

package jmap

import (
	"fmt"
	"reflect"
)

const MaxDepth = 15 // why go deeper

// Note: must ensure string-keyed map
type IsMap func(v interface{}) bool

func IsJsonMap(v interface{}) bool {
	_, ok := v.(JsonMap)
	return ok
}

func IsMinMap(v interface{}) bool {
	_, ok := v.(map[string]interface{})
	return ok
}

func IsGeneralMap(v interface{}) bool {
	t := reflect.TypeOf(v)
	return t != nil && t.Kind() == reflect.Map && t.Key().Kind() == reflect.String
}

func IsStringKeyMap(v interface{}) bool {
	return IsMinMap(v) || IsGeneralMap(v)
}

// JsonMap serves as a marker in a jmap structure
type JsonMap map[string]interface{}

type Mapper interface {
	IsMap(v interface{}) bool
	Unpack(v interface{}) ([]string, []interface{})
}

type Visitor interface {
	Visit(key string, host JsonMap, path []string)
	Visited(key string, host JsonMap, path []string)
}

type MinMapper struct{}

func (m MinMapper) IsMap(v interface{}) bool {
	return IsMinMap(v)
}
func (m MinMapper) Unpack(v interface{}) ([]string, []interface{}) {
	mapv := v.(map[string]interface{}) // should not panic if IsMap is correct
	klen := len(mapv)
	keys := make([]string, 0, klen)
	vals := make([]interface{}, 0, klen)
	for key, val := range mapv {
		keys = append(keys, key)
		vals = append(vals, val)
	}
	return keys, vals
}

type ReflectMapper struct{}

func (m ReflectMapper) IsMap(v interface{}) bool {
	return IsStringKeyMap(v)
}
func (m ReflectMapper) Unpack(v interface{}) ([]string, []interface{}) {
	rval := reflect.ValueOf(v)
	rkeys := rval.MapKeys() // should not panic if IsMap is correct
	klen := len(rkeys)

	keys := make([]string, 0, klen)
	vals := make([]interface{}, 0, klen)
	for i, rkey := range rkeys {
		if ch := rval.MapIndex(rkey); ch.IsValid() {
			keys[i] = rkey.Interface().(string) // should not panic if IsMap is correct
			vals[i] = ch.Interface()
		}
	}
	return keys, vals
}

func BuildJsonMap(v interface{}, depth int, mapper Mapper) (JsonMap, error) {
	if mapper == nil {
		mapper = MinMapper{}
	}

	if depth <= 0 {
		depth = MaxDepth
	}

	if !mapper.IsMap(v) {
		return nil, fmt.Errorf("Not a valid map")
	}

	ret := JsonMap{}
	return ret, buildJsonMap(v, depth, mapper, ret)
}

// v and ret at the same level; v passed mapper.IsMap already
func buildJsonMap(v interface{}, depth int, mapper Mapper, ret JsonMap) error {
	if depth <= 0 {
		return nil
	}

	keys, vals := mapper.Unpack(v)
	for i, key := range keys {
		val := vals[i]
		if mapper.IsMap(val) {
			ch := JsonMap{}
			err := buildJsonMap(val, depth-1, mapper, ch)
			ret[key] = ch // add it even when err != nil, to see partial results

			if err != nil {
				return err
			}
		} else {
			ret[key] = val
		}
	}
	return nil
}

// Traverse the JsonMap tree
func Traverse(host JsonMap, visitor Visitor) {
	traverse("", host, []string{}, visitor)
}
func traverse(key string, host JsonMap, path []string, visitor Visitor) {
	visitor.Visit(key, host, path)
	for k, v := range host {
		if jm, ok := v.(JsonMap); ok { // JsonMap as the marker
			traverse(k, jm, append(path, k), visitor)
		}
	}
	visitor.Visited(key, host, path)
}
