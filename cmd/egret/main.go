// The command line tool for running Egret apps.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/agtorre/gocolorize"
	flags "github.com/jessevdk/go-flags"
	"github.com/kenorld/egret/cmd/model"
	"github.com/kenorld/egret/cmd/utils"
	"go.uber.org/zap"
)

// Command structure cribbed from the genius organization of the "go" command.
type Command struct {
	UpdateConfig           func(c *model.CommandConfig, args []string) bool
	RunWith                func(c *model.CommandConfig) error
	UsageLine, Short, Long string
}

func (cmd *Command) Name() string {
	name := cmd.UsageLine
	i := strings.Index(name, " ")
	if i >= 0 {
		name = name[:i]
	}
	return name
}

var Commands = []*Command{
	cmdNew,
	cmdRun,
	cmdBuild,
	cmdPackage,
	cmdTest,
	cmdVersion,
}

func main() {
	if runtime.GOOS == "windows" {
		gocolorize.SetPlain(true)
	}
	c := &model.CommandConfig{}
	wd, _ := os.Getwd()

	fmt.Fprintf(os.Stdout, gocolorize.NewColor("blue").Paint(header))
	parser := flags.NewParser(c, flags.HelpFlag|flags.PassDoubleDash)
	if len(os.Args) < 2 {
		parser.WriteHelp(os.Stdout)
		os.Exit(1)
	}

	if err := ParseArgs(c, parser, os.Args[1:]); err != nil {
		fmt.Fprint(os.Stderr, err.Error()+"\n")
		os.Exit(1)
	}

	if err := c.UpdateImportPath(); err != nil {
		utils.Logger.Error(err.Error())
		parser.WriteHelp(os.Stdout)
		os.Exit(1)
	}

	command := Commands[c.Index]
	println("Revel executing:", command.Short)

	// Setting go paths
	c.InitGoPaths()

	// Setup package resolver
	c.InitPackageResolver()

	if err := command.RunWith(c); err != nil {
		utils.Logger.Error("Unable to execute", zap.Error(err))
		os.Exit(1)
	}
}

// Parse the arguments passed into the model.CommandConfig
func ParseArgs(c *model.CommandConfig, parser *flags.Parser, args []string) (err error) {
	var extraArgs []string
	if ini := flag.String("ini", "none", ""); *ini != "none" {
		if err = flags.NewIniParser(parser).ParseFile(*ini); err != nil {
			return
		}
	} else {
		if extraArgs, err = parser.ParseArgs(args); err != nil {
			return
		} else {
			switch parser.Active.Name {
			case "new":
				c.Index = model.NEW
			case "run":
				c.Index = model.RUN
			case "build":
				c.Index = model.BUILD
			case "package":
				c.Index = model.PACKAGE
			case "clean":
				c.Index = model.CLEAN
			case "test":
				c.Index = model.TEST
			case "version":
				c.Index = model.VERSION
			}
		}
	}

	if len(extraArgs) > 0 {
		utils.Logger.Info("Found additional arguements, setting them")
		if !Commands[c.Index].UpdateConfig(c, extraArgs) {
			buffer := &bytes.Buffer{}
			parser.WriteHelp(buffer)
			err = fmt.Errorf("Invalid command line arguements %v\n%s", extraArgs, buffer.String())
		}
	}

	return
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func errorf(format string, args ...interface{}) {
	// Ensure the user's command prompt starts on the next line.
	if !strings.HasSuffix(format, "\n") {
		format += "\n"
	}
	fmt.Fprintf(os.Stderr, format, args...)
	panic(LoggedError{}) // Panic instead of os.Exit so that deferred will run.
}

//http://patorjk.com/software/taag/#p=testall&h=1&c=c&f=Graffiti&t=Egret
const header = `
    U _____ u   ____     ____    U _____ u  _____   
    \| ___"|/U /"___|uU |  _"\ u \| ___"|/ |_ " _|  
     |  _|"  \| |  _ / \| |_) |/  |  _|"     | |    
     | |___   | |_| |   |  _ <    | |___    /| |\   
     |_____|   \____|   |_| \_\   |_____|  u |_|U   
     <<   >>   _)(|_    //   \\_  <<   >>  _// \\_  
    (__) (__) (__)__)  (__)  (__)(__) (__)(__) (__) 
		
		                                             
`
