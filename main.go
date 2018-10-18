package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/ahmedash95/hugo-dashboard/hugo"
	"github.com/gin-gonic/gin"
)

var port *string
var sitePath *string
var contentDir *string

var Hugo *hugo.Hugo

func main() {

	port = flag.String("port", "9999", "Run dashboard of your website on speicifc port")
	sitePath = flag.String("path", "", "hugo website directory path to serve")
	contentDir = flag.String("content-dir", "content", "hugo content directory if not content")
	flag.Parse()

	checkpath(*sitePath)

	hugo.Init(*sitePath, *contentDir)

	router := gin.New()
	router.LoadHTMLGlob("static/*.tmpl")
	router.GET("/", indexHandler)
	router.GET("/page", pageHandler)
	router.Run(fmt.Sprintf(":%s", *port))
}

func indexHandler(c *gin.Context) {
	h := hugo.Get()
	pages_list, _ := json.Marshal(h.GetPagesTree())
	c.HTML(http.StatusOK, "dashboard.tmpl", gin.H{
		"title":      h.Title,
		"theme":      h.Theme,
		"path":       h.Path,
		"pages":      h.GetPages(),
		"pages_list": string(pages_list),
	})
}

func checkpath(s string) {
	if s == "" {
		log.Fatal("path parameter is required")
	}
	if _, err := os.Stat(s); os.IsNotExist(err) {
		log.Fatal("hugo site path is not exists")
	}

	if _, err := os.Stat(fmt.Sprintf("%s/config.toml", s)); os.IsNotExist(err) {
		log.Fatal("Invalid hugo website")
	}
}

func pageHandler(c *gin.Context) {
	id, defined := c.GetQuery("p")
	if !defined {
		c.JSON(422, gin.H{
			"msg": "Parameter p should be defined and contain page path id",
		})
		return
	}
	page, err := hugo.FindPage(id)
	if err != nil {
		c.JSON(404, nil)
		return
	}
	c.JSON(200, page)
}
