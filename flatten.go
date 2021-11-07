package jmap

import (
	"fmt"
	"reflect"
	"strings"
)

const MaxDepth = 15 // why go deeper

// Note: no cloning whatsoever, only structural rewriting at topper levels, so use with care
// - all nasty side-effecting and cycling
// - if supplying flat, it can be partially modified
func Flatten(v interface{}, depth int, sep string, flat map[string]interface{}) (map[string]interface{}, error) {
	if flat == nil {
		flat = make(map[string]interface{})
	}
	if sep == "" {
		sep = "."
	}
	if depth <= 0 {
		depth = MaxDepth
	}

	var err error
	if smap, ok := v.(map[string]interface{}); ok { // well let me optimize pre-maturally
		err = flattenMap(smap, "", depth, sep, flat)
	} else if isStrMap(v) {
		err = flatten(v, "", depth, sep, flat)
	} else {
		err = fmt.Errorf("Not a string-keyed map")
	}
	return flat, err
}

func isStrMap(v interface{}) bool {
	t := reflect.TypeOf(v)
	return t != nil && t.Kind() == reflect.Map && t.Key().Kind() == reflect.String
}

func flattenMap(smap map[string]interface{}, prefix string, depth int, sep string, flat map[string]interface{}) error {
	if depth <= 0 {
		flat[prefix] = smap // ok since prefix cannot be ""
		return nil
	}

	if depth <= 0 || len(smap) == 0 {
		flat[prefix] = smap
		return nil
	}

	var err error
	for k, v := range smap {
		newKey := k
		if prefix != "" {
			newKey = prefix + sep + newKey
		}
		if sm, ok := v.(map[string]interface{}); ok {
			err = flattenMap(sm, newKey, depth-1, sep, flat)
		} else {
			err = flatten(v, newKey, depth-1, sep, flat)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func flatten(v interface{}, prefix string, depth int, sep string, flat map[string]interface{}) error {
	if depth <= 0 || !isStrMap(v) {
		flat[prefix] = v // ok since prefix cannot be ""
		return nil
	}

	val := reflect.ValueOf(v)
	keys := val.MapKeys()
	if len(keys) == 0 {
		flat[prefix] = v
		return nil
	}

	for _, key := range keys {
		if ch := val.MapIndex(key); ch.IsValid() {
			newKey := key.Interface().(string)
			if prefix != "" {
				newKey = prefix + sep + newKey
			}

			var err error
			chv := ch.Interface()
			if sm, ok := chv.(map[string]interface{}); ok {
				err = flattenMap(sm, newKey, depth-1, sep, flat)
			} else {
				err = flatten(chv, newKey, depth-1, sep, flat)
			}
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// Assuming unflatten what's flattened above, meaning only map[string]interface{} counts
// as children
func Unflatten(flat map[string]interface{}, sep string) (map[string]interface{}, error) {
	if len(flat) == 0 {
		return make(map[string]interface{}), nil
	}

	if sep == "" {
		sep = "."
	}
	return unflatten(flat, sep)
}

func unflatten(flat map[string]interface{}, sep string) (map[string]interface{}, error) {
	ret := make(map[string]interface{})

	for k, v := range flat {
		keys := strings.Split(k, sep)
		if n := len(keys); n > 0 {
			smap := ret
			for i := 0; i < n-1; i++ {
				key := keys[i]
				if ch, ok := smap[key]; ok {
					if sm, ok := ch.(map[string]interface{}); ok {
						smap = sm
					} else {
						return nil, fmt.Errorf("path %s has invalid intermediate at %d key %s", k, i, key)
					}

				} else {
					sm := make(map[string]interface{})
					smap[key] = sm
					smap = sm
				}
			}
			smap[keys[n-1]] = v
		}
	}
	return ret, nil
}
