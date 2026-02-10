package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadThumbnail(w http.ResponseWriter, r *http.Request) {
	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
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

	fmt.Println("uploading thumbnail for video", videoID, "by user", userID)

	// TODO: implement the upload here
	const maxMemory = 10 << 20

	if err := r.ParseMultipartForm(maxMemory); err != nil {
		respondWithError(w, http.StatusInternalServerError, "max bytes error", err)
		return
	}

	file, h, err := r.FormFile("thumbnail")
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "formfile error", err)
		return
	}

	mt := h.Header.Get("Content-Type")

	imageBte, err := io.ReadAll(file)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error reading image date", err)
		return
	}

	vid, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error getting video", err)
		return
	}

	if vid.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "Authrorization error", err)
		return
	}

	tmb := thumbnail{
		data:      imageBte,
		mediaType: mt,
	}

	videoThumbnails[videoID] = tmb

	tmburl := fmt.Sprintf("http://localhost:%s/api/thumbnails/%s", cfg.port, videoID.String())

	vid.ThumbnailURL = &tmburl

	err = cfg.db.UpdateVideo(vid)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "error updating video", err)
		return
	}

	respondWithJSON(w, http.StatusOK, vid)
}
