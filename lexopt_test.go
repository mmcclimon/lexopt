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

func (pt *parserTester) longOk(expect string) {
	if pt.Current != pt.Long(expect) {
		pt.t.Errorf(`.Current, expect .Long(%q), got %v`, expect, pt.Current)
	}
}

func (pt *parserTester) valueOk(expect string) {
	val, err := pt.Value()
	if err != nil {
		pt.t.Fatalf(".Value() return unexpected err: %s", err)
	}

	if val != expect {
		pt.t.Errorf(".Value(), expect %q, got %q", expect, val)
	}
}

func (pt *parserTester) noValueOk() {
	_, err := pt.Value()
	if err == nil {
		pt.t.Fatal("expected error from .Value()")
	}

	if !errors.Is(err, ErrNoToken) {
		pt.t.Errorf(".Value() retured weird error, expect %q, got %q", ErrNoToken, err)
	}
}

func TestPanics(t *testing.T) {
	pt := newTester(t)

	tests := map[string]func(string) Arg{
		"Short": pt.Short,
		"Long":  pt.Long,
	}

	for desc, method := range tests {
		t.Run(desc, func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Fatalf("%s did not panic on empty string", desc)
				}
			}()

			method("")
		})
	}
}

func TestSingleLongOpt(t *testing.T) {
	pt := newTester(t, "--foo")
	pt.nextOk()
	pt.longOk("foo")
	pt.emptyOk()
}

func TestNoOptions(t *testing.T) {
	pt := newTester(t)
	pt.emptyOk()
}

func TestDoubleDash(t *testing.T) {
	pt := newTester(t, "--foo", "--", "whatever")
	pt.nextOk()
	pt.emptyOk()
	pt.emptyOk()
}

func TestLongOptValues(t *testing.T) {
	tests := map[string][]string{
		"with equal":    {"--foo=bar"},
		"without equal": {"--foo", "bar"},
	}

	for desc, argv := range tests {
		t.Run(desc, func(t *testing.T) {
			pt := newTester(t, argv...)
			pt.nextOk()
			pt.longOk("foo")
			pt.valueOk("bar")
		})
	}
}

func TestNoValue(t *testing.T) {
	pt := newTester(t, "--foo")
	pt.nextOk()
	pt.longOk("foo")
	pt.noValueOk()
}
