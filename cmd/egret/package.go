package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/kenorld/egret"
	"github.com/kenorld/egret/cmd/model"
)

var cmdPackage = &Command{
	UsageLine: "package [import path] [run mode]",
	Short:     "package a Egret application (e.g. for deployment)",
	Long: `
Package the Egret web application named by the given import path.
This allows it to be deployed and run on a machine that lacks a Go installation.

The run mode is used to select which set of app.yaml configuration should
apply and may be used to determine logic in the application itself.

Run mode defaults to "dev".

For example:

    egret package github.com/kenorld/egret/samples/chat
`,
}

func init() {
	cmdPackage.RunWith = packageApp
	cmdPackage.UpdateConfig = updatePackageConfig
}

// Called when unable to parse the command line automatically and assumes an old launch
func updatePackageConfig(c *model.CommandConfig, args []string) bool {
	c.Index = model.PACKAGE
	c.Package.ImportPath = args[0]
	if len(args) > 1 {
		c.Package.Mode = args[1]
	}
	return true
}

func packageApp(args []string) {
	if len(args) == 0 {
		fmt.Fprint(os.Stderr, cmdPackage.Long)
		return
	}

	// Determine the run mode.
	mode := "dev"
	if len(args) >= 2 {
		mode = args[1]
	}

	appImportPath := args[0]
	egret.Init(mode, appImportPath, "")

	// Remove the archive if it already exists.
	destFile := filepath.Base(egret.BasePath) + ".tar.gz"
	os.Remove(destFile)

	// Collect stuff in a temp directory.
	tmpDir, err := ioutil.TempDir("", filepath.Base(egret.BasePath))
	panicOnError(err, "Failed to get temp dir")

	buildApp([]string{args[0], tmpDir, mode})

	// Create the zip file.
	archiveName := mustTarGzDir(destFile, tmpDir)

	fmt.Println("Your archive is ready:", archiveName)
}
