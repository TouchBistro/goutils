package color_test

import (
	"testing"

	"github.com/TouchBistro/goutils/color"
)

func TestColors(t *testing.T) {
	color.SetEnabled(true)
	tests := []struct {
		name    string
		colorFn func(string) string
		input   string
		want    string
	}{
		{
			"Black",
			color.Black,
			"foo bar",
			"\x1b[30mfoo bar\x1b[39m",
		},
		{
			"Red",
			color.Red,
			"foo bar",
			"\x1b[31mfoo bar\x1b[39m",
		},
		{
			"Green",
			color.Green,
			"foo bar",
			"\x1b[32mfoo bar\x1b[39m",
		},
		{
			"Yellow",
			color.Yellow,
			"foo bar",
			"\x1b[33mfoo bar\x1b[39m",
		},
		{
			"Blue",
			color.Blue,
			"foo bar",
			"\x1b[34mfoo bar\x1b[39m",
		},
		{
			"Magenta",
			color.Magenta,
			"foo bar",
			"\x1b[35mfoo bar\x1b[39m",
		},
		{
			"Cyan",
			color.Cyan,
			"foo bar",
			"\x1b[36mfoo bar\x1b[39m",
		},
		{
			"White",
			color.White,
			"foo bar",
			"\x1b[37mfoo bar\x1b[39m",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.colorFn(tt.input)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestStripReset(t *testing.T) {
	color.SetEnabled(true)
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"single reset", "foo \x1b[39mbar", "\x1b[31mfoo bar\x1b[39m"},
		{"multiple resets", "foo \x1b[39m\x1b[39mbar", "\x1b[31mfoo bar\x1b[39m"},
	}
	for _, tt := range tests {
		got := color.Red(tt.in)
		if got != tt.want {
			t.Errorf("got %q, want %q", got, tt.want)
		}
	}
}

func TestColorDisabled(t *testing.T) {
	color.SetEnabled(false)
	got := color.Red("foo bar")
	want := "foo bar"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func BenchmarkRed(b *testing.B) {
	color.SetEnabled(true)
	b.Run("no strip", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			color.Red("foo bar")
		}
	})
	b.Run("strip", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			color.Red("foo \x1b[39m\x1b[39mbar")
		}
	})
}

// Using Regex
// BenchmarkRed/no_strip-16         	  379442	      2852 ns/op	    1456 B/op	      23 allocs/op
// BenchmarkRed/strip-16            	  365137	      3242 ns/op	    1456 B/op	      23 allocs/op

// Using custom replace
// BenchmarkRed/no_strip-16         	 6109512	       190.6 ns/op	      64 B/op	       4 allocs/op
// BenchmarkRed/strip-16            	 5570493	       211.9 ns/op	      64 B/op	       4 allocs/op
