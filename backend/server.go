package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"unicode/utf8"

	"zmb-assistant/handlers"
	"zmb-assistant/services"

	"golang.org/x/text/encoding/charmap"
)

var sessions = NewSessionStore()

func setJSON(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
}

func writeErr(w http.ResponseWriter, status int, msg string) {
	setJSON(w)
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{"error": msg, "status": status})
}

func startServer() {
	mux := http.NewServeMux()

	// --- health ---
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	// --- small LLM ping ---
	mux.HandleFunc("/llm/ping", func(w http.ResponseWriter, r *http.Request) {
		setJSON(w)
		llm := services.NewLLM(os.Getenv("OPENAI_BASE_URL"), os.Getenv("OPENAI_API_KEY"))
		reply, err := llm.Chat("gpt-4o-mini",
			[]services.ChatMessage{{Role: "user", Content: "Скажи: pong"}}, 0)
		if err != nil {
			writeErr(w, http.StatusBadGateway, err.Error())
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"reply": reply})
	})

	// --- sessions ---
	mux.HandleFunc("/session/new", func(w http.ResponseWriter, r *http.Request) {
		setJSON(w)
		id := sessions.EnsureSessionID(w, r)
		sessions.Reset(id)
		_ = json.NewEncoder(w).Encode(map[string]any{"session_id": id, "status": "new"})
	})

	mux.HandleFunc("/session/clear", func(w http.ResponseWriter, r *http.Request) {
		setJSON(w)
		id := sessions.EnsureSessionID(w, r)
		sessions.Reset(id)
		_ = json.NewEncoder(w).Encode(map[string]any{"session_id": id, "status": "cleared"})
	})

	mux.HandleFunc("/session/history", func(w http.ResponseWriter, r *http.Request) {
		setJSON(w)
		id := sessions.EnsureSessionID(w, r)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"session_id": id,
			"history":    sessions.Get(id),
		})
	})

	// --- chat ---
	mux.HandleFunc("/llm/chat", func(w http.ResponseWriter, r *http.Request) {
		setJSON(w)
		if r.Method != http.MethodPost {
			writeErr(w, http.StatusMethodNotAllowed, "use POST")
			return
		}
		id := sessions.EnsureSessionID(w, r)
		var body struct {
			Message string  `json:"message"`
			Temp    float32 `json:"temperature,omitempty"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeErr(w, http.StatusBadRequest, "invalid JSON")
			return
		}
		// try CP1251 → UTF-8 if needed (Windows/Git Bash)
		if !utf8.ValidString(body.Message) {
			if dec, err := charmap.Windows1251.NewDecoder().String(body.Message); err == nil {
				body.Message = dec
			}
		}
		if body.Temp == 0 {
			body.Temp = 0.2
		}

		history := sessions.Get(id)
		if len(history) == 0 {
			history = append(history, services.ChatMessage{
				Role: "system",
				Content: "Сен — ZamanBank ассистентісің. Негізгі тіл: қазақ тілі. " +
					"Жауаптарың қысқа, жылы әрі нақты болсын. " +
					"Исламдық қаржыландыру қағидаларын сақта: риба/өсім жоқ, " +
					"спекуляциядан аулақ, активке негізделген ұсыныстар. " +
					"Қажет болса орысша түсіндір, бірақ әуелі қазақша жауап бер.",
			})
		}

		userMsg := services.ChatMessage{Role: "user", Content: body.Message}
		history = append(history, userMsg)
		sessions.Append(id, userMsg)

		llm := services.NewLLM(os.Getenv("OPENAI_BASE_URL"), os.Getenv("OPENAI_API_KEY"))
		reply, err := llm.Chat("gpt-4o-mini", history, body.Temp)
		if err != nil {
			writeErr(w, http.StatusBadGateway, err.Error())
			return
		}
		sessions.Append(id, services.ChatMessage{Role: "assistant", Content: reply})

		_ = json.NewEncoder(w).Encode(map[string]any{
			"session_id":  id,
			"reply":       reply,
			"history_len": len(sessions.Get(id)),
		})
	})

	// --- your domain handlers ---
	handlers.RegisterChat(mux)
	handlers.RegisterAI(mux, os.Getenv("OPENAI_BASE_URL"), os.Getenv("OPENAI_API_KEY"))
	handlers.RegisterProducts(mux, os.Getenv("DATA_DIR"))
	handlers.RegisterGoals(mux)
	handlers.RegisterSpending(mux, os.Getenv("DATA_DIR"))

	// --- serve OpenAPI file for Swagger UI ---
	mux.HandleFunc("/openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "openapi.yaml") // путь от каталога backend
	})

	// --- static files (optional) ---
	mux.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir("../"))))

	// --- wrappers: logging + CORS ---
	handler := withCORS(withLogging(mux))

	// --- start ---
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Listening on :%s", port)
	serveSwagger(mux) // /swagger UI
	log.Fatal(http.ListenAndServe(":"+port, handler))
}
