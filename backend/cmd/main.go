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

type App struct {
	db *sql.DB
}

type ProfileParams struct {
	Name           string `json:"name" binding:"required"`
	NativeLanguage string `json:"native_language" binding:"required"`
	TargetLanguage string `json:"target_language" binding:"required"`
}

func main() {
	app, err := NewApp()
	if err != nil {
		log.Fatal(err)
	}

	defer app.db.Close()

	err = Run(NewRouter(app))
	if err != nil {
		log.Fatal(err)
	}
}

func NewRouter(app *App) *gin.Engine {
	r := gin.Default()
	r.GET("/health", Health)
	r.POST("/profile/create", app.CreateProfile)
	return r
}

func NewApp() (*App, error) {
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}

	err = os.MkdirAll(filepath.Join(userConfigDir, "eloquentia"), 0755)
	if err != nil {
		return nil, err
	}
	dbPath := filepath.Join(userConfigDir, "eloquentia/eloquentia.db")
	_, err = os.Stat(dbPath)
	if errors.Is(err, os.ErrNotExist) {
		err = os.WriteFile(dbPath, nil, 0644)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`create table if not exists Users (id integer primary key autoincrement,
		name varchar(255),
		native_language varchar(255),
		target_language varchar(255))`)
	if err != nil {
		db.Close()
		return nil, err
	}

	return &App{db: db}, nil
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

func (a *App) CreateProfile(c *gin.Context) {
	var profileParams ProfileParams
	err := c.ShouldBindJSON(&profileParams)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Missing request parameters"})
		return
	}

	profile, err := a.db.Exec(`
		insert into Users (name, native_language, target_language) values (?, ?, ?)`,
		profileParams.Name,
		profileParams.NativeLanguage,
		profileParams.TargetLanguage,
	)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed querying database"})
		return
	}

	id, err := profile.LastInsertId()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal struct to JSON"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Success!", "id": id})
}

func Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
