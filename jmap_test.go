// Copyright (c) 2021 Jing-Ying Chen. Subject to the MIT License.

package jmap

import (
	"reflect"
	"testing"
)

func TestFlatten(t *testing.T) {
	var tests = []struct {
		v1  interface{}
		v2  map[string]interface{}
		err bool
	}{
		{
			1, nil, true,
		},
		{
			map[string]interface{}{"a": 1, "b": 2},
			map[string]interface{}{"b": 2, "a": 1},
			false,
		},
		{
			map[string]interface{}{"a": 1, "c": map[string]interface{}{"x": 3, "y": 4}, "b": 2},
			map[string]interface{}{"b": 2, "a": 1, "c.x": 3, "c.y": 4},
			false,
		},
	}

	for _, tt := range tests {
		ret, err := FlattenMap(tt.v1, nil, nil)
		if tt.err {
			if err == nil {
				t.Fatalf("%+v should fail to flatten", tt.v1)
			}
		} else {
			if err != nil {
				t.Fatalf("%+v got error %v", tt.v1, err)
			} else if !reflect.DeepEqual(ret, tt.v2) {
				t.Fatalf("%+v and %+v not equal", ret, tt.v2)
			}
		}

		if !tt.err {
			ret, err = UnflattenMap(tt.v2, nil)
			if err != nil {
				t.Fatalf("%+v should fail to unflatten", tt.v2)
			} else if !reflect.DeepEqual(ret, tt.v1) {
				t.Fatalf("%+v and %+v not equal", ret, tt.v1)
			}
		}
	}
}
