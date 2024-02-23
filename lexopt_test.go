package lexopt

import (
	"errors"
	"testing"
)

type parserTester struct {
	*Parser
}

func newTester(argv ...string) *parserTester {
	return &parserTester{New(argv)}
}

func (pt *parserTester) nextOk(t *testing.T) {
	if !pt.Next() {
		t.Error(".Next() returned false, expected value")
	}
}

func (pt *parserTester) emptyOk(t *testing.T) {
	if pt.Next() {
		t.Error(".Next() returned true, expected empty")
	}
}

func (pt *parserTester) longOk(t *testing.T, expect string) {
	if pt.Current != pt.Long(expect) {
		t.Errorf(`.Current, expect .Long(%q), got %v`, expect, pt.Current)
	}
}

func (pt *parserTester) valueOk(t *testing.T, expect string) {
	val, err := pt.Value()
	if err != nil {
		t.Fatalf(".Value() return unexpected err: %s", err)
	}

	if val != expect {
		t.Errorf(".Value(), expect %q, got %q", expect, val)
	}
}

func (pt *parserTester) noValueOk(t *testing.T) {
	_, err := pt.Value()
	if err == nil {
		t.Fatal("expected error from .Value()")
	}

	if !errors.Is(err, ErrNoToken) {
		t.Errorf(".Value() retured weird error, expect %q, got %q", ErrNoToken, err)
	}
}

func TestPanics(t *testing.T) {
	pt := newTester()

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
	pt := newTester("--foo")
	pt.nextOk(t)
	pt.longOk(t, "foo")
	pt.emptyOk(t)
}

func TestNoOptions(t *testing.T) {
	pt := newTester()
	pt.emptyOk(t)
}

func TestDoubleDash(t *testing.T) {
	pt := newTester("--foo", "--", "whatever")
	pt.nextOk(t)
	pt.emptyOk(t)
	pt.emptyOk(t)
}

func TestLongOptValues(t *testing.T) {
	tests := map[string][]string{
		"with equal":    {"--foo=bar"},
		"without equal": {"--foo", "bar"},
	}

	for desc, argv := range tests {
		t.Run(desc, func(t *testing.T) {
			pt := newTester(argv...)
			pt.nextOk(t)
			pt.longOk(t, "foo")
			pt.valueOk(t, "bar")
		})
	}
}

func TestNoValue(t *testing.T) {
	pt := newTester("--foo")
	pt.nextOk(t)
	pt.longOk(t, "foo")
	pt.noValueOk(t)
}
