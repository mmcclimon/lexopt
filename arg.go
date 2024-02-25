package lexopt

import (
	"strconv"
	"time"
)

type argType int

const (
	argUnmatched argType = iota
	argShort
	argLong
	argPositional
	argValue
)

type Arg struct {
	kind argType
	s    string
}

func toPositional(s string) Arg {
	return Arg{argPositional, s}
}

func noMatch() Arg {
	return Arg{kind: argUnmatched}
}

func toValue(s string) Arg {
	return Arg{argValue, s}
}

func Short(toMatch rune) Arg {
	return Arg{argShort, string(toMatch)}
}

func Long(toMatch string) Arg {
	return Arg{argLong, toMatch}
}

// Convenience functions for conversions.

func (a Arg) String() (string, error) {
	return a.s, nil
}

func (a Arg) Bool() (bool, error) {
	return strconv.ParseBool(a.s)
}

func (a Arg) Int() (int, error) {
	val, err := strconv.ParseInt(a.s, 10, strconv.IntSize)
	return int(val), err
}

func (a Arg) Int64() (int64, error) {
	return strconv.ParseInt(a.s, 10, 64)
}

func (a Arg) Uint() (uint, error) {
	val, err := strconv.ParseUint(a.s, 10, strconv.IntSize)
	return uint(val), err
}

func (a Arg) Uint64() (uint64, error) {
	return strconv.ParseUint(a.s, 10, 64)
}

func (a Arg) Float64() (float64, error) {
	return strconv.ParseFloat(a.s, 64)
}

func (a Arg) Duration() (time.Duration, error) {
	return time.ParseDuration(a.s)
}

func must[T any](val T, err error) T {
	if err != nil {
		panic(err)
	}

	return val
}

func (a Arg) MustString() string          { return must(a.String()) }
func (a Arg) MustBool() bool              { return must(a.Bool()) }
func (a Arg) MustInt() int                { return must(a.Int()) }
func (a Arg) MustInt64() int64            { return must(a.Int64()) }
func (a Arg) MustUint() uint              { return must(a.Uint()) }
func (a Arg) MustUint64() uint64          { return must(a.Uint64()) }
func (a Arg) MustFloat64() float64        { return must(a.Float64()) }
func (a Arg) MustDuration() time.Duration { return must(a.Duration()) }
