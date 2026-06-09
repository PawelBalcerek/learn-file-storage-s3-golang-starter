package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"mime"
	"os"
	"path/filepath"
)

func assetFileName(mediaType string) (string, error) {
	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}

	fileName := base64.RawURLEncoding.EncodeToString(randomBytes)

	extensions, err := mime.ExtensionsByType(mediaType)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s%s", fileName, extensions[0]), nil
}

func (cfg apiConfig) ensureAssetsDir() error {
	if _, err := os.Stat(cfg.assetsRoot); os.IsNotExist(err) {
		return os.Mkdir(cfg.assetsRoot, 0o755)
	}
	return nil
}

func (cfg apiConfig) assetPath(fileName string) string {
	return filepath.Join(cfg.assetsRoot, fileName)
}

func (cfg apiConfig) assetURL(fileName string) *string {
	return ptr(fmt.Sprintf("http://localhost:%s/assets/%s", cfg.port, fileName))
}

func (cfg apiConfig) assetS3URL(fileName string) *string {
	return ptr(fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", cfg.s3Bucket, cfg.s3Region, fileName))
}

func ptr[T any](v T) *T {
	return &v
}
