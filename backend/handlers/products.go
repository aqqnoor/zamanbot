package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"golang.org/x/text/encoding/charmap"
)

type Product struct {
	ID        string  `json:"id"`
	Type      string  `json:"type"`
	NameKK    string  `json:"name_kk"`
	NameRU    string  `json:"name_ru"`
	NameEN    string  `json:"name_en"`
	DescKK    string  `json:"desc_kk"`
	DescRU    string  `json:"desc_ru"`
	DescEN    string  `json:"desc_en"`
	Rate      float64 `json:"rate,omitempty"`
	Term      int     `json:"term,omitempty"`
	MinAmount float64 `json:"min_amount,omitempty"`
	Shariah   bool    `json:"shariah"`
	Link      string  `json:"link,omitempty"`
}

type GlossaryItem struct {
	Term  string `json:"term"`
	DefKK string `json:"definition_kk"`
	DefRU string `json:"definition_ru"`
	DefEN string `json:"definition_en"`
	Link  string `json:"faq_link,omitempty"`
}

type ProductView struct {
	ID        string  `json:"id"`
	Type      string  `json:"type"`
	Name      string  `json:"name"`
	Desc      string  `json:"desc"`
	Rate      float64 `json:"rate,omitempty"`
	Term      int     `json:"term,omitempty"`
	MinAmount float64 `json:"min_amount,omitempty"`
	Shariah   bool    `json:"shariah"`
	Link      string  `json:"link,omitempty"`
}

var (
	products []Product
	glossary []GlossaryItem
)

func RegisterProducts(mux *http.ServeMux, dataDir string) {
	if dataDir == "" { dataDir = "./data" }
	if err := loadJSON(filepath.Join(dataDir, "products.json"), &products); err != nil {
		log.Printf("[products] load error: %v", err)
	}
	if err := loadJSON(filepath.Join(dataDir, "glossary.json"), &glossary); err != nil {
		log.Printf("[glossary] load error: %v", err)
	}
	log.Printf("[glossary] loaded: %d", len(glossary))

	mux.HandleFunc("/products", func(w http.ResponseWriter, r *http.Request) {
		typ := r.URL.Query().Get("type")
		lang := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("lang")))
		if lang == "" { lang = "kk" }
		var out []ProductView
		for _, p := range products {
			if typ != "" && p.Type != typ { continue }
			var name, desc string
			switch lang {
			case "kk": name, desc = p.NameKK, p.DescKK
			case "ru": name, desc = p.NameRU, p.DescRU
			default:   name, desc = p.NameEN, p.DescEN
			}
			out = append(out, ProductView{
				ID: p.ID, Type: p.Type, Name: name, Desc: desc,
				Rate: p.Rate, Term: p.Term, MinAmount: p.MinAmount,
				Shariah: p.Shariah, Link: p.Link,
			})
		}
		_ = json.NewEncoder(w).Encode(out)
	})

	mux.HandleFunc("/glossary", func(w http.ResponseWriter, r *http.Request) {
		q := strings.TrimSpace(r.URL.Query().Get("term"))
		if q == "" { _ = json.NewEncoder(w).Encode(glossary); return }
		if !utf8.ValidString(q) { if dec, err := charmap.Windows1251.NewDecoder().String(q); err == nil { q = dec } }
		for _, g := range glossary {
			if strings.EqualFold(strings.TrimSpace(g.Term), q) { _ = json.NewEncoder(w).Encode(g); return }
		}
		ql := strings.ToLower(q)
		for _, g := range glossary {
			if strings.Contains(strings.ToLower(strings.TrimSpace(g.Term)), ql) {
				_ = json.NewEncoder(w).Encode(g); return
			}
		}
		http.Error(w, "not found", http.StatusNotFound)
	})
}

func loadJSON(path string, v any) error {
	f, err := os.Open(path); if err != nil { return err }
	defer f.Close()
	return json.NewDecoder(f).Decode(v)
}
