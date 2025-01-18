package main

import (
	"crypto/md5"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	CacheDir  = "/var/cache/nginx/assets"
	AssetsDir = "./assets"
)

var validFileName = regexp.MustCompile(`^[a-zA-Z0-9_-]+\.(jpg|jpeg|png|webp|gif)$`)

func validateFileName(basePath, fileName string) (string, error) {
	if !validFileName.MatchString(fileName) {
		return "", fmt.Errorf("invalid file name")
	}

	fullPath := filepath.Join(basePath, fileName)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return "", fmt.Errorf("file not found: %s", fileName)
	}

	return fullPath, nil
}

func assetsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	assetPath := strings.TrimPrefix(r.URL.Path, "/assets/")
	if assetPath == "" {
		http.Error(w, "File name not specified", http.StatusBadRequest)
		return
	}

	fullPath, err := validateFileName(AssetsDir, assetPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	http.ServeFile(w, r, fullPath)
}

func clearCacheHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	assetPath := strings.TrimPrefix(r.URL.Path, "/cache/")
	if assetPath == "" {
		http.Error(w, "File name not specified", http.StatusBadRequest)
		return
	}

	_, err := validateFileName(AssetsDir, assetPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	hash := fmt.Sprintf("%x", md5.Sum([]byte("/assets/"+assetPath)))
	cacheFile := filepath.Join(CacheDir, hash)

	if _, err := os.Stat(cacheFile); os.IsNotExist(err) {
		http.Error(w, fmt.Sprintf("Cache file not found for file: %s", assetPath), http.StatusNotFound)
		return
	}

	if err := os.Remove(cacheFile); err != nil {
		http.Error(w, fmt.Sprintf("Failed to remove cache file: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Cache file removed: %s", cacheFile)
}

func main() {
	http.HandleFunc("/assets/", assetsHandler)
	http.HandleFunc("/cache/", clearCacheHandler)

	log.Println("Go app is running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
