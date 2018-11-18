package main

import (
	"flag"

	"github.com/kenorld/egret"
	"github.com/kenorld/xlh-server/core/routes"
)

var (
	runMode    *string = flag.String("runMode", "dev", "Run mode.")
	port       *int    = flag.Int("port", 4411, "By default, read from app.conf")
	importPath *string = flag.String("importPath", "github.com/kenorld/xlh-server", "Go Import Path for the app.")
	srcPath    *string = flag.String("srcPath", "", "Path to the source root.")
)

func main() {
	flag.Parse()
	egret.Init(*runMode, *importPath, *srcPath)
	// DB Main
	//DbMain()
	routes.Register()

	// start the server
	egret.Serve(*port)
}

// func DbMain() {
// 	// Database Main Conexion
// 	Db := db.MgoDb{}
// 	Db.Init()
// 	// index keys
// 	keys := []string{"email"}
// 	Db.Index("auth", keys)
// }
