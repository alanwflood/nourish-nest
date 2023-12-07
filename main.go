package main

import (
	"NourishNestApp/db"
	"NourishNestApp/handlers"
	"NourishNestApp/logger"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
)

func main() {
	loadLocalEnvVariables()

	db.Init()
	e := echo.New()
	logger.SetupLogging(e)
	handlers.UseSessionMiddleware(e)
	handlers.HandleRoutes(e)

	staticFilesDir := os.Getenv("STATIC_FILES_DIR")
	logger.Log.Info(fmt.Sprintf("Serving static files from directory: '%s'", staticFilesDir))
	e.Static("/public", staticFilesDir)

	port := os.Getenv("PORT")
	logger.Log.Fatal(e.Start(port).Error())
}

func loadLocalEnvVariables() {
	err := godotenv.Load()
	if err != nil {
		logger.Log.Fatal("Error loading .env file")
	}
}
