package main

import (
	"fmt"
	"mime"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

func assetPath(videoId uuid.UUID, mediaType string) (string, error) {
	extensions, err := mime.ExtensionsByType(mediaType)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s%s", videoId, extensions[0]), nil
}

func (cfg apiConfig) ensureAssetsDir() error {
	if _, err := os.Stat(cfg.assetsRoot); os.IsNotExist(err) {
		return os.Mkdir(cfg.assetsRoot, 0o755)
	}
	return nil
}

func (cfg apiConfig) assetsPath(assetPath string) string {
	return filepath.Join(cfg.assetsRoot, assetPath)
}

func (cfg apiConfig) assetURL(assetPath string) *string {
	return ptr(fmt.Sprintf("http://localhost:%s/assets/%s", cfg.port, assetPath))
}

func ptr[T any](v T) *T {
	return &v
}
