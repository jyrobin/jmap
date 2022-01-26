// Copyright (c) 2021 Jing-Ying Chen. Subject to the MIT License.

package jmap

import (
	"encoding/json"
	"fmt"
	"strings"
)

type PriMap struct {
	vals       map[string]interface{}
	filter     func(v interface{}) bool
	normalizer func(v interface{}) (interface{}, error)
}

type PriMapOption func(*PriMap)

func Normalizer(f func(v interface{}) (interface{}, error)) PriMapOption {
	return func(priMap *PriMap) {
		priMap.normalizer = f
	}
}
func Filter(f func(v interface{}) bool) PriMapOption {
	return func(priMap *PriMap) {
		priMap.filter = f
	}
}

func NewPriMapError(vals map[string]interface{}, opts ...PriMapOption) (*PriMap, error) {
	ret := &PriMap{vals, IsPrimitive, NormalizeError}
	for _, opt := range opts {
		opt(ret)
	}
	err := ret.Replace(vals)
	return ret, err
}

func (pm *PriMap) Replace(vals map[string]interface{}) error {
	var err error
	var invalids []string
	vs := make(map[string]interface{}, len(vals))
	for k, v := range vals {
		if pm.filter(v) {
			// Note: normalization does not skip value, just report errors
			vs[k], err = pm.normalizer(v)
			if err != nil {
				invalids = append(invalids, k)
			}
		}
	}

	if len(invalids) > 0 {
		return fmt.Errorf("Invalid values: %s", strings.Join(invalids, ", "))
	}
	return nil
}

func NewPriMap(vals map[string]interface{}) *PriMap {
	ret, _ := NewPriMapError(vals)
	return ret
}

func IsInt(v interface{}) bool {
	switch v.(type) {
	case int:
	case int8:
	case int16:
	case int32:
	case int64:
	case uint:
	case uint8:
	case uint16:
	case uint32:
	case uint64:
		return true
	}
	return false
}

func IsFloat(v interface{}) bool {
	switch v.(type) {
	case float32:
	case float64:
		return true
	}
	return false
}

func IsPrimitive(v interface{}) bool {
	switch v.(type) {
	case string:
	case bool:
		return true
	}
	return IsInt(v) || IsFloat(v)
}

func NormalizeError(v interface{}) (interface{}, error) {
	switch val := v.(type) {
	case string:
	case bool:
		return v, nil
	case int:
	case int8:
	case int16:
	case int32:
	case uint:
	case uint8:
	case uint16:
	case uint32:
	case uint64:
		return int64(val), nil
	case float32:
		return float64(val), nil
	}
	return v, fmt.Errorf("Not a primitive")
}

func Normalize(v interface{}) interface{} {
	ret, _ := NormalizeError(v)
	return ret
}

func (pm *PriMap) Json(opts ...string) []byte {
	var ret []byte
	switch len(opts) {
	case 0:
		ret, _ = json.Marshal(pm.vals)
	case 1:
		ret, _ = json.MarshalIndent(pm.vals, "", opts[0])
	default:
		ret, _ = json.MarshalIndent(pm.vals, opts[1], opts[0])
	}
	return ret
}

func (pm *PriMap) MarshalJSON() ([]byte, error) {
	return json.Marshal(pm.vals)
}

func (pm *PriMap) UnmarshalJSON(b []byte) error {
	var vals map[string]interface{}
	if err := json.Unmarshal(b, &vals); err != nil {
		return err
	}
	return pm.Replace(vals)
}
