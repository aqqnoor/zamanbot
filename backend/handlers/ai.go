package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"os"

	"zmb-assistant/services"
)

func RegisterAI(mux *http.ServeMux, baseURL, apiKey string) {
	embedClient := services.NewEmbedding(baseURL, apiKey)
	whisper := services.NewWhisper(baseURL, apiKey)

	mux.HandleFunc("/embed", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "use POST", http.StatusMethodNotAllowed)
			return
		}
		var body struct {
			Text string `json:"text"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "bad json", http.StatusBadRequest)
			return
		}
		vec, err := embedClient.Embed(body.Text)

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"vector": vec})
	})

	mux.HandleFunc("/transcribe", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "use POST", http.StatusMethodNotAllowed)
			return
		}
		if err := r.ParseMultipartForm(25 << 20); err != nil {
			http.Error(w, "multipart required", http.StatusBadRequest)
			return
		}
		file, hdr, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "file required", http.StatusBadRequest)
			return
		}
		defer file.Close()

		// сохранить во временный файл для клиента whisper
		tmpPath := os.TempDir() + "/" + hdr.Filename
		dst, _ := os.Create(tmpPath)
		defer dst.Close()
		_, _ = io.Copy(dst, file)

		text, err := whisper.Transcribe(tmpPath)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"text": text})
	})
}
