// +build windows

package opts

// DefaultHost constant defines the default host string used by docker on Windows
var DefaultHost = "npipe://" + DefaultNamedPipe
var DefaultHTTPHost = "127.0.0.1"
