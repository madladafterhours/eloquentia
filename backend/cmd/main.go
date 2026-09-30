package main

import (
	"database/sql"
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
	_ "github.com/glebarez/go-sqlite"

	models "eloquentia/backend/internal"
)

type App struct {
	db *sql.DB
}

type Profile struct {
	ID int64 `json:"id"`
	models.ProfileParams
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
	r.GET("/profile/fetch", app.GetProfile)
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
	dbPath := filepath.Join(userConfigDir, "eloquentia", "eloquentia.db")

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
	var profileParams models.ProfileParams
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
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to get inserted id"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Success!", "id": id})
}

func (a *App) GetProfile(c *gin.Context) {
	id, err := strconv.ParseInt(c.Query("id"), 10, 64)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid or missing id"})
		return
	}

	var profile Profile
	err = a.db.QueryRow(`select id, name, native_language, target_language from Users where id = ?`, id).
		Scan(&profile.ID, &profile.Name, &profile.NativeLanguage, &profile.TargetLanguage)
	if errors.Is(err, sql.ErrNoRows) {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "Profile not found"})
		return
	}
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed querying database"})
		return
	}

	c.JSON(http.StatusOK, profile)
}

func Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
