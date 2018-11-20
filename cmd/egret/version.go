// Copyright (c) 2012-2016 The Egret Framework Authors, All rights reserved.
// Egret Framework source code and usage is governed by a MIT style
// license that can be found in the LICENSE file.

// Copyright (c) 2012-2016 The Egret Framework Authors, All rights reserved.
// Egret Framework source code and usage is governed by a MIT style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"

	"bytes"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/kenorld/egret"
	"github.com/kenorld/egret/cmd/model"
	"github.com/kenorld/egret/cmd/utils"
	"go.uber.org/zap"
)

type (
	// The version container
	VersionCommand struct {
		Command        *model.CommandConfig // The command
		egretVersion   *model.Version       // The Egret framework version
		modulesVersion *model.Version       // The Egret modules version
		cmdVersion     *model.Version       // The tool version
	}
)

var cmdVersion = &Command{
	UsageLine: "egret version",
	Short:     "displays the Egret Framework and Go version",
	Long: `
Displays the Egret Framework and Go version.

For example:

    egret version [<application path>]
`,
}

func init() {
	v := &VersionCommand{}
	cmdVersion.UpdateConfig = v.UpdateConfig
	cmdVersion.RunWith = v.RunWith
}

// Update the version
func (v *VersionCommand) UpdateConfig(c *model.CommandConfig, args []string) bool {
	if len(args) > 0 {
		c.Version.ImportPath = args[0]
	}
	return true
}

// Displays the version of go and Egret
func (v *VersionCommand) RunWith(c *model.CommandConfig) (err error) {
	utils.Logger.Info("Requesting version information", "config", c)
	v.Command = c

	// Update the versions with the local values
	v.updateLocalVersions()

	needsUpdates := true
	versionInfo := ""
	for x := 0; x < 2 && needsUpdates; x++ {
		needsUpdates = false
		versionInfo, needsUpdates = v.doRepoCheck(x == 0)
	}

	fmt.Println(versionInfo)
	cmd := exec.Command(c.GoCmd, "version")
	cmd.Stdout = os.Stdout
	if e := cmd.Start(); e != nil {
		fmt.Println("Go command error ", e)
	} else {
		cmd.Wait()
	}

	return
}

// Checks the Egret repos for the latest version
func (v *VersionCommand) doRepoCheck(updateLibs bool) (versionInfo string, needsUpdate bool) {
	var (
		title        string
		localVersion *model.Version
	)
	for _, repo := range []string{"egret"} {
		versonFromRepo, err := v.versionFromRepo(repo, "", "version.go")
		if err != nil {
			utils.Logger.Info("Failed to get version from repo", zap.String("repo", repo), zap.Error(err))
		}
		switch repo {
		case "egret":
			title, repo, localVersion = "Egret Framework", "github.com/kenorld/egret", v.egretVersion
		}

		// Only do an update on the first loop, and if specified to update
		shouldUpdate := updateLibs && v.Command.Version.Update
		if v.Command.Version.Update {
			if localVersion == nil || (versonFromRepo != nil && versonFromRepo.Newer(localVersion)) {
				needsUpdate = true
				if shouldUpdate {
					v.doUpdate(title, repo, localVersion, versonFromRepo)
					v.updateLocalVersions()
				}
			}
		}
		versionInfo = versionInfo + v.outputVersion(title, repo, localVersion, versonFromRepo)
	}
	return
}

// Checks for updates if needed
func (v *VersionCommand) doUpdate(title, repo string, local, remote *model.Version) {
	utils.Logger.Info("Updating package", zap.String("package", title), zap.String("repo", repo))
	fmt.Println("Attempting to update package", title)
	if err := v.Command.PackageResolver(repo); err != nil {
		utils.Logger.Error("Unable to update repo", zap.String("repo", repo), zap.Error(err))
	} else if repo == "github.com/kenorld/egret/cmd/egret" {
		// One extra step required here to run the install for the command
		utils.Logger.Fatal("Egret command tool was updated, you must manually run the following command before continuing\ngo install github.com/kenorld/egret/cmd/egret")
	}
	return
}

