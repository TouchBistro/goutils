package spinner

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

var frames = [...]string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// Spinner represents the state of the spinner.
type Spinner struct {
	interval time.Duration
	out      io.Writer
	mu       *sync.RWMutex
	// stopChan is used to stop the spinner
	stopChan chan struct{}
	active   bool
	// last string written to out
	lastOutput string
	startMsg   string
	stopMsg    string
	// msg written on each frame
	msg string
	// total number of items
	count int
	// number of items completed
	completed int
	maxMsgLen int
}

// New creates a new spinner instance using the given options.
func New(opts ...Option) *Spinner {
	s := &Spinner{
		interval: 100 * time.Millisecond,
		out:      os.Stderr,
		mu:       &sync.RWMutex{},
		stopChan: make(chan struct{}, 1),
		active:   false,
		// default to 1 since we don't show progress on 1 anyway
		count:     1,
		maxMsgLen: 80,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// Option is a function that takes a spinner and applies
// a configuration to it.
type Option func(*Spinner)

// WithInterval sets how often the spinner updates.
// This controls the speed of the spinner.
// By default the interval is 100ms.
func WithInterval(d time.Duration) Option {
	return func(s *Spinner) {
		s.interval = d
	}
}

// WithWriter sets the writer that should be used for writting the spinner to.
func WithWriter(w io.Writer) Option {
	return func(s *Spinner) {
		s.out = w
	}
}

// WithStartMessage sets a string that should be written after the spinner
// when the spinnner is started.
func WithStartMessage(m string) Option {
	return func(s *Spinner) {
		s.startMsg = m
	}
}

// WithStopMessage sets a string that should be written when the spinner is stopped.
// This message will replace the spinner.
func WithStopMessage(m string) Option {
	return func(s *Spinner) {
		s.stopMsg = m
	}
}

// WithCount sets the total number of items to track the progress of.
func WithCount(c int) Option {
	return func(s *Spinner) {
		s.count = c
	}
}

// WithMaxMessageLength sets the maximum length of the message that is written
// by the spinner. If the message is longer then this length it will be truncated.
// The default max length is 80.
func WithMaxMessageLength(l int) Option {
	return func(s *Spinner) {
		s.maxMsgLen = l
	}
}

// Start will start the spinner.
// If the spinner is already running, Start will do nothing.
func (s *Spinner) Start() {
	s.mu.Lock()
	if s.active {
		s.mu.Unlock()
		return
	}
	s.active = true
	s.setMsg(s.startMsg)
	s.mu.Unlock()
	go s.run()
}

// Stop stops the spinner if it is currently running.
// If the spinner is not running, Stop will do nothing.
func (s *Spinner) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.active {
		return
	}

	s.active = false
	s.erase()
	if s.stopMsg != "" {
		// Make sure there's a trailing newline
		if s.stopMsg[len(s.stopMsg)-1] != '\n' {
			s.stopMsg += "\n"
		}
		fmt.Fprint(s.out, s.stopMsg)
	}
	s.stopChan <- struct{}{}
}

// Inc increments the progress of the spinner. If the spinner
// has already reached full progress, Inc does nothing.
func (s *Spinner) Inc() {
	s.IncWithMessage("")
}

// IncWithMessage increments the progress of the spinner and updates
// the spinner message to m. If the spinner has already reached
// full progress, IncWithMessage does nothing.
func (s *Spinner) IncWithMessage(m string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.completed >= s.count {
		return
	}
	s.completed++
	s.setMsg(m)
}

// IncWithMessagef increments the progress of the spinner and updates
// the spinner message to the format specifier. If the spinner has already
// reached full progress, IncWithMessagef does nothing.
func (s *Spinner) IncWithMessagef(format string, args ...interface{}) {
	s.IncWithMessage(fmt.Sprintf(format, args...))
}

// setMsg sets the spinner message to m. If m is longer then s.maxMsgLen it will
// be truncated. If m is empty, setMsg will do nothing.
// The caller must already hold s.lock.
func (s *Spinner) setMsg(m string) {
	if m == "" {
		return
	}
	// Truncate msg if it's too long
	if len(m) > s.maxMsgLen {
		// DISCUSS(@cszatmary): Should the ... count as part of the length or no?
		m = m[:s.maxMsgLen] + "..."
	}
	// Make sure message has a leading space to pad between it and the
	// spinner icon
	if m[0] != ' ' {
		m = " " + m
	}
	s.msg = m
}

// run runs the spinner. It should be called in a separate goroutine because
// it will run forever until it receives a value on s.stopChan.
func (s *Spinner) run() {
	for {
		for i := 0; i < len(frames); i++ {
			select {
			case <-s.stopChan:
				return
			default:
				s.mu.Lock()
				if !s.active {
					s.mu.Unlock()
					return
				}
				s.erase()

				line := fmt.Sprintf("\r%s%s ", frames[i], s.msg)
				if s.count > 1 {
					line += fmt.Sprintf("(%d/%d) ", s.completed, s.count)
				}
				fmt.Fprint(s.out, line)
				s.lastOutput = line
				d := s.interval

				s.mu.Unlock()
				time.Sleep(d)
			}
		}
	}
}

// erase deletes written characters. The caller must already hold s.lock.
func (s *Spinner) erase() {
	n := utf8.RuneCountInString(s.lastOutput)
	if runtime.GOOS == "windows" {
		clearString := "\r" + strings.Repeat(" ", n) + "\r"
		fmt.Fprint(s.out, clearString)
		s.lastOutput = ""
		return
	}

	// "\033[K" for macOS Terminal
	for _, c := range []string{"\b", "\127", "\b", "\033[K"} {
		fmt.Fprint(s.out, strings.Repeat(c, n))
	}
	// erases to end of line
	fmt.Fprintf(s.out, "\r\033[K")
	s.lastOutput = ""
}
