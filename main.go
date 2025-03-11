package main

import (
	"minim/cmd"
	"minim/model"

	"github.com/jxskiss/mcli"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// cmd.Setup()
	model.Init()

	mcli.Add("status", cmd.CmdStatus, "View the status")

	mcli.AddGroup("server", "Commands for managing Minimalytics server")
	mcli.Add("server start", cmd.CmdServerStart, "Start the server")
	mcli.Add("server stop", cmd.CmdServerStop, "Stop the server")
	mcli.Add("server restart", cmd.CmdServerRestart, "Restart the server")

	mcli.AddGroup("ui", "Commands for managing the web UI for Minimalytics")
	mcli.Add("ui enable", cmd.CmdUiEnable, "Enable the Minim UI")
	mcli.Add("ui disable", cmd.CmdUiDisable, "Enable the Minim UI")

	mcli.AddHidden("execserver", cmd.CmdExecServer, "")

	mcli.Run()
}
