package main

import (
	"github.com/m1khal3v/gophkeeper/internal/client/app"
	"github.com/m1khal3v/gophkeeper/internal/common/buildlog"
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

	app.Run()
}
