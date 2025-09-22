package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/nextsurfer/book-manage-api/internal/routes"
	"github.com/nextsurfer/book-manage-api/internal/tools"
)

func init() {
	pwd, _ := os.Getwd()
	log.Println("Current working directory: ", pwd)
	tools.InitConfig(pwd + "/conf.json")
	if !tools.Config().Debug {
		// log.Println("Config: gin.SetMode(gin.ReleaseMode)")
		gin.SetMode(gin.ReleaseMode)
	} else {
		// log.Println("Config: gin.SetMode(gin.DebugMode)")
		gin.SetMode(gin.DebugMode)
	}
	tools.InitCommonTools()
	tools.InitLevelToml(pwd + "/level.toml")
}

func main() {

	routes.Run()
}
