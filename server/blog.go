package server

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/dimfeld/httptreemux"
	"github.com/kabukky/journey/database"
	"github.com/kabukky/journey/filenames"
	"github.com/kabukky/journey/structure/methods"
	"github.com/kabukky/journey/templates"
)

func indexHandler(w http.ResponseWriter, r *http.Request, params map[string]string) {
	number := params["number"]
	if number == "" {
		// Render index template (first page)
		err := templates.ShowIndexTemplate(w, r, 1)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		return
	}
	page, err := strconv.Atoi(number)
	if err != nil || page <= 1 {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	// Render index template
	err = templates.ShowIndexTemplate(w, r, page)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	return
}

func authorHandler(w http.ResponseWriter, r *http.Request, params map[string]string) {
	slug := params["slug"]
	function := params["function"]
	number := params["number"]
	if function == "" {
		// Render author template (first page)
		err := templates.ShowAuthorTemplate(w, r, slug, 1)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		return
	} else if function == "rss" {
		// Render author rss feed
		err := templates.ShowAuthorRss(w, slug)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		return
	}
	page, err := strconv.Atoi(number)
	if err != nil || page <= 1 {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	// Render author template
	err = templates.ShowAuthorTemplate(w, r, slug, page)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	return
}

func tagHandler(w http.ResponseWriter, r *http.Request, params map[string]string) {
	slug := params["slug"]
	function := params["function"]
	number := params["number"]
	if function == "" {
		// Render tag template (first page)
		err := templates.ShowTagTemplate(w, r, slug, 1)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		return
	} else if function == "rss" {
		// Render tag rss feed
		err := templates.ShowTagRss(w, slug)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		return
	}
	page, err := strconv.Atoi(number)
	if err != nil || page <= 1 {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	// Render tag template
	err = templates.ShowTagTemplate(w, r, slug, page)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	return
}

func postHandler(w http.ResponseWriter, r *http.Request, params map[string]string) {
	slug := params["slug"]
	if slug == "" {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	} else if slug == "rss" {
		// Render index rss feed
		err := templates.ShowIndexRss(w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		return
	}

	// Render post template
	err := templates.ShowPostTemplate(w, r, slug)
	if err != nil && err.Error() == "sql: no rows in result set" {
		http.Error(w, "Post Not found.", http.StatusNotFound)
		return
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	return
}

func postEditHandler(w http.ResponseWriter, r *http.Request, params map[string]string) {
	slug := params["slug"]

	if slug == "" {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	// Redirect to edit
	post, err := database.RetrievePostBySlug(slug)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	url := fmt.Sprintf("/admin/#/edit/%d", post.Id)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func assetsHandler(w http.ResponseWriter, r *http.Request, params map[string]string) {
	// Read lock global blog
	methods.Blog.RLock()
	defer methods.Blog.RUnlock()
	http.ServeFile(w, r, filepath.Join(filenames.ThemesFilepath, methods.Blog.ActiveTheme, "assets", params["filepath"]))
	return
}

func imagesHandler(w http.ResponseWriter, r *http.Request, params map[string]string) {
	http.ServeFile(w, r, filepath.Join(filenames.ImagesFilepath, params["filepath"]))
	return
}

func publicHandler(w http.ResponseWriter, r *http.Request, params map[string]string) {
	http.ServeFile(w, r, filepath.Join(filenames.PublicFilepath, params["filepath"]))
	return
}

func faviconHandler(w http.ResponseWriter, r *http.Request, params map[string]string) {
	ex, err := os.Executable()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}
	iconPath := filepath.Join(filepath.Dir(ex), "favicon.ico")
	if _, err = os.Stat(iconPath); os.IsNotExist(err) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("404 File not found."))
		return
	}

	file, err := os.OpenFile(iconPath, os.O_RDONLY, os.ModePerm)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}
	defer file.Close()

	_, _ = io.Copy(w, file)
}

func InitializeBlog(router *httptreemux.TreeMux) {
	// For index
	router.GET("/", indexHandler)
	router.GET("/favicon.ico", faviconHandler)
	router.GET("/:slug/edit", postEditHandler)
	router.GET("/:slug/", postHandler)
	router.GET("/page/:number/", indexHandler)
	// For author
	router.GET("/author/:slug/", authorHandler)
	router.GET("/author/:slug/:function/", authorHandler)
	router.GET("/author/:slug/:function/:number/", authorHandler)
	// For tag
	router.GET("/tag/:slug/", tagHandler)
	router.GET("/tag/:slug/:function/", tagHandler)
	router.GET("/tag/:slug/:function/:number/", tagHandler)
	// For serving asset files
	router.GET("/assets/*filepath", assetsHandler)
	router.GET("/images/*filepath", imagesHandler)
	router.GET("/content/images/*filepath", imagesHandler) // This is here to keep compatibility with Ghost
	router.GET("/public/*filepath", publicHandler)
}
