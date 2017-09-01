package cmd

import (
	"github.com/mobingi/alm-agent/metavars"
	"github.com/stvp/rollbar"
)

func sendReport(err error) {
	if metavars.ReportDisabled {
		return
	}

	// report async.
	rollbar.Error(rollbar.ERR, err)
	rollbar.Wait()
	return
}
