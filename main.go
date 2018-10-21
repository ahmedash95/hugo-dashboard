package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

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
	router.POST("/page", savePageHandler)
	router.POST("/create/file", createFileHandler)
	router.POST("/create/dir", createDirectoryHandler)
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

func savePageHandler(c *gin.Context) {
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

	page_content, _ := c.GetPostForm("content")

	saveErr := page.UpdateContent(page_content)

	if saveErr != nil {
		c.JSON(422, gin.H{
			"msg": saveErr,
		})
		return
	}

	c.JSON(200, gin.H{
		"msg": "changes has been saved succesfully",
	})
}

func createFileHandler(c *gin.Context) {
	path, _ := c.GetPostForm("path")
	fpath := hugo.Get().ContentPath + "/" + path
	spath := strings.Split(path, "/")
	fname := spath[len(spath)-1]
	f, err := os.Create(fpath)

	if err != nil {
		c.JSON(422, gin.H{
			"msg": err,
		})
		return
	}

	defer f.Close()

	hugo.Get().AddPage("/"+path, hugo.Page{
		Title:   fname,
		Path:    fpath,
		Content: "",
	})

	fmt.Println(hugo.Get().GetPages())

	c.JSON(200, gin.H{
		"msg": fmt.Sprintf("file has been created to path %s", path),
	})
}

func createDirectoryHandler(c *gin.Context) {
	path, _ := c.GetPostForm("path")
	fpath := hugo.Get().ContentPath + "/" + path

	err := os.MkdirAll(fpath, 777)

	if err != nil {
		c.JSON(422, gin.H{
			"msg": err,
		})
		return
	}

	c.JSON(200, gin.H{
		"msg": fmt.Sprintf("directory has been created to path %s", path),
	})
}
