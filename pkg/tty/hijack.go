package tty

import (
	"context"
	"io"
	"net"
	"runtime"
	"sync"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/ioutils"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/moby/term"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/zeromake/docker-debug/pkg/stream"
)

// The default escape key sequence: ctrl-p, ctrl-q
// TODO: This could be moved to `pkg/term`.
var defaultEscapeKeys = []byte{16, 17}

// HijackedIOStreamer handles copying input to and output from streams to the
// connection.
type HijackedIOStreamer struct {
	Streams      stream.Streams
	InputStream  io.ReadCloser
	OutputStream io.Writer
	ErrorStream  io.Writer

	Resp types.HijackedResponse

	TTY        bool
	DetachKeys string
}

// Stream handles setting up the IO and then begins streaming stdin/stdout
// to/from the hijacked connection, blocking until it is either done reading
// output, the user inputs the detach key sequence when in TTY mode, or when
// the given context is cancelled.
func (h *HijackedIOStreamer) Stream(ctx context.Context) error {
	restoreInput, err := h.setupInput()
	if err != nil {
		return errors.Errorf("unable to setup input stream: %s", err)
	}

	defer restoreInput()
	defer h.Resp.Close()
	outputDone := h.beginOutputStream(restoreInput)
	inputDone, detached := h.beginInputStream(ctx, restoreInput)

	select {
	case err = <-outputDone:
		return errors.WithStack(err)
	case <-inputDone:
		// Input stream has closed.
		if h.OutputStream != nil || h.ErrorStream != nil {
			// Wait for output to complete streaming.
			select {
			case err = <-outputDone:
				return errors.WithStack(err)
			case <-ctx.Done():
				return errors.WithStack(ctx.Err())
			}
		}
		return nil
	case err = <-detached:
		// Got a detach key sequence.
		return errors.WithStack(err)
	case <-ctx.Done():
		return errors.WithStack(ctx.Err())
	}
}

func (h *HijackedIOStreamer) setupInput() (restore func(), err error) {
	if h.InputStream == nil || !h.TTY {
		// No need to setup input TTY.
		// The restore func is a nop.
		return func() {}, nil
	}

	if err := setRawTerminal(h.Streams); err != nil {
		return nil, errors.Errorf("unable to set IO streams as raw terminal: %s", err)
	}

	// Use sync.Once so we may call restore multiple times but ensure we
	// only restore the terminal once.
	var restoreOnce sync.Once
	restore = func() {
		restoreOnce.Do(func() {
			_ = restoreTerminal(h.Streams, h.InputStream)
		})
	}

	// Wrap the input to detect detach escape sequence.
	// Use default escape keys if an invalid sequence is given.
	escapeKeys := defaultEscapeKeys
	if h.DetachKeys != "" {
		customEscapeKeys, err := term.ToBytes(h.DetachKeys)
		if err != nil {
			logrus.Warnf("invalid detach escape keys, using default: %s", err)
		} else {
			escapeKeys = customEscapeKeys
		}
	}

	h.InputStream = ioutils.NewReadCloserWrapper(
		term.NewEscapeProxy(h.InputStream, escapeKeys),
		h.InputStream.Close,
	)

	return restore, nil
}

func (h *HijackedIOStreamer) beginOutputStream(restoreInput func()) <-chan error {
	if h.OutputStream == nil && h.ErrorStream == nil {
		// There is no need to copy output.
		return nil
	}

	outputDone := make(chan error)
	go func() {
		var err error

		// When TTY is ON, use regular copy
		if h.OutputStream != nil && h.TTY {
			_, err = io.Copy(h.OutputStream, h.Resp.Reader)
			restoreInput()
		} else {
			_, err = stdcopy.StdCopy(h.OutputStream, h.ErrorStream, h.Resp.Reader)
		}

		logrus.Debug("[hijack] End of stdout")

		if err != nil {
			logrus.Debugf("Error receiveStdout: %s", err)
		}

		outputDone <- errors.WithStack(err)
	}()

	return outputDone
}

var errInvalidWrite = errors.New("invalid write result")

func Copy(ctx context.Context, dst net.Conn, src io.Reader) (written int64, err error) {
	size := 32 * 1024
	buf := make([]byte, size)
	for {
		select {
		case <-ctx.Done():
			err = ctx.Err()
			return
		default:
		}
		nr, er := src.Read(buf)
		if nr > 0 {
			// docker container is stop check
			err = dst.SetReadDeadline(time.Now().Add(time.Second * 3))
			if err != nil {
				break
			}
			nw, ew := dst.Write(buf[0:nr])
			if nw < 0 || nr < nw {
				nw = 0
				if ew == nil {
					ew = errInvalidWrite
				}
			}
			written += int64(nw)
			if ew != nil {
				err = ew
				break
			}
			if nr != nw {
				err = io.ErrShortWrite
				break
			}
		}
		if er != nil {
			if er != io.EOF {
				err = er
			}
			break
		}
	}
	return
}

func (h *HijackedIOStreamer) beginInputStream(ctx context.Context, restoreInput func()) (doneC <-chan struct{}, detachedC <-chan error) {
	inputDone := make(chan struct{})
	detached := make(chan error)

	go func() {
		if h.InputStream != nil {
			_, err := Copy(ctx, h.Resp.Conn, h.InputStream)
			restoreInput()

			logrus.Debug("[hijack] End of stdin")

			if _, ok := err.(term.EscapeError); ok {
				detached <- errors.WithStack(err)
				return
			}

			if err != nil {
				logrus.Debugf("Error sendStdin: %s", err)
			}
		}

		if err := h.Resp.CloseWrite(); err != nil {
			logrus.Debugf("Couldn't send EOF: %s", err)
		}

		close(inputDone)
	}()

	return inputDone, detached
}

func setRawTerminal(streams stream.Streams) error {
	if err := streams.In().SetRawTerminal(); err != nil {
		return errors.WithStack(err)
	}
	return streams.Out().SetRawTerminal()
}

func restoreTerminal(streams stream.Streams, in io.Closer) error {
	_ = streams.In().RestoreTerminal()
	_ = streams.Out().RestoreTerminal()
	if in != nil && runtime.GOOS != "darwin" && runtime.GOOS != "windows" {
		return in.Close()
	}
	return nil
}
