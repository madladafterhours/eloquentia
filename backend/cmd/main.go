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
	_ "github.com/glebarez/go-sqlite"
)

type ProfileParams struct {
	Name           string `json:"name"`
	NativeLanguage string `json:"native_language"`
	TargetLanguage string `json:"target_language"`
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
	var profileParams ProfileParams
	err := c.BindJSON(&profileParams)
	if err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": "Unable to bind request JSON"})
		return
	}

	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "Unable to find user config dir"})
		return
	}

	err = os.MkdirAll(filepath.Join(userConfigDir, "eloquentia"), 0755)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "Unable to create database file. Check permissions."})
		return
	}
	dbPath := filepath.Join(userConfigDir, "eloquentia/eloquentia.db")
	_, err = os.Stat(dbPath)
	if errors.Is(err, os.ErrNotExist) {
		err = os.WriteFile(dbPath, nil, 0644)
		if err != nil {
			c.AbortWithStatusJSON(500, gin.H{"error": "Unable to create database file. Check permissions."})
			return
		}
	} else if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "Unable to open database file. Check permissions."})
		return
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "Unable to open database"})
		return
	}
	defer db.Close()

	_, err = db.Exec(`create table if not exists Users (id integer primary key autoincrement,
		name varchar(255),
		native_language varchar(255),
		target_language varchar(255))`)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "Failed querying database"})
		return
	}

	_, err = db.Exec(`
		insert into users (name, native_language, target_language) values (?, ?, ?)`,
		profileParams.Name,
		profileParams.NativeLanguage,
		profileParams.TargetLanguage,
	)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "Failed querying database"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Success!"})
}

func Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
