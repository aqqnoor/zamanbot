package handlers

import (
	"encoding/json"
	"math"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

type Goal struct {
	ID             int     `json:"id"`
	Title          string  `json:"title"`
	TargetAmount   float64 `json:"target_amount"`
	DeadlineMonths int     `json:"deadline_months"`
	Income         float64 `json:"income,omitempty"`
	SavingsRate    float64 `json:"savings_rate,omitempty"`
	Lang           string  `json:"lang,omitempty"`
}

type GoalPlan struct {
	MonthlyAmount float64   `json:"monthly_amount"`
	Months        int       `json:"months"`
	Steps         []float64 `json:"steps"`
	Tips          []string  `json:"tips"`
}

var (
	goalsMu sync.Mutex
	goalsDB = map[int]Goal{}
	nextID  = 1
)

func RegisterGoals(mux *http.ServeMux) {
	mux.HandleFunc("/goals", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			var g Goal
			if err := json.NewDecoder(r.Body).Decode(&g); err != nil { http.Error(w, "bad json", http.StatusBadRequest); return }
			if strings.TrimSpace(g.Title) == "" { http.Error(w, "title is required", http.StatusBadRequest); return }
			if g.TargetAmount <= 0 { http.Error(w, "target_amount must be > 0", http.StatusBadRequest); return }
			if g.DeadlineMonths <= 0 { g.DeadlineMonths = 12 }
			goalsMu.Lock(); g.ID = nextID; nextID++; goalsDB[g.ID] = g; goalsMu.Unlock()
			_ = json.NewEncoder(w).Encode(map[string]any{"id": g.ID})
		case http.MethodGet:
			goalsMu.Lock()
			out := make([]Goal, 0, len(goalsDB))
			for _, g := range goalsDB { out = append(out, g) }
			goalsMu.Unlock()
			_ = json.NewEncoder(w).Encode(out)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/goals/", func(w http.ResponseWriter, r *http.Request) {
		parts := splitPath(r.URL.Path)
		if len(parts) == 3 && parts[2] == "plan" && r.Method == http.MethodPost {
			id, _ := strconv.Atoi(parts[1])
			goalsMu.Lock(); g, ok := goalsDB[id]; goalsMu.Unlock()
			if !ok { http.Error(w, "goal not found", http.StatusNotFound); return }
			lang := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("lang")))
			if lang == "" { lang = g.Lang }
			if lang == "" { lang = "kk" }
			plan := buildPlan(g, lang)
			_ = json.NewEncoder(w).Encode(plan); return
		}
		http.NotFound(w, r)
	})
}

func buildPlan(g Goal, lang string) GoalPlan {
	months := g.DeadlineMonths; if months <= 0 { months = 12 }
	monthly := math.Ceil(g.TargetAmount / float64(months))
	steps := make([]float64, months)
	for i := 0; i < months; i++ { steps[i] = monthly }
	tips := tipsByLang(lang, g.SavingsRate)
	return GoalPlan{MonthlyAmount: monthly, Months: months, Steps: steps, Tips: tips}
}

func tipsByLang(lang string, rate float64) []string {
	switch lang {
	case "ru":
		t := []string{
			"Сократите лишние расходы (транспорт, фастфуд, импульсные покупки).",
			"Откройте депозит и настройте автосписание раз в месяц.",
		}
		if rate < 0.1 { t = append(t, "Поднимите норму сбережений минимум до 10%.") }
		return t
	case "en":
		t := []string{
			"Cut non-essential spending (transport, fast food, impulse buys).",
			"Open a deposit and set monthly auto-transfers.",
		}
		if rate < 0.1 { t = append(t, "Raise your savings rate to at least 10%.") }
		return t
	default:
		t := []string{
			"Артық шығындарды қысқартыңыз (жол, фастфуд, импульсивті сатып алу).",
			"Депозит ашыңыз: ай сайын автоматты аударым жасаңыз.",
		}
		if rate < 0.1 { t = append(t, "Жинақ нормасын кемінде 10% деңгейіне көтеріңіз.") }
		return t
	}
}

func splitPath(p string) []string {
	out := []string{}; cur := ""
	for i := 0; i < len(p); i++ {
		if p[i] == '/' { if cur != "" { out = append(out, cur); cur = "" }; continue }
		cur += string(p[i])
	}
	if cur != "" { out = append(out, cur) }
	return out
}
