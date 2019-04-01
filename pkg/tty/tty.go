package tty

import (
	"context"
	"github.com/sirupsen/logrus"
	"os"
	goSignal "os/signal"
	"runtime"
	"time"

	"github.com/zeromake/docker-debug/pkg/stream"
	"github.com/zeromake/moby/api/types"
	"github.com/zeromake/moby/client"
	"github.com/zeromake/moby/pkg/signal"
)

// ResizeTtyTo re sizes tty to specific height and width
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
		logrus.Debugf("Error resize: %s", err)
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
		sigChan := make(chan os.Signal, 1)
		goSignal.Notify(sigChan, signal.SIGWINCH)
		go func() {
			for range sigChan {
				resizeTty()
			}
		}()
	}
	return nil
}
