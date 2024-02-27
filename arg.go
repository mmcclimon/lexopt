package lexopt

import (
	"strconv"
	"time"
)

type argType int

const (
	argInvalid argType = iota
	argShort
	argLong
	argPlain
)

// Arg represents a command line argument. Arg has methods for parsing into
// various Go types, the same as the standard library [flag] package.
type Arg struct {
	kind argType
	s    string
}

// Short represents a short option. Short options have a single dash and are a
// single character.
func Short(toMatch rune) Arg {
	return Arg{argShort, string(toMatch)}
}

// Short represents a short option. Long options have a double dash, and can
// be any length.
func Long(toMatch string) Arg {
	return Arg{argLong, toMatch}
}

// Value represents a value argument. This could be a positional argument
// (obtained with [Parser.Next]) or an option value (obtained with
// [Parser.Value]).
func Value(toMatch string) Arg {
	return Arg{argPlain, toMatch}
}

// Convenience functions for conversions.

// String returns the raw string value of the argument. Unlike the rest of the
// conversions, String will never fail. Note that the value is returned with
// no leading dashes.
func (a Arg) String() string {
	return a.s
}

// MustString is exactly the same as [Arg.String], provided for symmetry with
// all the other Must methods.
func (a Arg) MustString() string { return a.String() }

// DashedString returns a formatted version of Arg. Short options are preceded
// with a single dash, long options with a double dash, and all other args as
// the raw arg value.
func (a Arg) DashedString() string {
	switch a.kind {
	case argShort:
		return "-" + a.s
	case argLong:
		return "--" + a.s
	default:
		return a.s
	}
}

// Bool converts Arg to a boolean, using [strconv.ParseBool].
func (a Arg) Bool() (bool, error) {
	return strconv.ParseBool(a.s)
}

// MustBool is like [Arg.Bool], but panics if the conversion fails.
func (a Arg) MustBool() bool { return must(a.Bool()) }

// Bool converts Arg to an int, using [strconv.ParseInt].
func (a Arg) Int() (int, error) {
	val, err := strconv.ParseInt(a.s, 10, strconv.IntSize)
	return int(val), err
}

// MustInt is like [Arg.Int], but panics if the conversion fails.
func (a Arg) MustInt() int { return must(a.Int()) }

// Int64 converts Arg to an int64, using [strconv.ParseInt].
func (a Arg) Int64() (int64, error) {
	return strconv.ParseInt(a.s, 10, 64)
}

// MustInt64 is like [Arg.Int64], but panics if the conversion fails.
func (a Arg) MustInt64() int64 { return must(a.Int64()) }

// Uint converts Arg to a uint, using [strconv.ParseUint].
func (a Arg) Uint() (uint, error) {
	val, err := strconv.ParseUint(a.s, 10, strconv.IntSize)
	return uint(val), err
}

// MustUint is like [Arg.Uint], but panics if the conversion fails.
func (a Arg) MustUint() uint { return must(a.Uint()) }

// Uint64 converts Arg to a uint64, using [strconv.ParseUint].
func (a Arg) Uint64() (uint64, error) {
	return strconv.ParseUint(a.s, 10, 64)
}

// MustUint64 is like [Arg.Uint64], but panics if the conversion fails.
func (a Arg) MustUint64() uint64 { return must(a.Uint64()) }

// Float64 converts Arg to a float64, using [strconv.ParseFloat].
func (a Arg) Float64() (float64, error) {
	return strconv.ParseFloat(a.s, 64)
}

// MustFloat64 is like [Arg.Float64], but panics if the conversion fails.
func (a Arg) MustFloat64() float64 { return must(a.Float64()) }

// Duration converts Arg to a [time.Duration], using [time.ParseDuration].
func (a Arg) Duration() (time.Duration, error) {
	return time.ParseDuration(a.s)
}

// MustDuration is like [Arg.Duration], but panics if the conversion fails.
func (a Arg) MustDuration() time.Duration { return must(a.Duration()) }

func must[T any](val T, err error) T {
	if err != nil {
		panic(err)
	}

	return val
}
