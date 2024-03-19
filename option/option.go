package option

import (
	"encoding/json"
	"strconv"
)

// Option is a simple generic type for representing
// missing/null values where the zero value is valid.
type Option[T any] struct {
	Valid bool
	Value T
}

// None returns an empty Option
func None[T any]() Option[T] {
	return Option[T]{
		Valid: false,
	}
}

// Some returns a filled Option
func Some[T any](val T) Option[T] {
	return Option[T]{
		Valid: true,
		Value: val,
	}
}

// Set fills the Option with a value
func (o *Option[T]) Set(val T) {
	o.Valid = true
	o.Value = val
}

// Wrapper type for Option[float32] to allow for
// JSON/Text Unmarshal implementations
type Float32 Option[float32]

func (f *Float32) MarshalText() ([]byte, error) {
	if f.Valid {
		s := strconv.FormatFloat(float64(f.Value), 'g', -1, 32)
		return []byte(s), nil
	} else {
		return []byte("null"), nil
	}
}

func (f *Float32) UnmarshalText(text []byte) error {
	if len(text) == 0 {
		f.Valid = false
		return nil
	}

	val, err := strconv.ParseFloat(string(text), 32)
	if err != nil {
		return err
	}

	f.Valid = true
	f.Value = float32(val)

	return nil
}

func (f *Float32) UnmarshalJSON(text []byte) error {
	if string(text) == "null" {
		f.Valid = false
		return nil
	}

	var n float32
	err := json.Unmarshal(text, &n)
	if err  != nil {
		return err
	}

	f.Valid = true
	f.Value = n

	return nil
}

func (f *Float32) MarshalJSON() ([]byte, error) {
	if !f.Valid {
		return []byte("null"), nil
	} else {
		return json.Marshal(f.Value)
	}
}

func (o *Float32) Set(val float32) {
	o.Valid = true
	o.Value = val
}

func (o *Float32) String() string {
	if !o.Valid {
		return "nan"
	} else {
		return strconv.FormatFloat(float64(o.Value), 'g', -1, 32)
	}
}
