package exit

import (
	"context"
	"net/http"
	"os"
)

const (
	// ExitCodeSuccess = exit code 0
	ExitCodeSuccess = 0
	// ExitCodeError = exit code 1
	ExitCodeError = 1
	// ExitCodeRestart = exit code 5
	ExitCodeRestart = 5
)

var (
	IsShared   = false
	Server     *http.Server
)

func Exit(code int) {
	if Server != nil {
		Server.Shutdown(context.Background())
	}

	if IsShared {
		return
	}

	os.Exit(code)
}
