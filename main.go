package main

import (
	"minim/cmd"

	"github.com/jxskiss/mcli"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	mcli.Add("status", cmd.CmdStatus, "View the status")

	mcli.AddGroup("server", "Commands for managing Minimalytics server")
	mcli.Add("server start", cmd.CmdServerStart, "Start the server")
	mcli.Add("server stop", cmd.CmdServerStop, "Stop the server")

	mcli.AddHidden("execserver", cmd.CmdExecServer, "")

	mcli.Run()
}
