package tty

import (
	"context"
	"fmt"
	"os"
	gosignal "os/signal"
	"runtime"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/zeromake/moby/api/types"
	"github.com/zeromake/moby/client"
	"github.com/zeromake/moby/pkg/signal"
	"github.com/zeromake/docker-debug/pkg/stream"
)

// ResizeTtyTo resizes tty to specific height and width
func ResizeTtyTo(ctx context.Context, client client.ContainerAPIClient, id string, height, width uint, isExec bool) {
	if height == 0 && width == 0 {
		return
	}

	options := types.ResizeOptions{
		Height: height,
		Width:  width,
	}

	var err error
	if isExec {
		err = client.ContainerExecResize(ctx, id, options)
	} else {
		err = client.ContainerResize(ctx, id, options)
	}

	if err != nil {
		// logrus.Debugf("Error resize: %s", err)
	}
}

// MonitorTtySize updates the container tty size when the terminal tty changes size
func MonitorTtySize(ctx context.Context, client client.ContainerAPIClient, out *stream.OutStream, id string, isExec bool) error {
	resizeTty := func() {
		height, width := out.GetTtySize()
		ResizeTtyTo(ctx, client, id, height, width, isExec)
	}

	resizeTty()

	if runtime.GOOS == "windows" {
		go func() {
			prevH, prevW := out.GetTtySize()
			for {
				time.Sleep(time.Millisecond * 250)
				h, w := out.GetTtySize()

				if prevW != w || prevH != h {
					resizeTty()
				}
				prevH = h
				prevW = w
			}
		}()
	} else {
		sigchan := make(chan os.Signal, 1)
		gosignal.Notify(sigchan, signal.SIGWINCH)
		go func() {
			for range sigchan {
				resizeTty()
			}
		}()
	}
	return nil
}
func ForwardAllSignals(ctx context.Context, client client.ContainerAPIClient, cid string) chan os.Signal {
	sigc := make(chan os.Signal, 128)
	signal.CatchAll(sigc)
	go func() {
		for s := range sigc {
			if s == signal.SIGCHLD || s == signal.SIGPIPE {
				continue
			}
			var sig string
			for sigStr, sigN := range signal.SignalMap {
				if sigN == s {
					sig = sigStr
					break
				}
			}
			if sig == "" {
				fmt.Printf("Unsupported signal: %v. Discarding.\n", s)
				continue
			}

			if err := client.ContainerKill(ctx, cid, sig); err != nil {
				logrus.Debugf("Error sending signal: %s", err)
			}
		}
	}()
	return sigc
}
