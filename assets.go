package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"mime"
	"os"
	"os/exec"
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

func videoAssetPrefix(filePath string) (string, error) {
	cmd := exec.Command("ffprobe", "-v", "error", "-print_format", "json", "-show_streams", filePath)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("failed to attach stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("failed to start ffprobe command: %w", err)
	}

	var result struct {
		Streams []struct {
			DisplayAspectRatio string `json:"display_aspect_ratio"`
		} `json:"streams"`
	}
	if err := json.NewDecoder(stdout).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode ffprobe result: %w", err)
	}

	if err := cmd.Wait(); err != nil {
		return "", fmt.Errorf("failed to wait for ffprobe command exit or copying result: %w", err)
	}

	if len(result.Streams) == 0 {
		return "", errors.New("empty video streams")
	}

	switch result.Streams[0].DisplayAspectRatio {
	default:
		return "other/", nil
	case "16:9":
		return "landscape/", nil
	case "9:16":
		return "portrait/", nil
	}
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

func (cfg apiConfig) localAssetURL(fileName string) *string {
	return ptr(fmt.Sprintf("http://localhost:%s/assets/%s", cfg.port, fileName))
}

func (cfg apiConfig) s3AssetURL(key string) *string {
	return ptr(fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", cfg.s3Bucket, cfg.s3Region, key))
}

func ptr[T any](v T) *T {
	return &v
}
