package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/kenorld/egret"
	"github.com/kenorld/egret/cmd/harness"
	"github.com/kenorld/egret/cmd/model"
)

var cmdBuild = &Command{
	UsageLine: "build [import path] [target path] [run mode]",
	Short:     "build a Egret application (e.g. for deployment)",
	Long: `
Build the Egret web application named by the given import path.
This allows it to be deployed and run on a machine that lacks a Go installation.

The run mode is used to select which set of app.yaml configuration should
apply and may be used to determine logic in the application itself.

Run mode defaults to "dev".

WARNING: The target path will be completely deleted, if it already exists!

For example:

    egret build github.com/kenorld/egret/samples/chat /tmp/chat
`,
}

func init() {
	cmdBuild.RunWith = buildApp
	cmdBuild.UpdateConfig = updateBuildConfig
}

// The update config updates the configuration command so that it can run
func updateBuildConfig(c *model.CommandConfig, args []string) bool {
	c.Index = model.BUILD
	// If arguments were passed in then there must be two
	if len(args) < 2 {
		fmt.Fprintf(os.Stderr, "%s\n%s", cmdBuild.UsageLine, cmdBuild.Long)
		return false
	}

	c.Build.ImportPath = args[0]
	c.Build.TargetPath = args[1]
	if len(args) > 2 {
		c.Build.Mode = args[2]
	}
	return true
}
func buildApp(c *model.CommandConfig) {
	appImportPath, destPath, mode := c.ImportPath, c.Build.TargetPath, "dev"
	if len(c.Build.Mode) > 0 {
		mode = c.Build.Mode
	}

	c.Build.TargetPath, _ = filepath.Abs(destPath)
	c.Build.Mode = mode
	c.Build.ImportPath = appImportPath

	if !egret.Initialized {
		egret.Init(mode, appImportPath, "")
	}

	// First, verify that it is either already empty or looks like a previous
	// build (to avoid clobbering anything)
	if exists(destPath) && !empty(destPath) && !exists(path.Join(destPath, "run.sh")) {
		errorf("Abort: %s exists and does not look like a build directory.", destPath)
	}

	os.RemoveAll(destPath)
	srcPath := path.Join(destPath, "src")
	mustCopyDir(path.Join(srcPath, filepath.FromSlash(appImportPath)), egret.BasePath, true, nil)
	os.MkdirAll(destPath, 0777)

	app, eerr := harness.Build()
	panicOnError(eerr, "Failed to build")

	// Included are:
	// - run scripts
	// - binary
	// - egret
	// - app

	// Egret and the app are in a directory structure mirroring import path
	destBinaryPath := path.Join(destPath, filepath.Base(app.BinaryPath))
	tmpEgretPath := path.Join(srcPath, filepath.FromSlash(egret.EgretImportPath))
	mustCopyFile(destBinaryPath, app.BinaryPath)
	mustChmod(destBinaryPath, 0755)
	mustCopyDir(path.Join(tmpEgretPath, "conf"), path.Join(egret.EgretPath, "core", "conf"), true, nil)
	mustCopyDir(path.Join(tmpEgretPath, "views"), path.Join(egret.EgretPath, "core", "views"), true, nil)

	tmplData, runShPath := map[string]interface{}{
		"BinName":    filepath.Base(app.BinaryPath),
		"ImportPath": appImportPath,
		"Mode":       mode,
	}, path.Join(destPath, "run.sh")

	mustRenderTemplate(
		runShPath,
		filepath.Join(egret.EgretPath, "cmd", "egret", "package_run.sh.template"),
		tmplData)

	mustChmod(runShPath, 0755)

	mustRenderTemplate(
		filepath.Join(destPath, "run.bat"),
		filepath.Join(egret.EgretPath, "cmd", "egret", "package_run.bat.template"),
		tmplData)
}
