package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"zmb-assistant/services"
)

type chatReq struct {
	Message     string  `json:"message"`
	Temperature float32 `json:"temperature,omitempty"`
	Lang        string  `json:"lang,omitempty"`
}

type chatResp struct {
	Reply   string   `json:"reply"`
	Hints   []string `json:"hints,omitempty"`
	Actions []struct {
		Type    string      `json:"type"`
		Payload interface{} `json:"payload,omitempty"`
	} `json:"actions,omitempty"`
}

func RegisterChat(mux *http.ServeMux) {
	mux.HandleFunc("/chat", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost { http.Error(w, "use POST", http.StatusMethodNotAllowed); return }
		var req chatReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil { http.Error(w, "bad json", http.StatusBadRequest); return }
		if req.Temperature == 0 { req.Temperature = 0.2 }
		if req.Lang == "" {
			al := strings.ToLower(r.Header.Get("Accept-Language"))
			for _, cand := range []string{"kk","ru","en"} {
				if strings.HasPrefix(al, cand) { req.Lang = cand; break }
			}
			if req.Lang == "" { req.Lang = "kk" }
		}
		llm := services.NewLLM(os.Getenv("OPENAI_BASE_URL"), os.Getenv("OPENAI_API_KEY"))
		sys := services.ChatMessage{Role:"system", Content:
			"ZamanBank ассистенті. Негізгі тіл: қазақ тілі. Қысқа, жылы, нақты. " +
			"Исламдық қаржыландыру қағидаларын сақта. Спекуляция жоқ. " +
			"Қажет болса орысша/ағылшынша жауап бер."}
		user := services.ChatMessage{Role:"user", Content: req.Message}
		reply, err := llm.Chat("gpt-4o-mini", []services.ChatMessage{sys, user}, req.Temperature)
		if err != nil { http.Error(w, err.Error(), http.StatusBadGateway); return }

		resp := chatResp{Reply: reply, Hints: hintsByLang(req.Lang)}
		_ = json.NewEncoder(w).Encode(resp)
	})
}

func hintsByLang(lang string) []string {
	switch lang {
	case "ru":
		return []string{"На какой срок цель?", "Какой ежемесячный доход?", "Какие продукты вас интересуют?"}
	case "en":
		return []string{"What is your target term?", "What's your monthly income?", "Which products are you interested in?"}
	default:
		return []string{"Мақсат мерзімі қандай?", "Ай сайынғы табыс қандай?", "Қандай өнімдер қызықтырады?"}
	}
}
