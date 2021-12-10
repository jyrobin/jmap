// Copyright (c) 2021 Jing-Ying Chen
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package jmap

import (
	"fmt"
	"reflect"
	"strings"
)

// Note FlattenValue/UnflattenMap servce as the baseline implementation for Jmap reference

type Config struct {
	Depth     int
	Separator string
	IsMap     IsMap
	Prefix    string
}

// FlattenMap create a flat map[string]interface{} from a given input string-keyed map v.
// It uses IsMap to decide whether to recursive into a map's field value.
// Note: no cloning whatsoever, only structural rewriting at topper levels, so use with care
// - potential nasty side-effecting and cycling
// - the providing map can be partially modified
func FlattenMap(v interface{}, cfg *Config, ret map[string]interface{}) (map[string]interface{}, error) {
	if cfg == nil {
		cfg = &Config{}
	}
	sep := cfg.Separator
	if sep == "" {
		sep = "."
	}
	depth := cfg.Depth
	if depth <= 0 { // does not accept 0
		depth = MaxDepth
	}
	isMap := cfg.IsMap
	if isMap == nil {
		isMap = IsMinMap // default minimal reflection
	}
	prefix := strings.TrimSpace(cfg.Prefix)

	if ret == nil {
		ret = make(map[string]interface{})
	}

	if !isMap(v) {
		return ret, fmt.Errorf("Not a string-keyed map")
	}

	return ret, flattenMap(v, prefix, depth, sep, isMap, ret)
}

func flattenMap(v interface{}, prefix string, depth int, sep string, isMap IsMap, ret map[string]interface{}) error {
	if depth <= 0 || !isMap(v) {
		ret[prefix] = v // ok since prefix cannot be ""
		return nil
	}

	val := reflect.ValueOf(v)
	keys := val.MapKeys() // should not panic due to isMap
	if len(keys) == 0 {
		ret[prefix] = v
		return nil
	}

	for _, key := range keys {
		if ch := val.MapIndex(key); ch.IsValid() {
			newKey := key.Interface().(string) // should not panic due to isMap
			if prefix != "" {
				newKey = prefix + sep + newKey
			}
			if err := flattenMap(ch.Interface(), newKey, depth-1, sep, isMap, ret); err != nil {
				return err
			}
		}
	}
	return nil
}

// UnflattenMap turns a flattened map flat created by FlatternMap into a nested tree
// where all "internal" nodes of flat are map[string]interface{}
func UnflattenMap(flat map[string]interface{}, cfg *Config) (map[string]interface{}, error) {
	ret := make(map[string]interface{})
	if len(flat) == 0 {
		return ret, nil
	}

	if cfg == nil {
		cfg = &Config{}
	}
	sep := cfg.Separator
	if sep == "" {
		sep = "."
	}
	prefix := strings.TrimSpace(cfg.Prefix)
	plen := len(prefix)

	for k, v := range flat {
		if plen > 0 {
			if len(k) <= plen || k[0:plen] != prefix {
				continue
			}
			k = k[plen:]
		}

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
