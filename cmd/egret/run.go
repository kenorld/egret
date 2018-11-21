package main

import (
	"go/build"
	"os"
	"path/filepath"
	"strconv"

	"github.com/kenorld/egret"
	"github.com/kenorld/egret/cmd/harness"
	"github.com/kenorld/egret/cmd/model"
	"github.com/kenorld/egret/cmd/utils"
	"go.uber.org/zap"
)

var cmdRun = &Command{
	UsageLine: "run [import path] [run mode] [port]",
	Short:     "run a Egret application",
	Long: `
Run the Egret web application named by the given import path.

For example, to run the chat room sample application:

    egret run github.com/kenorld/egret/samples/chat dev

The run mode is used to select which set of app.yaml configuration should
apply and may be used to determine logic in the application itself.

Run mode defaults to "dev".

You can set a port as an optional third parameter.  For example:

    egret run github.com/kenorld/egret/samples/chat prod 8080`,
}

func init() {
	cmdRun.Run = runApp
	cmdRun.RunWith = runApp
	cmdRun.UpdateConfig = updateRunConfig
}

func updateRunConfig(c *model.CommandConfig, args []string) bool {
	convertPort := func(value string) int {
		if value != "" {
			port, err := strconv.Atoi(value)
			if err != nil {
				utils.Logger.Fatalf("Failed to parse port as integer: %s", c.Run.Port)
			}
			return port
		}
		return 0
	}
	switch len(args) {
	case 3:
		// Possible combinations
		// revel run [import-path] [run-mode] [port]
		c.Run.ImportPath = args[0]
		c.Run.Mode = args[1]
		c.Run.Port = convertPort(args[2])
	case 2:
		// Possible combinations
		// 1. revel run [import-path] [run-mode]
		// 2. revel run [import-path] [port]
		// 3. revel run [run-mode] [port]

		// Check to see if the import path evaluates out to something that may be on a gopath
		if runIsImportPath(args[0]) {
			// 1st arg is the import path
			c.Run.ImportPath = args[0]

			if _, err := strconv.Atoi(args[1]); err == nil {
				// 2nd arg is the port number
				c.Run.Port = convertPort(args[1])
			} else {
				// 2nd arg is the run mode
				c.Run.Mode = args[1]
			}
		} else {
			// 1st arg is the run mode
			c.Run.Mode = args[0]
			c.Run.Port = convertPort(args[1])
		}
	case 1:
		// Possible combinations
		// 1. revel run [import-path]
		// 2. revel run [port]
		// 3. revel run [run-mode]
		if runIsImportPath(args[0]) {
			// 1st arg is the import path
			c.Run.ImportPath = args[0]
		} else if _, err := strconv.Atoi(args[0]); err == nil {
			// 1st arg is the port number
			c.Run.Port = convertPort(args[0])
		} else {
			// 1st arg is the run mode
			c.Run.Mode = args[0]
		}
	case 0:
		// Attempt to set the import path to the current working director.
		c.Run.ImportPath, _ = os.Getwd()
	}
	c.Index = model.RUN
	return true
}

// Returns true if this is an absolute path or a relative gopath
func runIsImportPath(pathToCheck string) bool {
	if _, err := build.Import(pathToCheck, "", build.FindOnly); err == nil {
		return true
	}
	return filepath.IsAbs(pathToCheck)
}

func runApp(c *model.CommandConfig) {
	if c.Run.Mode == "" {
		c.Run.Mode = "dev"
	}

	// Find and parse app.yaml
	egret.Init(mode, args[0], "")
	egret.LoadMimeConfig()

	// Determine the override port, if any.
	port := egret.HttpPort
	if len(args) == 3 {
		var err error
		if port, err = strconv.Atoi(args[2]); err != nil {
			errorf("Failed to parse port as integer: %s", args[2])
		}
	}

	utils.Logger.Info("Running...",
		zap.String("AppName", egret.AppName),
		zap.String("ImportPath", egret.ImportPath),
		zap.String("Mode", mode),
		zap.String("BasePath", egret.BasePath),
	)

	// If the app is run in "watched" mode, use the harness to run it.
	if egret.Config.GetBoolDefault("watch.enabled", true) && egret.Config.GetBoolDefault("watch.code", true) {
		utils.Logger.Info("Running in watched mode.")
		egret.HttpPort = port
		harness.NewHarness().Run() // Never returns.
	}

	// Else, just build and run the app.
	utils.Logger.Info("Running in live build mode.")
	app, err := harness.Build()
	if err != nil {
		errorf("Failed to build app: %s", err)
	}
	app.Port = port

	app.Cmd().Run()
}
