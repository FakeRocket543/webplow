package handler

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"webplow/internal/auth"
	"webplow/internal/config"
)

type errResponse struct {
	Error string `json:"error"`
}

type Handler struct {
	cfg    *config.Config
	store  *auth.Store
	client *http.Client
}

func New(cfg *config.Config, store *auth.Store) *Handler {
	return &Handler{
		cfg:   cfg,
		store: store,
		client: &http.Client{
			Timeout: 60 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 20,
				IdleConnTimeout:     90 * time.Second,
				DisableCompression:  true,
			},
		},
	}
}

func writeErr(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(errResponse{msg})
}

func (h *Handler) Convert(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	if _, ok := h.store.Valid(r.Header.Get("X-API-Key")); !ok {
		writeErr(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, h.cfg.MaxFileSize)
	if err := r.ParseMultipartForm(h.cfg.MaxFileSize); err != nil {
		writeErr(w, http.StatusBadRequest, "Failed to parse form")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeErr(w, http.StatusBadRequest, "No file uploaded")
		return
	}
	defer file.Close()

	filename := fmt.Sprintf("img_%d_%s", time.Now().UnixNano(), filepath.Base(header.Filename))
	tempFile := filepath.Join(h.cfg.TempDir, filename)

	out, err := os.Create(tempFile)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "Failed to save file")
		return
	}
	defer os.Remove(tempFile)

	if _, err := io.Copy(out, file); err != nil {
		out.Close()
		writeErr(w, http.StatusInternalServerError, "Failed to write file")
		return
	}
	out.Close()

	localURL := "local:///" + filename
	encoded := base64.RawURLEncoding.EncodeToString([]byte(localURL))
	imgproxyPath := fmt.Sprintf("/insecure/q:85/rs:fit:0:0/f:webp/%s", encoded)

	resp, err := h.client.Get(h.cfg.ImgproxyURL + imgproxyPath)
	if err != nil {
		writeErr(w, http.StatusBadGateway, "Backend unreachable")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		writeErr(w, http.StatusInternalServerError, "Conversion failed")
		return
	}

	w.Header().Set("Content-Type", "image/webp")
	if cl := resp.Header.Get("Content-Length"); cl != "" {
		w.Header().Set("Content-Length", cl)
	}
	io.Copy(w, resp.Body)
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	resp, err := h.client.Get(h.cfg.ImgproxyURL + "/health")
	if err != nil || resp.StatusCode != 200 {
		writeErr(w, http.StatusServiceUnavailable, "Backend unhealthy")
		return
	}
	resp.Body.Close()
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"ok"}`))
}
