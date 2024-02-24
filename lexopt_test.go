package lexopt

import (
	"errors"
	"testing"
)

type parserTester struct {
	*Parser
	t *testing.T
}

func newTester(t *testing.T, argv ...string) *parserTester {
	return &parserTester{New(argv), t}
}

func (pt *parserTester) nextOk() {
	if !pt.Next() {
		pt.t.Error(".Next() returned false, expected value")
	}
}

func (pt *parserTester) emptyOk() {
	if pt.Next() {
		pt.t.Error(".Next() returned true, expected empty")
	}
}

func (pt *parserTester) nextErrOk(expectErr error) {
	pt.emptyOk()

	if err := pt.Err(); !errors.Is(err, expectErr) {
		pt.t.Errorf(".Err() did not get expected err, want %q, got %q", expectErr, err)
	}
}

func (pt *parserTester) longOk(expect string) {
	pt.nextOk()
	if pt.Current != Long(expect) {
		pt.t.Errorf(`.Current, expect .Long(%q), got %v`, expect, pt.Current)
	}
}

func (pt *parserTester) shortOk(expect string) {
	pt.nextOk()
	if pt.Current != Short(expect) {
		pt.t.Errorf(`.Current, expect .Short(%q), got %v`, expect, pt.Current)
	}
}

func (pt *parserTester) positionalOk(expect string) {
	pt.nextOk()
	if pt.Current != toPositional(expect) {
		pt.t.Errorf(".Next(), expect %q, got %q", expect, pt.Current)
	}
}

func (pt *parserTester) valueOk(expect string) {
	val, err := pt.Value()
	if err != nil {
		pt.t.Fatalf(".Value() return unexpected err: %s", err)
	}

	if val.String() != expect {
		pt.t.Errorf(".Value(), expect %q, got %q", expect, val)
	}
}

func (pt *parserTester) noValueOk() {
	_, err := pt.Value()
	if err == nil {
		pt.t.Fatal("expected error from .Value()")
	}

	if !errors.Is(err, ErrNoValue) {
		pt.t.Errorf(".Value() retured weird error, expect %q, got %q", ErrNoToken, err)
	}
}

func TestPanics(t *testing.T) {
	assertPanics := func(desc string, method func(string) Arg, input string) {
		t.Run(desc, func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Fatalf("did not panic with input %q", input)
				}
			}()

			method(input)
		})
	}

	assertPanics("short empty", Short, "")
	assertPanics("short overlong", Short, "abc")
	assertPanics("long", Long, "")
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
	pt.shortOk("x")
	pt.shortOk("b")
	pt.emptyOk()
}

func TestShortCuddled(t *testing.T) {
	pt := newTester(t, "-xb", "foo")
	pt.shortOk("x")
	pt.shortOk("b")
	pt.positionalOk("foo")
	pt.emptyOk()
}

func TestShortValues(t *testing.T) {
	t.Run("cuddled", func(t *testing.T) {
		pt := newTester(t, "-uno")
		pt.shortOk("u")
		pt.valueOk("no")
	})

	t.Run("cuddled with multiples", func(t *testing.T) {
		pt := newTester(t, "-vuno")
		pt.shortOk("v")
		pt.shortOk("u")
		pt.valueOk("no")
	})

	t.Run("space", func(t *testing.T) {
		pt := newTester(t, "-u", "no")
		pt.shortOk("u")
		pt.valueOk("no")
	})

	t.Run("space with multiple", func(t *testing.T) {
		pt := newTester(t, "-vu", "no")
		pt.shortOk("v")
		pt.shortOk("u")
		pt.valueOk("no")
	})

	t.Run("cuddled with equal", func(t *testing.T) {
		pt := newTester(t, "-u=no")
		pt.shortOk("u")
		pt.valueOk("no")
	})

	t.Run("equal as short", func(t *testing.T) {
		pt := newTester(t, "-u=")
		pt.shortOk("u")
		pt.shortOk("=")
	})

	t.Run("equal as value", func(t *testing.T) {
		pt := newTester(t, "-u=")
		pt.shortOk("u")
		pt.noValueOk()
	})

	t.Run("unconsumed equal", func(t *testing.T) {
		pt := newTester(t, "-u=foo")
		pt.shortOk("u")
		pt.nextErrOk(ErrUnexpectedValue)
	})
}
