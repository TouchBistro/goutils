package textutil_test

import (
	"strings"
	"testing"

	"github.com/TouchBistro/goutils/textutil"
)

var expandVariablesTests = []struct {
	name string
	in   string
	out  string
}{
	{"empty", "", ""},
	{"no vars", "nothing to expand", "nothing to expand"},
	{"just a var", "${HOME}", "/home/foo"},
	{"var in middle", "start ${HOME} end", "start /home/foo end"},
	{"multiple vars", "foo ${first} bar ${second} baz", "foo abc bar def baz"},
	{"$", "$", "$"},
	{"$}", "$}", "$}"},
	{"${", "${", "${"},    // invalid syntax, will ignore
	{"${}", "${}", "${}"}, // invalid syntax, will ignore
	{"contains not vars", "start $HOME ${first} $$", "start $HOME abc $$"},
	{"non-alphanum var", "path: ${@env:HOME}", "path: $HOME"},
	{"side by side", "${first}${second}", "abcdef"},
}

func testMapping(name string) string {
	if strings.HasPrefix(name, "@env:") {
		return "$" + strings.TrimPrefix(name, "@env:")
	}
	switch name {
	case "HOME":
		return "/home/foo"
	case "first":
		return "abc"
	case "second":
		return "def"
	}
	return "UNKNOWN_VAR"
}

func TestExpandVariables(t *testing.T) {
	for _, tt := range expandVariablesTests {
		t.Run(tt.name, func(t *testing.T) {
			got := textutil.ExpandVariables([]byte(tt.in), testMapping)
			if string(got) != string(tt.out) {
				t.Errorf("got %q, want %q", got, tt.out)
			}
		})
	}
}

func TestExpandVariablesString(t *testing.T) {
	for _, tt := range expandVariablesTests {
		t.Run(tt.name, func(t *testing.T) {
			got := textutil.ExpandVariablesString(tt.in, testMapping)
			if got != tt.out {
				t.Errorf("got %q, want %q", got, tt.out)
			}
		})
	}
}

func BenchmarkExpandVariables(b *testing.B) {
	b.Run("no-op", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			textutil.ExpandVariables([]byte("noop noop noop noop"), func(s string) string { return "" })
		}
	})
	b.Run("multiple", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			textutil.ExpandVariables([]byte("${foo} ${foo} ${foo} ${foo}"), func(s string) string { return "bar" })
		}
	})
}

func BenchmarkExpandVariablesString(b *testing.B) {
	b.Run("no-op", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			textutil.ExpandVariablesString("noop noop noop noop", func(s string) string { return "" })
		}
	})
	b.Run("multiple", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			textutil.ExpandVariablesString("${foo} ${foo} ${foo} ${foo}", func(s string) string { return "bar" })
		}
	})
}
