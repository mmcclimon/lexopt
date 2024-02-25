package lexopt

import (
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"
)

type parserTester struct {
	*Parser
	t *testing.T
}

func newTester(t *testing.T, argv ...string) *parserTester {
	return &parserTester{New(argv), t}
}

func newTesterWs(t *testing.T, argv string) *parserTester {
	return &parserTester{New(strings.Fields(argv)), t}
}

func (pt *parserTester) nextOk() {
	pt.t.Helper()
	if !pt.Next() {
		pt.t.Error(".Next() returned false, expected value")
	}
}

func (pt *parserTester) emptyOk() {
	pt.t.Helper()
	if pt.Next() {
		pt.t.Error(".Next() returned true, expected empty")
	}
}

func (pt *parserTester) nextErrOk(expectErr error) {
	pt.t.Helper()
	pt.emptyOk()
	if err := pt.Err(); !errors.Is(err, expectErr) {
		pt.t.Errorf(".Err() did not get expected err, want %q, got %q", expectErr, err)
	}
}

func (pt *parserTester) longOk(expect string) {
	pt.t.Helper()
	pt.nextOk()
	if pt.Current != Long(expect) {
		pt.t.Errorf(`.Current, expect .Long(%q), got %v`, expect, pt.Current)
	}
}

func (pt *parserTester) shortOk(expect rune) {
	pt.t.Helper()
	pt.nextOk()
	if pt.Current != Short(expect) {
		pt.t.Errorf(`.Current, expect .Short(%q), got %v`, expect, pt.Current)
	}
}

func (pt *parserTester) positionalOk(expect string) {
	pt.t.Helper()
	pt.nextOk()
	if pt.Current != toPositional(expect) {
		pt.t.Errorf(".Next(), expect %q, got %q", expect, pt.Current)
	}
}

func (pt *parserTester) valueOk(expect string) {
	pt.t.Helper()
	val, err := pt.Value()
	if err != nil {
		pt.t.Fatalf(".Value() return unexpected err: %s", err)
	}

	if val != toValue(expect) {
		pt.t.Errorf(".Value(), expect %q, got %q", expect, val)
	}
}

func noValOk[T any](desc string, pt *parserTester, method func() (T, error)) {
	pt.t.Helper()
	_, err := method()
	if err == nil {
		pt.t.Fatalf("expected error from %s", desc)
	}

	if !errors.Is(err, ErrNoValue) {
		pt.t.Errorf("%s returned weird error, expect %q, got %q", desc, ErrNoValue, err)
	}

}

func (pt *parserTester) noValueOk() {
	pt.t.Helper()
	noValOk(".Value()", pt, pt.Value)
}

func (pt *parserTester) valuesOk(expectStrs ...string) {
	pt.t.Helper()

	expect := make([]Arg, len(expectStrs))
	for i, val := range expectStrs {
		expect[i] = toValue(val)
	}

	values, err := pt.Values()
	if err != nil {
		pt.t.Fatalf(".Values returned unexpected error, %s", err)
	}

	if !reflect.DeepEqual(values, expect) {
		pt.t.Errorf(".Values incorrect: want %v, got %v", expect, values)
	}
}

func (pt *parserTester) noValuesOk() {
	pt.t.Helper()
	noValOk(".Values()", pt, pt.Values)
}

func TestSingleLongOpt(t *testing.T) {
	pt := newTester(t, "--foo")
	pt.longOk("foo")
	pt.emptyOk()
}

func TestNoOptions(t *testing.T) {
	pt := newTester(t)
	pt.emptyOk()
}

func TestDoubleDash(t *testing.T) {
	pt := newTester(t, "--foo", "--", "whatever")
	pt.longOk("foo")
	pt.positionalOk("whatever")
}

func TestValueAfterEndOfOptions(t *testing.T) {
	pt := newTester(t, "--foo=bar", "--", "wat", "whatever")
	pt.longOk("foo")
	pt.valueOk("bar")
	pt.positionalOk("wat")
	pt.noValueOk()
}

func TestLongValues(t *testing.T) {
	tests := map[string][]string{
		"with equal":    {"--foo=bar"},
		"without equal": {"--foo", "bar"},
	}

	for desc, argv := range tests {
		t.Run(desc, func(t *testing.T) {
			pt := newTester(t, argv...)
			pt.longOk("foo")
			pt.valueOk("bar")
			pt.emptyOk()
		})
	}

	t.Run("unconsumed value", func(t *testing.T) {
		pt := newTester(t, "--foo=bar")
		pt.longOk("foo")
		pt.nextErrOk(ErrUnexpectedValue)
	})
}

func TestNoValue(t *testing.T) {
	pt := newTester(t, "--foo")
	pt.longOk("foo")
	pt.noValueOk()
}

func TestDashIsValue(t *testing.T) {
	t.Run("as long opt value", func(t *testing.T) {
		pt := newTester(t, "--file", "-")
		pt.longOk("file")
		pt.valueOk("-")
	})

	t.Run("as standalone value", func(t *testing.T) {
		pt := newTester(t, "-")
		pt.positionalOk("-")
	})
}

func TestShortBasic(t *testing.T) {
	pt := newTester(t, "-x", "-b")
	pt.shortOk('x')
	pt.shortOk('b')
	pt.emptyOk()
}

