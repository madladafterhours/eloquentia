package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	err := run(newRouter())
	if err != nil {
		log.Fatal(err)
	}
}

func newRouter() *gin.Engine {
	r := gin.Default()
	r.GET("/health", health)
	r.GET("/test", testEndpoint)
	return r
}

func run(r *gin.Engine) error {
	portEnv := os.Getenv("ELOQUENTIA_PORT")

	if portEnv == "" {
		portEnv = "8090"
	}

	port := flag.String("port", portEnv, "Port for the API")

	flag.Parse()
	return r.Run("127.0.0.1:" + *port)
}

func testEndpoint(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"response": "Hello World!"})
}

func health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
