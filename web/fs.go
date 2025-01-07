package web

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/gin-gonic/gin"
)

var pages fs.FS

var (
	//go:embed pages/*
	assetsFS embed.FS
)

func init() {
	// assets 嵌入文件系统
	var err error
	pages, err = fs.Sub(assetsFS, "pages")
	if err != nil {
		logError("Failed when processing pages: %s", err)
	}
}

func PagesEmbedFS(c *gin.Context) {
	path := c.Request.URL.Path

	if path == "" || path == "/" {
		http.FileServer(http.FS(pages)).ServeHTTP(c.Writer, c.Request)
		return
	}

	if pagesPathRegex.MatchString(path) {

		// 去除path开头的/
		if len(path) > 1 && path[0] == '/' {
			if path[1:] != "" {
				path = path[1:]
			}
		}

		// 预检测fs内是否存在path.html
		if _, err := fs.Stat(pages, path+".html"); err == nil {
			path = path + ".html"
		}

		data, err := fs.ReadFile(pages, path)
		if err != nil {
			logError("Failed when processing %s: %s", path, err)
			c.String(http.StatusNotFound, "Not Found")
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", data)
		return

	}

	/*
		// 若路径等于 /login,/chart,/single 则伪静态到对应的.html 文件
		if path == "/login" || path == "/login.html" || path == "/chart" || path == "/single" {
			path = path + ".html"
			// 去除path开头的/
			if len(path) > 1 && path[0] == '/' {
				path = path[1:]
			}
			data, err := fs.ReadFile(pages, path)
			if err != nil {
				logError("Failed when processing %s: %s", path, err)
				c.String(http.StatusNotFound, "Not Found")
				return
			}
			c.Data(http.StatusOK, "text/html; charset=utf-8", data)
			return
		}*/

	http.FileServer(http.FS(pages)).ServeHTTP(c.Writer, c.Request)
	return
}

// for debug only
/*
func listFiles(fsys fs.FS) {
	entries, err := fs.ReadDir(fsys, ".")
	if err != nil {
		logError("Failed to read directory: %s", err)
		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			fmt.Printf("Directory: %s", entry.Name())
		} else {
			fmt.Printf("File: %s", entry.Name())
		}
	}
}
*/
