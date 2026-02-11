package main

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

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

	thumbnail, h, err := r.FormFile("thumbnail")
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "formfile error", err)
		return
	}

	mt, _, err := mime.ParseMediaType(h.Header.Get("Content-Type"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "mime type parse failed", err)
		return
	}

	allowedMediaTypes := map[string]struct{}{
		"image/jpeg": {},
		"image/png":  {},
	}

	if _, ok := allowedMediaTypes[mt]; !ok {
		respondWithError(w, http.StatusBadRequest, fmt.Sprintf("%v mime type not allowed", mt), err)
		return
	}

	extension := strings.FieldsFunc(mt, func(r rune) bool {
		return r == '/'
	})

	if len(extension) != 2 {
		respondWithError(w, http.StatusBadRequest, "file extension parsing error", err)
	}

	fileExtension := extension[1]

	fmt.Println(fileExtension)

	imageBte, err := io.ReadAll(thumbnail)
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

	filename := fmt.Sprintf("%s.%s", videoID.String(), fileExtension)
	fullpath := filepath.Join(cfg.assetsRoot, filename)

	file, err := os.Create(fullpath)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error creating a file", err)
		return
	}

	_, err = io.Copy(file, strings.NewReader(string(imageBte)))
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Can not copy data to disk", err)
		return
	}

	tmbURL := fmt.Sprintf("http://localhost:%s/assets/%s.%s", cfg.port, videoID.String(), fileExtension)

	vid.ThumbnailURL = &tmbURL

	err = cfg.db.UpdateVideo(vid)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "error updating video", err)
		return
	}

	respondWithJSON(w, http.StatusOK, vid)
}
