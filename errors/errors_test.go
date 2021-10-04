package errors_test

import (
	"fmt"
	"testing"

	"github.com/TouchBistro/goutils/errors"
)

type errkind uint8

const (
	invalid errkind = iota
	internal
)

func (k errkind) Kind() string {
	switch k {
	case invalid:
		return "invalid operation"
	case internal:
		return "internal error"
	}
	return "unknown error"
}

func TestErrorFormat(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		format string
		want   string
	}{
		{
			name:   "new error",
			err:    errors.New(internal, "something blew up", errors.Op("test.Foo")),
			format: "%s",
			want:   "internal error: something blew up",
		},
		{
			name: "string format",
			err: errors.Wrap(
				internal,
				"unable to create file",
				errors.Op("test.Foo"),
				fmt.Errorf("dir not exist"),
			),
			format: "%s",
			want:   "internal error: unable to create file: dir not exist",
		},
		{
			name: "detailed format",
			err: errors.Wrap(
				internal,
				"unable to create file",
				errors.Op("test.Foo"),
				fmt.Errorf("dir not exist"),
			),
			format: "%+v",
			want:   "test.Foo: internal error: unable to create file: dir not exist",
		},
		{
			name: "detailed format with nested error",
			err: errors.Wrap(
				invalid,
				"cannot find file",
				errors.Op("test.Bar"),
				errors.Wrap(
					internal,
					"no file for path",
					errors.Op("test.Foo"),
					fmt.Errorf("file not exist"),
				),
			),
			format: "%+v",
			want:   "test.Bar: invalid operation: cannot find file:\n\ttest.Foo: internal error: no file for path: file not exist",
		},
		{
			name: "hoists kind wrapping error",
			err: errors.Annotate(
				"cannot find file",
				errors.Op("test.Bar"),
				errors.New(
					internal,
					"no file for path",
					errors.Op("test.Foo"),
				),
			),
			format: "%s",
			want:   "internal error: cannot find file: no file for path",
		},
		{
			name: "removes duplicate kind",
			err: errors.Wrap(
				internal,
				"cannot find file",
				errors.Op("test.Bar"),
				errors.New(
					internal,
					"no file for path",
					errors.Op("test.Foo"),
				),
			),
			format: "%s",
			want:   "internal error: cannot find file: no file for path",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := fmt.Sprintf(tt.format, tt.err)
			if s != tt.want {
				t.Errorf("got\n\t%s\nwant\n\t%s", s, tt.want)
			}
		})
	}
}

func TestDoesNotMutatePreviousError(t *testing.T) {
	err1 := errors.New(invalid, "you can't do that", "")
	err2 := errors.Annotate("no mutation", "", err1)
	want := "invalid operation: no mutation: you can't do that"
	if err2.Error() != want {
		t.Errorf("got\n\t%s\nwant\n\t%s", err2, want)
	}

	// Check that err1.Kind was not mutated when err2 wrapped it
	kind := err1.(*errors.Error).Kind
	if kind != invalid {
		t.Errorf("got kind\n\t%s\nwant\n\t%s", kind.Kind(), invalid.Kind())
	}
}

func TestListFormat(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		format string
		want   string
	}{
		{
			name: "string format",
			err: errors.List{
				errors.New(internal, "something went wrong", errors.Op("test.Foo")),
				fmt.Errorf("something blew up"),
				errors.String("oops"),
			},
			format: "%s",
			want:   "internal error: something went wrong\nsomething blew up\noops",
		},
		{
			name: "detailed format",
			err: errors.List{
				errors.New(internal, "something went wrong", errors.Op("test.Foo")),
				fmt.Errorf("something blew up"),
				errors.String("oops"),
			},
			format: "%+v",
			want:   "test.Foo: internal error: something went wrong\nsomething blew up\noops",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := fmt.Sprintf(tt.format, tt.err)
			if s != tt.want {
				t.Errorf("got\n\t%s\nwant\n\t%s", s, tt.want)
			}
		})
	}
}

type pathError struct {
	path string
	msg  string
}

func (e *pathError) Error() string {
	return e.path + ": " + e.msg
}

func TestIs(t *testing.T) {
	const eof errors.String = "EOF"
	err := errors.Wrap(internal, "unexpected end of file", errors.Op("config.Read"), eof)
	if !errors.Is(err, eof) {
		t.Error("want err to contain eof")
	}
}

func TestAs(t *testing.T) {
	pathErr := &pathError{"/foo/bar", "file not found"}
	err := errors.Wrap(invalid, "source does not exist", errors.Op("config.Read"), pathErr)
	var gotErr *pathError
	if !errors.As(err, &gotErr) {
		t.Fatal("want err to contain an error of type *pathError")
	}
	if *gotErr != *pathErr {
		t.Errorf("got err\n\t%s\nwant\n\t%s", gotErr, pathErr)
	}
}
