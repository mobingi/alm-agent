package cmd

import (
	"github.com/mobingi/alm-agent/metavars"
	"github.com/stvp/rollbar"
)

func sendReport(error) {
	if !metavars.ReportEnabled {
		return
	}

	// report async.
	rollbar.Error(rollbar.ERR, err)
	rollbar.Wait()
	return
}
