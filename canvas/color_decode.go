package canvas

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sync"
)

// Helper struct for parsing color values
type colorValue struct {
	Color
}

func (c *colorValue) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		// Mimic the behaviour of json.Unmarshal when decoding
		// null into an interface
		c.Color = nil
		return nil
	}
	var colorStr string
	err := json.Unmarshal(data, &colorStr)
	if err != nil {
		return err
	}

	color, err := ParseColor(colorStr)
	if err != nil {
		return err
	}
	if color != nil {
		c.Color = color
	}

	return nil
}

// Helper function for decoding structs that contain [Color]
// interfaces. Not intended for external use.
func UnmarshalColorStruct(data []byte, val any) error {
	v := reflect.ValueOf(val)

	// Convert the type into a "safe" version
	safeTy := makeDecodableType(v.Type())

	// Construct an instance of the new type
	safeVal := reflect.New(safeTy)

	// Copy the given value into the safe version
	assign(safeVal.Elem(), v)

	// Decode into the "safe" type, we don't return the error here
	// to mimic the error behaviour of json.Unmarshal
	err := json.Unmarshal(data, safeVal.Interface())

	// Assign the decoded value back to the destination
	assign(v, safeVal.Elem())

	return err
}

var (
	colorType       reflect.Type = reflect.TypeFor[Color]()
	colorValueType  reflect.Type = reflect.TypeFor[colorValue]()
)

var typeCache sync.Map // map[reflect.Type]reflect.Type

type typeConverter struct {
	typeStack []reflect.Type
}

func makeDecodableType(t reflect.Type) reflect.Type {
	ty, ok := typeCache.Load(t)
	if ok {
		return ty.(reflect.Type)
	}

	c := typeConverter{}
	ty = c.convert(t)

	ty, _ = typeCache.LoadOrStore(t, ty)
	return ty.(reflect.Type)
}

func (c *typeConverter) convert(t reflect.Type) reflect.Type {
	// If we've seen this type before, we have a recursive type.
	// Go reflection doesn't allow for the construction of recursive
	// types, so we make do with just returning the original type.
	// This works fine if `t` implements `json.Unmarshaler`, as it
	// will call `UnmarshalJSON` instead of continuing to reflect and
	// we'll use the converted version of `t` in that case
	if c.seenType(t) {
		return t
	}
	// Manage the stack of seen types
	c.pushType(t)
	defer c.popType()

	switch t.Kind() {
	case reflect.Struct:

		numFields := t.NumField()
		newFields := make([]reflect.StructField, 0, numFields)

		// Convert the fields, if we can't or won't convert the type,
		// we keep the original field to ensure that the new struct type
		// matches the old one
		anyNew := false
		for i := 0; i < numFields; i++ {
			f := t.Field(i)
			if f.IsExported() {
				// Only include exported fields, this is to handle
				// unexported anonymous/embedded fields

				newType := c.convert(f.Type)
				anyNew = anyNew || f.Type != newType
				f.Type = newType

				newFields = append(newFields, f)
			}
		}

		if !anyNew {
			ty, _ := typeCache.LoadOrStore(t, t)
			return ty.(reflect.Type)
		}

		return reflect.StructOf(newFields)

	case reflect.Pointer:
		return reflect.PointerTo(c.convert(t.Elem()))
	case reflect.Map:
		// Assume the key type is safe
		keyTy := t.Key()
		elemType := makeDecodableType(t.Elem())
		return reflect.MapOf(keyTy, elemType)
	case reflect.Array:
		len := t.Len()
		return reflect.ArrayOf(len, c.convert(t.Elem()))
	case reflect.Slice:
		return reflect.SliceOf(c.convert(t.Elem()))
	case reflect.Interface:
		if t == colorType {
			// Replace instances of `Color` with `colorValue`
			return colorValueType
		}
	}

	return t
}

func (c *typeConverter) seenType(ty reflect.Type) bool {
	for _, t := range c.typeStack {
		if t == ty { return true }
	}

	return false
}

func (c *typeConverter) pushType(t reflect.Type) {
	c.typeStack = append(c.typeStack, t)
}

func (c *typeConverter) popType() {
	if len(c.typeStack) > 0 {
		c.typeStack = c.typeStack[:len(c.typeStack)-1]
	}
}

// Assigns src to dst, converting `Color` and `colorValue` types
// as appropriate
func assign(dst, src reflect.Value) {
	if src.Type() == colorValueType {
		// The source value is `colorValue`, so copy the
		// the `Color` field to `dst`
		// This has to be early, as `colorValue` is assignable to
		// `Color`
		color := src.Field(0)
		dst.Set(color)
	} else if dst.Type() == colorValueType {
		// We're going the other way, assigning a `Color` to a `colorValue`
		//	dst.Color = src
		if !src.IsNil() {
			color := dst.Field(0)
			color.Set(src)
		}
	} else if src.Type().AssignableTo(dst.Type()) {
		// `src` is a assignable to `dst`
		dst.Set(src)
	} else {
		if dst.Kind() != src.Kind() {
			panic(fmt.Sprintf("Error assigning values: %s to %s", src.Kind(), dst.Kind()))
		}
		switch dst.Kind() {
		case reflect.Pointer:
			if src.IsNil() {
				// src is nil, don't bother recursing and just set `dst` to nil
				dst.Set(reflect.Zero(dst.Type()))
				return
			}
			if dst.IsNil() {
				// Ensure the destination is not nil
				dst.Set(reflect.New(dst.Type().Elem()))
			}
			assign(dst.Elem(), src.Elem())
		case reflect.Struct:
			dst.Set(reflect.Zero(dst.Type()))
			numFields := src.NumField()

			for i := 0; i < numFields; i++ {
				srcField := src.Type().Field(i)

				dstField, ok := dst.Type().FieldByName(srcField.Name)
				if ok && len(dstField.Index) == 1 {
					srcFieldVal := src.Field(i)
					dstFieldVal := dst.FieldByIndex(dstField.Index)

					assign(dstFieldVal, srcFieldVal)
				}
			}
		case reflect.Array, reflect.Slice:
			if src.IsNil() {
				dst.Set(reflect.Zero(dst.Type()))
				return
			}

			if dst.Kind() == reflect.Slice {
				if dst.IsNil() || dst.Cap() < src.Len() {
					// Either `dst` == nil, or it doesn't have enough capacity to store
					// `src`, allocate a new slice
					dst.Set(reflect.MakeSlice(dst.Type(), src.Len(), src.Len()))
				} else {
					dst.SetLen(src.Len())
				}
			}

			for i := 0; i < dst.Len(); i++ {
				assign(dst.Index(i), src.Index(i))
			}
		case reflect.Map:
			if src.IsNil() {
				dst.Set(reflect.Zero(dst.Type()))
				return
			}

			if dst.IsNil() {
				dst.Set(reflect.MakeMap(dst.Type()))
			}

			iter := src.MapRange()
			for iter.Next() {
				srcKey := iter.Key()
				srcVal := iter.Value()
				val := dst.MapIndex(srcKey)
				if !val.IsValid() {
					val = reflect.New(dst.Type().Elem()).Elem()
				}
				assign(val, srcVal)
				dst.SetMapIndex(srcKey, val)
			}
		default:
			panic(fmt.Sprintf("Unhandled assignment for '%s' (%v <- %v)", dst.Kind(), dst, src))
		}
	}
}