func TestShortCuddled(t *testing.T) {
	pt := newTester(t, "-xb", "foo")
	pt.shortOk('x')
	pt.shortOk('b')
	pt.positionalOk("foo")
	pt.emptyOk()
}

func TestShortValues(t *testing.T) {
	t.Run("cuddled", func(t *testing.T) {
		pt := newTester(t, "-uno")
		pt.shortOk('u')
		pt.valueOk("no")
	})

	t.Run("cuddled with multiples", func(t *testing.T) {
		pt := newTester(t, "-vuno")
		pt.shortOk('v')
		pt.shortOk('u')
		pt.valueOk("no")
	})

	t.Run("space", func(t *testing.T) {
		pt := newTester(t, "-u", "no")
		pt.shortOk('u')
		pt.valueOk("no")
	})

	t.Run("space with multiple", func(t *testing.T) {
		pt := newTester(t, "-vu", "no")
		pt.shortOk('v')
		pt.shortOk('u')
		pt.valueOk("no")
	})

	t.Run("cuddled with equal", func(t *testing.T) {
		pt := newTester(t, "-u=no")
		pt.shortOk('u')
		pt.valueOk("no")
	})

	t.Run("equal as value", func(t *testing.T) {
		pt := newTester(t, "-u=")
		pt.shortOk('u')
		pt.valueOk("")
	})

	t.Run("unconsumed equal", func(t *testing.T) {
		pt := newTester(t, "-u=foo")
		pt.shortOk('u')
		pt.nextErrOk(ErrUnexpectedValue)
	})
}

func TestOptionalValue(t *testing.T) {
	optOk := func(pt *parserTester, expect string) {
		val, ok := pt.OptionalValue()
		if !ok {
			t.Fatal(".OptionalValue() unexpectedly returned false")
		}

		if val != toValue(expect) {
			t.Errorf(".OptionalValue() returned bad value, want %v, want %v", expect, val)
		}
	}

	noOptOk := func(pt *parserTester) {
		val, ok := pt.OptionalValue()
		if ok {
			t.Fatalf(".OptionalValue() unexpectedly returned true: %v", val)
		}
	}

	pt := newTesterWs(t, "-a=foo --long=bar")
	pt.shortOk('a')
	optOk(pt, "foo")
	pt.longOk("long")
	optOk(pt, "bar")

	pt = newTesterWs(t, "-a foo --long bar")
	pt.shortOk('a')
	noOptOk(pt)
	pt.valueOk("foo")
	pt.longOk("long")
	noOptOk(pt)
}

func TestDumpState(t *testing.T) {
	var w strings.Builder
	pt := newTester(t, "-l")
	pt.dumpState(&w)

	if !strings.Contains(w.String(), "--- parser state ---") {
		pt.t.Fatalf("got nonsense from dumpState: %s", w.String())
	}
}

func TestArgConversions(t *testing.T) {
	a := toValue("-42")
	runConvOk(t, a, "int", -42, a.Int, a.MustInt)
	runConvErr(t, a, "bad uint", a.Uint, a.MustUint)
	runConvErr(t, a, "bad uint64", a.Uint64, a.MustUint64)

	a = toValue("-42")
	runConvOk(t, a, "int64", int64(-42), a.Int64, a.MustInt64)

	a = toValue("99")
	runConvOk(t, a, "uint", 99, a.Uint, a.MustUint)

	a = toValue("99")
	runConvOk(t, a, "uint64", uint64(99), a.Uint64, a.MustUint64)

	a = toValue("3.14")
	runConvOk(t, a, "float", float64(3.14), a.Float64, a.MustFloat64)

	a = toValue("true")
	runConvOk(t, a, "bool", true, a.Bool, a.MustBool)

	a = toValue("5m")
	runConvOk(t, a, "duration", time.Duration(5*time.Minute), a.Duration, a.MustDuration)

	a = toValue("hello")
	runConvOk(t, a, "string", "hello", a.String, a.MustString)
	runConvErr(t, a, "bad int", a.Int, a.MustInt)
	runConvErr(t, a, "bad int64", a.Int64, a.MustInt64)
	runConvErr(t, a, "bad float", a.Float64, a.MustFloat64)
	runConvErr(t, a, "bad bool", a.Bool, a.MustBool)
	runConvErr(t, a, "bad duration", a.Duration, a.MustDuration)

}

func runConvOk[T comparable](
	t *testing.T,
	a Arg,
	desc string,
	expect T,
	tryMethod func() (T, error),
	mustMethod func() T,
) {
	t.Run(desc, func(t *testing.T) {
		val, err := tryMethod()
		if err != nil {
			t.Fatalf("conversion returned unexpected err: %s", err)
		}

		if val != expect {
			t.Errorf("bad conversion: want %v, got %v", expect, val)
		}

		mustVal := mustMethod()
		if mustVal != expect {
			t.Errorf("bad conversion: want %v, got %v", expect, mustVal)
		}
	})
}

func runConvErr[T comparable](
	t *testing.T,
	a Arg,
	desc string,
	tryMethod func() (T, error),
	mustMethod func() T,
) {
	t.Run(desc, func(t *testing.T) {
		_, err := tryMethod()
		if err == nil {
			t.Fatal("conversion did not return expected err")
		}

		defer func() {
			r := recover()
			if r == nil {
				t.Errorf("expected panic, and did not")
			}
		}()

		mustMethod()
	})
}
