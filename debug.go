// +build !release

package main

import "github.com/bugsnag/bugsnag-go"

func init() {
	bugsnag.Configure(bugsnag.Configuration{
		APIKey:       "9d4e8ee05e1f9501a5b09ad982c5fc7f",
		ReleaseStage: "development",
	})
}
