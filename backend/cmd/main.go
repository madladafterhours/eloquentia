package main

import (
	"database/sql"
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

type ProfileParams struct {
	Name           string
	NativeLanguage string
	TargetLanguage string
}

func main() {
	err := Run(NewRouter())
	if err != nil {
		log.Fatal(err)
	}
}

func NewRouter() *gin.Engine {
	r := gin.Default()
	r.GET("/health", Health)
	r.POST("/profile/create", CreateProfile)
	return r
}

func Run(r *gin.Engine) error {
	portEnv := os.Getenv("ELOQUENTIA_PORT")

	if portEnv == "" {
		portEnv = "8090"
	}

	port := flag.String("port", portEnv, "Port for the API")

	flag.Parse()
	return r.Run("127.0.0.1:" + *port)
}

func CreateProfile(c *gin.Context) {
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		c.JSON(500, gin.H{"error": "Unable to find user config dir"})
	}

	dbPath := filepath.Join(userConfigDir, "eloquentia.db")
	_, err = os.Stat(dbPath)
	if errors.Is(err, os.ErrNotExist) {
		os.WriteFile(dbPath, nil, 0644)
	} else if err != nil {
		c.JSON(500, gin.H{"error": "Unable to open database file. Check permissions."})
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		c.JSON(500, gin.H{"error": "Unable to open database"})
	}
	defer db.Close()

	_, err = db.Exec(`create table if not exists Users (id integer primary key autoincrement,
		name varchar(255),
		native_language varchar(255),
		target_language(255))`)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed querying database"})
	}

	c.JSON(http.StatusOK, gin.H{"response": "Hello World!"})
}

func Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
