// Derived from code Copyright 2009 The Go Authors. All rights reserved.

// Deep equality test via reflection

package deepequal

import (
	"reflect"
)

// During deepValueEqual, must keep track of checks that are
// in progress.  The comparison algorithm assumes that all
// checks in progress are true when it reencounters them.
// Visited comparisons are stored in a map indexed by visit.
type visit struct {
	a1  uintptr
	a2  uintptr
	typ reflect.Type
}

// Tests for deep equality using reflected types. The map argument tracks
// comparisons that have already been seen, which allows short circuiting on
// recursive types.
func deepValueEqual(v1, v2 reflect.Value, visited map[visit]bool, depth int) (bool, string) {
	if !v1.IsValid() || !v2.IsValid() {
		return v1.IsValid() == v2.IsValid(), "invalid values are not equal"
	}
	if v1.Type() != v2.Type() {
		return false, "values are of differing types"
	}

	// if depth > 10 { panic("deepValueEqual") }	// for debugging
	hard := func(k reflect.Kind) bool {
		switch k {
		case reflect.Array, reflect.Map, reflect.Slice, reflect.Struct:
			return true
		}
		return false
	}

	if v1.CanAddr() && v2.CanAddr() && hard(v1.Kind()) {
		addr1 := v1.UnsafeAddr()
		addr2 := v2.UnsafeAddr()
		if addr1 > addr2 {
			// Canonicalize order to reduce number of entries in visited.
			addr1, addr2 = addr2, addr1
		}

		// Short circuit if references are identical ...
		if addr1 == addr2 {
			return true, ""
		}

		// ... or already seen
		typ := v1.Type()
		v := visit{addr1, addr2, typ}
		if visited[v] {
			return true, ""
		}

		// Remember for later.
		visited[v] = true
	}

	switch v1.Kind() {
	case reflect.Array:
		for i := 0; i < v1.Len(); i++ {
			if equal, reason := deepValueEqual(v1.Index(i), v2.Index(i), visited, depth+1); !equal {
				return false, reason
			}
		}
		return true, ""
	case reflect.Slice:
		if v1.IsNil() != v2.IsNil() {
			return false, "one slice is nil, the other is not"
		}
		if v1.Len() != v2.Len() {
			return false, "slices have different lengths"
		}
		if v1.Pointer() == v2.Pointer() {
			return true, ""
		}
		for i := 0; i < v1.Len(); i++ {
			if equal, reason := deepValueEqual(v1.Index(i), v2.Index(i), visited, depth+1); !equal {
				return false, reason
			}
		}
		return true, ""
	case reflect.Interface:
		if v1.IsNil() || v2.IsNil() {
			return v1.IsNil() == v2.IsNil(), "both interfaces must be nil"
		}
		return deepValueEqual(v1.Elem(), v2.Elem(), visited, depth+1)
	case reflect.Ptr:
		return deepValueEqual(v1.Elem(), v2.Elem(), visited, depth+1)
	case reflect.Struct:
		for i, n := 0, v1.NumField(); i < n; i++ {
			if equal, reason := deepValueEqual(v1.Field(i), v2.Field(i), visited, depth+1); !equal {
				return false, reason
			}
		}
		return true, ""
	case reflect.Map:
		if v1.IsNil() != v2.IsNil() {
			return false, "one map is nil, one is not"
		}
		if v1.Len() != v2.Len() {
			return false, "maps have different lengths"
		}
		if v1.Pointer() == v2.Pointer() {
			return true, ""
		}
		for _, k := range v1.MapKeys() {
			if equal, reason := deepValueEqual(v1.MapIndex(k), v2.MapIndex(k), visited, depth+1); !equal {
				return false, reason
			}
		}
		return true, ""
	case reflect.Func:
		if v1.IsNil() && v2.IsNil() {
			return true, ""
		}
		// Can't do better than this:
		return false, "non-nil functions never compare equal"
	default:
		// Normal equality suffices
		return v1.Interface() == v2.Interface(), "scalar values differ"
	}
}

// Compare tests for deep equality. It uses normal == equality where
// possible but will scan elements of arrays, slices, maps, and fields of
// structs. In maps, keys are compared with == but elements use deep
// equality. DeepEqual correctly handles recursive types. Functions are equal
// only if they are both nil.
// An empty slice is not equal to a nil slice.
func Compare(a1, a2 interface{}) (bool, string) {
	if a1 == nil || a2 == nil {
		return a1 == a2, "nil values are of different types"
	}
	v1 := reflect.ValueOf(a1)
	v2 := reflect.ValueOf(a2)
	if v1.Type() != v2.Type() {
		return false, "values are of different types"
	}
	return deepValueEqual(v1, v2, make(map[visit]bool), 0)
}
