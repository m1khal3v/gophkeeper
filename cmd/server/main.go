package main

import (
	"github.com/m1khal3v/gophkeeper/internal/common/buildlog"
	"github.com/m1khal3v/gophkeeper/internal/server/app"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func main() {
	buildlog.Print(buildVersion, buildDate, buildCommit)

	app, err := app.New()
	if err != nil {
		panic(err)
	}

	err = app.Run()
	if err != nil {
		panic(err)
	}
}
