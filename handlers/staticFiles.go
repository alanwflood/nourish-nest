package handlers

import (
	"crypto/md5"
	"embed"
	"encoding/hex"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
)

var embededFiles embed.FS

func calculateMD5(filename string) (string, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}

	hashInBytes := md5.Sum(content)
	hashString := hex.EncodeToString(hashInBytes[:])

	return hashString, nil
}

func getFileHashes(fsys fs.FS) {
	hashes := make(map[string]string)

	err := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			fmt.Println(err) // can't walk here,
			return nil       // but continue walking elsewhere
		}

		if !d.IsDir() {
			md5hash, err := calculateMD5(path)
			if err != nil {
				fmt.Printf("Error calculating MD5 hash for %s: %v\n", path, err)
			} else {
				hashes[path] = md5hash
			}
		}

		return nil
	})
	if err != nil {
		fmt.Printf("Error walking the path %v: %v\n", err)
		return
	}

	// Print the map
	for path, hash := range hashes {
		fmt.Printf("File: %s, MD5 hash: %s\n", path, hash)
	}
}

func getFileSystem() http.FileSystem {
	staticFilesDir := os.Getenv("STATIC_FILES_DIR")
	fsys, err := fs.Sub(embededFiles, staticFilesDir)
	if err != nil {
		panic(err)
	}

	log.Printf("Serving static files from directory: '%s'", staticFilesDir)
	return http.FS(fsys)
}

func UseFileServerHandler(e *echo.Echo) {
	assetHandler := http.FileServer(getFileSystem())
	e.GET("/public/*", echo.WrapHandler(assetHandler))
}
