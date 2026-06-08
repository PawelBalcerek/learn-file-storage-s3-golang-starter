package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"io"
	"log"
	"mime"
	"net/http"
	"os"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadThumbnail(w http.ResponseWriter, r *http.Request) {
	videoID, err := uuid.Parse(r.PathValue("videoID"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}

	log.Printf("uploading thumbnail for video %s by user %s", videoID, userID)

	const maxMemory = 10 << 20
	r.ParseMultipartForm(maxMemory)
	formFile, fileHeader, err := r.FormFile("thumbnail")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to parse form file", err)
		return
	}
	defer formFile.Close()

	mediaType, _, err := mime.ParseMediaType(fileHeader.Header.Get("Content-Type"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Failed to parse file media type", err)
		return
	}

	if mediaType != "image/jpeg" && mediaType != "image/png" {
		respondWithError(w, http.StatusBadRequest, "Invalid media type", nil)
		return
	}

	video, err := cfg.db.GetVideo(videoID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithJSON(w, http.StatusNotFound, "Video could not be found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to get video", err)
		return
	}

	if video.UserID != userID {
		respondWithJSON(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to generate random bytes", err)
		return
	}

	fileName := base64.RawURLEncoding.EncodeToString(randomBytes)
	assetPath, err := assetPath(fileName, mediaType)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve asset path", err)
		return
	}

	file, err := os.Create(cfg.assetsPath(assetPath))
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create file", err)
		return
	}
	defer file.Close()

	if _, err := io.Copy(file, formFile); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to copy file from form file", err)
		return
	}

	video.ThumbnailURL = cfg.assetURL(assetPath)
	if err := cfg.db.UpdateVideo(video); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to update video", err)
		return
	}

	respondWithJSON(w, http.StatusOK, video)
}
