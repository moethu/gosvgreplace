package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

var responseContentType = "text/html; charset=utf-8"
var acceptedContentType = "image/svg+xml"
var validationRegex = `^[A-Za-z0-9°%\.,\-_\#]*$`
var validateInput *regexp.Regexp
var invalidBody = "Invalid request body"
var errorRetrievingSource = "Error retrieving source"

type Payload struct {
	Source        string            `json:"source"`
	RemoveHyphens bool              `json:"removehyphens"`
	Replace       map[string]string `json:"replace"`
}

func main() {
	flag.Parse()
	log.SetFlags(0)

	validateInput = regexp.MustCompile(validationRegex)

	router := gin.Default()
	port := ":4211"
	srv := &http.Server{
		Addr:         port,
		Handler:      router,
		ReadTimeout:  600 * time.Second,
		WriteTimeout: 600 * time.Second,
	}

	router.POST("/render", renderSvg)
	log.Println("Starting HTTP Server on Port", port)

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutdown Server")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown: ", err)
	}
	log.Println("Server exiting")
}

func renderSvg(c *gin.Context) {
	jsonData, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		c.Data(http.StatusNotFound, responseContentType, []byte(invalidBody))
		return
	}

	var pload Payload
	err = json.Unmarshal(jsonData, &pload)
	if err != nil {
		c.Data(http.StatusNotFound, responseContentType, []byte(invalidBody))
		return
	}

	svgstring, err := getSVG(pload.Source, pload.RemoveHyphens)
	if err != nil {
		c.Data(http.StatusNotFound, responseContentType, []byte(errorRetrievingSource))
		return
	}

	for k, v := range pload.Replace {
		if validateInput.MatchString(v) {
			svgstring = strings.ReplaceAll(svgstring, k, v)
		}
	}

	c.Data(http.StatusOK, responseContentType, []byte(svgstring))
}

func getSVG(url string, removeHyphens bool) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}

	ctype := resp.Header.Get("Content-type")
	if ctype != acceptedContentType {
		return "", errors.New("Invalid Content-type")
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	svgstring := string(body)
	if removeHyphens {
		svgstring = strings.ReplaceAll(svgstring, "'", "")
	}
	return svgstring, nil
}