// Prints out the local and remote versions, calls update if needed
func (v *VersionCommand) outputVersion(title, repo string, local, remote *model.Version) (output string) {
	buffer := &bytes.Buffer{}
	remoteVersion := "Unknown"
	if remote != nil {
		remoteVersion = remote.VersionString()
	}
	localVersion := "Unknown"
	if local != nil {
		localVersion = local.VersionString()
	}

	fmt.Fprintf(buffer, "%s\t:\t%s\t(%s remote master branch)\n", title, localVersion, remoteVersion)
	return buffer.String()
}

// Returns the version from the repository
func (v *VersionCommand) versionFromRepo(repoName, branchName, fileName string) (version *model.Version, err error) {
	if branchName == "" {
		branchName = "master"
	}
	// Try to download the version of file from the repo, just use an http connection to retrieve the source
	// Assuming that the repo is github
	fullurl := "https://raw.githubusercontent.com/egret/" + repoName + "/" + branchName + "/" + fileName
	resp, err := http.Get(fullurl)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	utils.Logger.Info("Got version file", zap.String("from", fullurl), zap.String("content", string(body)))

	return v.versionFromBytes(body)
}

// Returns version information from a file called version on the gopath
func (v *VersionCommand) compareAndUpdateVersion(remoteVersion *model.Version, localVersion *model.Version) (err error) {
	return
}
func (v *VersionCommand) versionFromFilepath(sourcePath string) (version *model.Version, err error) {
	utils.Logger.Info("Fullpath to egret", zap.String("dir", sourcePath))

	sourceStream, err := ioutil.ReadFile(filepath.Join(sourcePath, "version.go"))
	if err != nil {
		return
	}
	return v.versionFromBytes(sourceStream)
}

// Returns version information from a file called version on the gopath
func (v *VersionCommand) versionFromBytes(sourceStream []byte) (version *model.Version, err error) {
	fset := token.NewFileSet() // positions are relative to fset

	// Parse src but stop after processing the imports.
	f, err := parser.ParseFile(fset, "", sourceStream, parser.ParseComments)
	if err != nil {
		err = utils.NewBuildError("Failed to parse Egret version error:", "error", err)
		return
	}
	version = &model.Version{}

	// Print the imports from the file's AST.
	for _, s := range f.Decls {
		genDecl, ok := s.(*ast.GenDecl)
		if !ok {
			continue
		}
		if genDecl.Tok != token.CONST {
			continue
		}
		for _, a := range genDecl.Specs {
			spec := a.(*ast.ValueSpec)
			r := spec.Values[0].(*ast.BasicLit)
			switch spec.Names[0].Name {
			case "Version":
				version.ParseVersion(strings.Replace(r.Value, `"`, "", -1))
			case "BuildDate":
				version.BuildDate = r.Value
			case "MinimumGoVersion":
				version.MinGoVersion = r.Value
			}
		}
	}
	return
}

// Fetch the local version of egret from the file system
func (v *VersionCommand) updateLocalVersions() {
	v.cmdVersion = &model.Version{}
	v.cmdVersion.ParseVersion(egret.Version)
	v.cmdVersion.BuildDate = egret.BuildDate
	v.cmdVersion.MinGoVersion = egret.MinimumGoVersion

	var modulePath, egretPath string
	_, egretPath, err := utils.FindSrcPaths(v.Command.ImportPath, model.EgretImportPath, v.Command.PackageResolver)
	if err != nil {
		utils.Logger.Warn("Unable to extract version information from Egret library", zap.Error(err))
		return
	}
	egretPath = egretPath + model.EgretImportPath
	utils.Logger.Info("Fullpath to egret", zap.String("dir", egretPath))
	v.egretVersion, err = v.versionFromFilepath(egretPath)
	if err != nil {
		utils.Logger.Warn("Unable to extract version information from Egret", zap.Error(err))
	}

	_, modulePath, err = utils.FindSrcPaths(v.Command.ImportPath, model.EgretModulesImportPath, v.Command.PackageResolver)
	if err != nil {
		utils.Logger.Warn("Unable to extract version information from Egret library", zap.Error(err))
		return
	}
	modulePath = modulePath + model.EgretModulesImportPath
	v.modulesVersion, err = v.versionFromFilepath(modulePath)
	if err != nil {
		utils.Logger.Warn("Unable to extract version information from Egret Modules", zap.Error(err))
	}

	return
}
