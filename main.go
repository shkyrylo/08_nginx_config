package main

import (
	"crypto/md5"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const CacheDir = "/var/cache/nginx/assets"

func assetsHandler(w http.ResponseWriter, r *http.Request) {
	assetPath := strings.TrimPrefix(r.URL.Path, "/assets/")
	if assetPath == "" {
		http.Error(w, "Asset not specified", http.StatusBadRequest)
		return
	}

	fullPath := filepath.Join("./assets", assetPath)

	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		http.Error(w, "Asset not found", http.StatusNotFound)
		return
	}

	http.ServeFile(w, r, fullPath)
}

func clearCacheHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	keys, ok := r.URL.Query()["key"]
	if !ok || len(keys[0]) < 1 {
		http.Error(w, "Missing 'key' parameter", http.StatusBadRequest)
		return
	}
	cacheKey := keys[0]

	hash := fmt.Sprintf("%x", md5.Sum([]byte(cacheKey)))
	cacheFile := filepath.Join(CacheDir, hash)

	fmt.Println("Removing cache file:", cacheFile)

	if _, err := os.Stat(cacheFile); os.IsNotExist(err) {
		http.Error(w, fmt.Sprintf("Cache file not found for key: %s", cacheKey), http.StatusNotFound)
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
	http.HandleFunc("/cache", clearCacheHandler)

	log.Println("Go app is running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
