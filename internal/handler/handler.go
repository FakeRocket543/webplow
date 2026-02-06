package handler

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
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
	logger *log.Logger
}

func New(cfg *config.Config, store *auth.Store) *Handler {
	var logger *log.Logger
	if cfg.LogFile != "" {
		f, err := os.OpenFile(cfg.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err == nil {
			logger = log.New(f, "", 0)
		}
	}
	return &Handler{
		cfg:   cfg,
		store: store,
		logger: logger,
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

func (h *Handler) logRequest(user, filename string, inSize int64, status int, dur time.Duration) {
	if h.logger == nil {
		return
	}
	rec, _ := json.Marshal(map[string]interface{}{
		"time":     time.Now().UTC().Format(time.RFC3339),
		"user":     user,
		"file":     filename,
		"in_bytes": inSize,
		"status":   status,
		"ms":       dur.Milliseconds(),
	})
	h.logger.Println(string(rec))
}

func writeErr(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(errResponse{msg})
}

func (h *Handler) Convert(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	user, ok := h.store.Valid(r.Header.Get("X-API-Key"))
	if !ok {
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

	written, err := io.Copy(out, file)
	if err != nil {
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
		h.logRequest(user, header.Filename, written, 502, time.Since(start))
		writeErr(w, http.StatusBadGateway, "Backend unreachable")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		h.logRequest(user, header.Filename, written, 500, time.Since(start))
		writeErr(w, http.StatusInternalServerError, "Conversion failed")
		return
	}

	w.Header().Set("Content-Type", "image/webp")
	if cl := resp.Header.Get("Content-Length"); cl != "" {
		w.Header().Set("Content-Length", cl)
	}
	io.Copy(w, resp.Body)

	h.logRequest(user, header.Filename, written, 200, time.Since(start))
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
