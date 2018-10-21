package hugo

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/cheggaaa/pb.v1"

	"github.com/spf13/viper"
)

type Hugo struct {
	Title       string `json:"title"`
	Path        string `json:"path"`
	Theme       string `json:"theme"`
	BaseURI     string `json:"base_uri"`
	ContentPath string `json:"content_path"`
	pages       map[string]Page
}

type Page struct {
	Title   string `json:"title"`
	Date    string `json:"date"`
	Content string `json:"content"`
	Path    string `json:"path"`
}

var hugoSite *Hugo

/**
	Init is responsible to load all date we need for sepcific hugo site
**/
func Init(path string, contentDir string) {

	viper.AddConfigPath(path) // optionally look for config in the working directory

	err := viper.ReadInConfig() // Find and read the config file

	if err != nil { // Handle errors reading the config file
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}

	hugoSite = &Hugo{
		Title:       viper.GetString("title"),
		Path:        path,
		Theme:       viper.GetString("theme"),
		BaseURI:     viper.GetString("baseurl"),
		ContentPath: path + "/" + contentDir,
	}

	loadPages(hugoSite)
}

func loadPages(h *Hugo) {
	contentPath := h.ContentPath
	var files []string
	var total int
	err := filepath.Walk(contentPath,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			f, err := os.Stat(path)
			if err != nil {
				fmt.Println(err.Error())
			}
			// the path point to a file not directory
			// and we should consider it as a file
			if !f.IsDir() && strings.Contains(path, ".md") {
				files = append(files, path)
				total++
			}
			return nil
		})
	if err != nil {
		log.Println(err)
	}

	bar := pb.StartNew(total)

	for _, path := range files {
		bar.Increment()
		splitedPath := strings.Split(path, "/")
		fileName := splitedPath[len(splitedPath)-1]
		content, err := ioutil.ReadFile(path)
		if err != nil {
			fmt.Printf("Error reading file %s\n", path)
			fmt.Println(err.Error())
		}
		id := strings.Replace(path, contentPath, "", 1)
		h.AddPage(id, Page{
			Title:   fileName,
			Path:    path,
			Content: string(content),
		})
	}
	bar.Finish()
}

func (h *Hugo) AddPage(id string, p Page) {
	if h.pages == nil {
		h.pages = make(map[string]Page)
	}
	h.pages[id] = p
}

func Get() *Hugo {
	return hugoSite
}

func (h *Hugo) GetPages() map[string]Page {
	return h.pages
}

func (h *Hugo) GetPagesTree() []string {
	basePath := h.ContentPath
	var tree []string
	for _, p := range h.GetPages() {
		tree = append(tree, strings.Replace(p.Path, basePath, "", 1))
	}
	return tree
}

func FindPage(s string) (Page, error) {
	c, Ok := hugoSite.GetPages()[s]
	if !Ok {
		return Page{}, errors.New("Page Not Found")
	}
	return c, nil
}

func (p *Page) UpdateContent(s string) error {
	p.Content = s
	content := []byte(s)
	err := ioutil.WriteFile(p.Path, content, 777)
	return err
}
