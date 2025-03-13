package main

import (
	"minim/cmd"

	"github.com/jxskiss/mcli"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// cmd.Setup()
	cmd.Init()

	mcli.Add("status", cmd.CmdStatus, "View the status")

	mcli.AddGroup("server", "Commands for managing Minimalytics server")
	mcli.Add("server start", cmd.CmdServerStart, "Start the server")
	mcli.Add("server stop", cmd.CmdServerStop, "Stop the server")
	mcli.Add("server restart", cmd.CmdServerRestart, "Restart the server")

	mcli.AddGroup("web", "Commands for managing the web UI for Minimalytics")
	mcli.Add("web enable", cmd.CmdUiEnable, "Enable the Minim UI")
	mcli.Add("web disable", cmd.CmdUiDisable, "Enable the Minim UI")

	mcli.AddHidden("execserver", cmd.CmdExecServer, "")

	mcli.Run()
}
