// +build windows

package opts

import (
	"fmt"
	"golang.org/x/sys/windows/registry"
	"strings"
)

// DefaultHost constant defines the default host string used by docker on Windows
var DefaultHost = "npipe://" + DefaultNamedPipe
var DefaultHTTPHost = "127.0.0.1"

func init()  {
	var key, err = registry.OpenKey(registry.LOCAL_MACHINE, "SOFTWARE\\Microsoft\\Windows NT\\CurrentVersion", registry.QUERY_VALUE)
	if err != nil {
		return
	}
	version, _, err := key.GetStringValue("CurrentVersion")
	if err != nil {
		return
	}
	if strings.EqualFold(version, "6.1") {
		DefaultHTTPHost = "192.168.99.100"
		DefaultHost = fmt.Sprintf("tcp://%s:%d", DefaultHTTPHost, DefaultHTTPPort)
		IsWindows7 = true
		restHost()
	}
}

