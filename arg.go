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

type Arg struct {
	kind argType
	s    string
}

func Short(toMatch rune) Arg {
	return Arg{argShort, string(toMatch)}
}

func Long(toMatch string) Arg {
	return Arg{argLong, toMatch}
}

func Value(toMatch string) Arg {
	return Arg{argPlain, toMatch}
}

// Convenience functions for conversions.

// Unlike the rest of the conversions, String will never fail.
func (a Arg) String() string {
	return a.s
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

func (a Arg) MustString() string          { return a.String() }
func (a Arg) MustBool() bool              { return must(a.Bool()) }
func (a Arg) MustInt() int                { return must(a.Int()) }
func (a Arg) MustInt64() int64            { return must(a.Int64()) }
func (a Arg) MustUint() uint              { return must(a.Uint()) }
func (a Arg) MustUint64() uint64          { return must(a.Uint64()) }
func (a Arg) MustFloat64() float64        { return must(a.Float64()) }
func (a Arg) MustDuration() time.Duration { return must(a.Duration()) }
