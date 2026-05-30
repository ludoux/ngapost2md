package server

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"path/filepath"
	"strings"
)

//go:embed frontend/*
var frontendFS embed.FS

func frontendHandler() http.Handler {
	subFS, err := fs.Sub(frontendFS, "frontend")
	if err != nil {
		log.Fatalln("无法创建前端子文件系统:", err)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if path == "/" {
			path = "/index.html"
		}

		// 清理路径，防止目录遍历
		cleanPath := filepath.Clean(path)
		// 为windows修复
		cleanPath = strings.ReplaceAll(cleanPath, "\\", "/")
		if strings.Contains(cleanPath, "..") {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		// fs.ReadFile 要求路径不能以 / 开头（fs.ValidPath 规则）
		// 去掉前导 /，使其成为相对于 subFS 根目录的相对路径
		fsPath := strings.TrimPrefix(cleanPath, "/")

		// 直接从嵌入的文件系统读取文件
		data, err := fs.ReadFile(subFS, fsPath)
		if err != nil {
			// 文件不存在，返回 index.html（SPA-like fallback）
			data, err = fs.ReadFile(subFS, "index.html")
			if err != nil {
				http.Error(w, "Not Found", http.StatusNotFound)
				return
			}
			fsPath = "index.html"
		}

		// 设置 Content-Type
		ext := strings.ToLower(filepath.Ext(fsPath))
		switch ext {
		case ".html":
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
		case ".css":
			w.Header().Set("Content-Type", "text/css; charset=utf-8")
		case ".js":
			w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
		case ".json":
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
		case ".png":
			w.Header().Set("Content-Type", "image/png")
		case ".jpg", ".jpeg":
			w.Header().Set("Content-Type", "image/jpeg")
		case ".gif":
			w.Header().Set("Content-Type", "image/gif")
		case ".svg":
			w.Header().Set("Content-Type", "image/svg+xml")
		case ".ico":
			w.Header().Set("Content-Type", "image/x-icon")
		case ".woff":
			w.Header().Set("Content-Type", "font/woff")
		case ".woff2":
			w.Header().Set("Content-Type", "font/woff2")
		case ".webp":
			w.Header().Set("Content-Type", "image/webp")
		}

		w.Write(data)
	})
}
