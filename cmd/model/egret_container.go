// This package will be shared between Egret and Egret CLI eventually
package model

import (
	"go/build"

	"github.com/kenorld/egret/cmd/utils"
	"github.com/kenorld/egret/conf"

	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

type (
	// The container object for describing all Egrets variables
	EgretContainer struct {
		BuildPaths struct {
			Egret string
		}
		Paths struct {
			Import   string
			Source   string
			Base     string
			App      string
			Views    string
			Code     []string
			Template []string
			Config   []string
		}
		PackageInfo struct {
			Config   conf.Context
			Packaged bool
			DevMode  bool
			Vendor   bool
		}
		Application struct {
			Name string
			Root string
		}

		ImportPath    string            // The import path
		SourcePath    string            // The full source path
		RunMode       string            // The current run mode
		EgretPath     string            // The path to the Egret source code
		BasePath      string            // The base path to the application
		AppPath       string            // The application path (BasePath + "/app")
		ViewsPath     string            // The application views path
		CodePaths     []string          // All the code paths
		TemplatePaths []string          // All the template paths
		ConfPaths     []string          // All the configuration paths
		Config        *conf.Context     // The global config object
		Packaged      bool              // True if packaged
		DevMode       bool              // True if running in dev mode
		HTTPPort      int               // The http port
		HTTPAddr      string            // The http address
		HTTPSsl       bool              // True if running https
		HTTPSslCert   string            // The SSL certificate
		HTTPSslKey    string            // The SSL key
		AppName       string            // The application name
		AppRoot       string            // The application root from the config `app.root`
		CookiePrefix  string            // The cookie prefix
		CookieDomain  string            // The cookie domain
		CookieSecure  bool              // True if cookie is secure
		SecretStr     string            // The secret string
		MimeConfig    *conf.Context     // The mime configuration
		ModulePathMap map[string]string // The module path map
	}

	WrappedEgretCallback struct {
		FireEventFunction func(key Event, value interface{}) (response EventResponse)
		ImportFunction    func(pkgName string) error
	}
)

// Simple Wrapped EgretCallback
func NewWrappedEgretCallback(fe func(key Event, value interface{}) (response EventResponse), ie func(pkgName string) error) EgretCallback {
	return &WrappedEgretCallback{fe, ie}
}

// Function to implement the FireEvent
func (w *WrappedEgretCallback) FireEvent(key Event, value interface{}) (response EventResponse) {
	if w.FireEventFunction != nil {
		response = w.FireEventFunction(key, value)
	}
	return
}
func (w *WrappedEgretCallback) PackageResolver(pkgName string) error {
	return w.ImportFunction(pkgName)
}

// EgretImportPath Egret framework import path
var EgretImportPath = "github.com/kenorld/egret/egret"
var EgretModulesImportPath = "github.com/kenorld/egret/modules"

// This function returns a container object describing the egret application
// eventually this type of function will replace the global variables.
func NewEgretPaths(mode, importPath, srcPath string, callback EgretCallback) (rp *EgretContainer, err error) {
	rp = &EgretContainer{ModulePathMap: map[string]string{}}
	// Ignore trailing slashes.
	rp.ImportPath = strings.TrimRight(importPath, "/")
	rp.SourcePath = srcPath
	rp.RunMode = mode

	// If the SourcePath is not specified, find it using build.Import.
	var egretSourcePath string // may be different from the app source path
	if rp.SourcePath == "" {
		rp.SourcePath, egretSourcePath, err = utils.FindSrcPaths(importPath, EgretImportPath, callback.PackageResolver)
		if err != nil {
			return
		}
	} else {
		// If the SourcePath was specified, assume both Egret and the app are within it.
		rp.SourcePath = filepath.Clean(rp.SourcePath)
		egretSourcePath = rp.SourcePath
	}

	// Setup paths for application
	rp.EgretPath = filepath.Join(egretSourcePath, filepath.FromSlash(EgretImportPath))
	rp.BasePath = filepath.Join(rp.SourcePath, filepath.FromSlash(importPath))
	rp.PackageInfo.Vendor = utils.Exists(filepath.Join(rp.BasePath, "vendor"))
	rp.AppPath = filepath.Join(rp.BasePath, "app")

	// Sanity check , ensure app and conf paths exist
	if !utils.DirExists(rp.AppPath) {
		return rp, fmt.Errorf("No application found at path %s", rp.AppPath)
	}
	if !utils.DirExists(filepath.Join(rp.BasePath, "conf")) {
		return rp, fmt.Errorf("No configuration found at path %s", filepath.Join(rp.BasePath, "conf"))
	}

	rp.ViewsPath = filepath.Join(rp.AppPath, "views")
	rp.CodePaths = []string{rp.AppPath}
	rp.TemplatePaths = []string{}

	if rp.ConfPaths == nil {
		rp.ConfPaths = []string{}
	}

	// Config load order
	// 1. framework (egret/conf/*)
	// 2. application (conf/*)
	// 3. user supplied configs (...) - User configs can override/add any from above
	rp.ConfPaths = append(
		[]string{
			filepath.Join(rp.EgretPath, "conf"),
			filepath.Join(rp.BasePath, "conf"),
		},
		rp.ConfPaths...)

	rp.Config, err = conf.LoadContext("app.conf", rp.ConfPaths)
	if err != nil {
		return rp, fmt.Errorf("Unable to load configuartion file %s", err)
	}

	// Ensure that the selected runmode appears in app.conf.
	// If empty string is passed as the mode, treat it as "DEFAULT"
	if mode == "" {
		mode = "dev"
	}
	if !rp.Config.HasSection(mode) {
		return rp, fmt.Errorf("app.conf: No mode found: %s %s", "run-mode", mode)
	}

	rp.Config.SetSection(mode)

	// Configure properties from app.conf
	rp.DevMode = rp.Config.GetBoolDefault("mode.dev", false)
	rp.HTTPPort = rp.Config.GetIntDefault("http.port", 9000)
	rp.HTTPAddr = rp.Config.GetStringDefault("http.addr", "")
	rp.HTTPSsl = rp.Config.GetBoolDefault("http.ssl", false)
	rp.HTTPSslCert = rp.Config.GetStringDefault("http.sslcert", "")
	rp.HTTPSslKey = rp.Config.GetStringDefault("http.sslkey", "")
	if rp.HTTPSsl {
		if rp.HTTPSslCert == "" {
			return rp, errors.New("No http.sslcert provided.")
		}
		if rp.HTTPSslKey == "" {
			return rp, errors.New("No http.sslkey provided.")
		}
	}
	//
	rp.AppName = rp.Config.GetStringDefault("app.name", "(not set)")
	rp.AppRoot = rp.Config.GetStringDefault("app.root", "")
	rp.CookiePrefix = rp.Config.GetStringDefault("cookie.prefix", "REVEL")
	rp.CookieDomain = rp.Config.GetStringDefault("cookie.domain", "")
	rp.CookieSecure = rp.Config.GetBoolDefault("cookie.secure", rp.HTTPSsl)
	rp.SecretStr = rp.Config.GetStringDefault("app.secret", "")

	callback.FireEvent(REVEL_BEFORE_MODULES_LOADED, nil)
	if err := rp.loadModules(callback); err != nil {
		return rp, err
	}

	callback.FireEvent(REVEL_AFTER_MODULES_LOADED, nil)

	return
}

// LoadMimeConfig load mime-types.conf on init.
func (rp *EgretContainer) LoadMimeConfig() (err error) {
	rp.MimeConfig, err = conf.LoadContext("mime-types.conf", rp.ConfPaths)
	if err != nil {
		return fmt.Errorf("Failed to load mime type config: %s %s", "error", err)
	}
	return
}

// Loads modules based on the configuration setup.
// This will fire the REVEL_BEFORE_MODULE_LOADED, REVEL_AFTER_MODULE_LOADED
// for each module loaded. The callback will receive the EgretContainer, name, moduleImportPath and modulePath
// It will automatically add in the code paths for the module to the
// container object
func (rp *EgretContainer) loadModules(callback EgretCallback) (err error) {
	// for _, key := range rp.Config.GetStringMapString("module") {
	// 	moduleImportPath := rp.Config.GetStringDefault(key, "")
	// 	if moduleImportPath == "" {
	// 		continue
	// 	}

	// 	modulePath, err := ResolveImportPath(moduleImportPath)
	// 	if err != nil {
	// 		log.Fatalln("Failed to load module. Import of", moduleImportPath, "failed:", err)
	// 		addModule(key[len("module."):], moduleImportPath, modulePath)
	// 	}
	// 	// callback.FireEvent(REVEL_BEFORE_MODULE_LOADED, []interface{}{rp, name, moduleImportPath, modulePath})
	// 	// rp.addModulePaths(name, moduleImportPath, modulePath)
	// 	// callback.FireEvent(REVEL_AFTER_MODULE_LOADED, []interface{}{rp, name, moduleImportPath, modulePath})
	// }

	return
}

// Adds a module paths to the container object
func (rp *EgretContainer) addModulePaths(name, importPath, modulePath string) {
	if codePath := filepath.Join(modulePath, "app"); utils.DirExists(codePath) {
		rp.CodePaths = append(rp.CodePaths, codePath)
		rp.ModulePathMap[name] = modulePath
		if viewsPath := filepath.Join(modulePath, "app", "views"); utils.DirExists(viewsPath) {
			rp.TemplatePaths = append(rp.TemplatePaths, viewsPath)
		}
	}

	// Hack: There is presently no way for the testrunner module to add the
	// "test" subdirectory to the CodePaths.  So this does it instead.
	if importPath == rp.Config.GetStringDefault("module.testrunner", "github.com/kenorld/egret/modules/testrunner") {
		joinedPath := filepath.Join(rp.BasePath, "tests")
		rp.CodePaths = append(rp.CodePaths, joinedPath)
	}
	if testsPath := filepath.Join(modulePath, "tests"); utils.DirExists(testsPath) {
		rp.CodePaths = append(rp.CodePaths, testsPath)
	}
}

// ResolveImportPath returns the filesystem path for the given import path.
// Returns an error if the import path could not be found.
func (rp *EgretContainer) ResolveImportPath(importPath string) (string, error) {
	if rp.Packaged {
		return filepath.Join(rp.SourcePath, importPath), nil
	}

	modPkg, err := build.Import(importPath, rp.AppPath, build.FindOnly)
	if err != nil {
		return "", err
	}
	if rp.PackageInfo.Vendor && !strings.HasPrefix(modPkg.Dir, rp.BasePath) {
		return "", fmt.Errorf("Module %s was found outside of path %s.", importPath, modPkg.Dir)
	}
	return modPkg.Dir, nil
}
