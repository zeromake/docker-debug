package stream

import (
	"github.com/docker/docker/pkg/term"
)

// Streams interface
type Streams interface {
	Out() *OutStream
	In() *InStream
}

// CommonStream is an input stream used by the DockerCli to read user input
type CommonStream struct {
	fd         uintptr
	isTerminal bool
	state      *term.State
}

// FD returns the file descriptor number for this stream
func (s *CommonStream) FD() uintptr {
	return s.fd
}

// IsTerminal returns true if this stream is connected to a terminal
func (s *CommonStream) IsTerminal() bool {
	return s.isTerminal
}

// RestoreTerminal restores normal mode to the terminal
func (s *CommonStream) RestoreTerminal() error {
	if s.state != nil {
		return term.RestoreTerminal(s.fd, s.state)
	}
	return nil
}

// SetIsTerminal sets the boolean used for isTerminal
func (s *CommonStream) SetIsTerminal(isTerminal bool) {
	s.isTerminal = isTerminal
}
