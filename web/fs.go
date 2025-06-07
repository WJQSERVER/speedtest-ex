package web

import (
	"embed"
	"fmt"
	"io/fs"
)

var (
	//go:embed pages/*
	assetsFS embed.FS
	pages    fs.FS
)

func init() {
	// assets 嵌入文件系统
	var err error
	pages, err = fs.Sub(assetsFS, "pages")
	if err != nil {
		fmt.Printf("Failed when processing pages: %s", err)
	}
}
