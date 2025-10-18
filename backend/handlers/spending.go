package handlers

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"io"
	"math"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
)

type Txn struct {
	Amount float64 `json:"amount"`
	Desc   string  `json:"desc"`
}

var (
	userTxns = map[string][]Txn{} // key: session or dummy "default"
	peers    []map[string]float64 // simple rows with "income","food","transport",...
)

func RegisterSpending(mux *http.ServeMux, dataDir string) {
	if dataDir == "" {
		dataDir = "./data"
	}
	_ = loadJSON(filepath.Join(dataDir, "peers.json"), &peers)

	mux.HandleFunc("/statements/upload", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "use POST", http.StatusMethodNotAllowed)
			return
		}
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			http.Error(w, "multipart required", http.StatusBadRequest)
			return
		}
		file, _, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "file required", http.StatusBadRequest)
			return
		}
		defer file.Close()

		count, err := readCSVInto("default", file, ',')
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"status": "ok", "count": count})
	})

	mux.HandleFunc("/spending", func(w http.ResponseWriter, r *http.Request) {
		tx := userTxns["default"]

		cats := map[string]float64{}
		for _, t := range tx {
			cat := catByDesc(t.Desc)
			cats[cat] += math.Abs(t.Amount)
		}

		var total float64
		for _, v := range cats {
			total += v
		}

		type Row struct {
			Name   string  `json:"name"`
			Amount float64 `json:"amount"`
			Share  float64 `json:"share"`
		}

		var out []Row
		for k, v := range cats {
			share := 0.0
			if total > 0 {
				share = v / total
			}
			out = append(out, Row{Name: k, Amount: v, Share: share})
		}

		tips := []string{
			"Оптимизируйте «Еда вне дома» до 10–12% от дохода.",
			"Планируйте транспорт заранее — проездные дешевле.",
		}

		_ = json.NewEncoder(w).Encode(map[string]any{"by_category": out, "tips": tips})
	})

	mux.HandleFunc("/compare/peers", func(w http.ResponseWriter, r *http.Request) {
		if len(peers) == 0 {
			_ = json.NewEncoder(w).Encode(map[string]any{"insights": []string{}})
			return
		}
		median := map[string]float64{}
		for k, v := range peers[0] {
			median[k] = v
		}
		ins := []string{
			"У похожих пользователей доля сбережений 12–18%.",
			"Чаще всего они достигают цели «резерв 1 млн» за 8–12 месяцев.",
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"percentiles": median, "insights": ins})
	})
}

func readCSVInto(key string, r io.Reader, sep rune) (int, error) {
	rd := csv.NewReader(r)
	rd.Comma = sep

	var list []Txn
	for {
		rec, err := rd.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return 0, errors.New("bad csv")
		}
		if len(rec) < 2 {
			continue
		}
		amt, err := strconv.ParseFloat(strings.TrimSpace(rec[0]), 64)
		if err != nil {
			return 0, errors.New("amount must be number in first column")
		}
		desc := strings.TrimSpace(rec[1])
		if desc == "" {
			return 0, errors.New("desc (second column) is required")
		}
		list = append(list, Txn{Amount: amt, Desc: desc})
	}
	userTxns[key] = append(userTxns[key], list...)
	return len(list), nil
}

func catByDesc(d string) string {
	dd := strings.ToLower(d)
	switch {
	case strings.Contains(dd, "market"),
		strings.Contains(dd, "food"),
		strings.Contains(dd, "cafe"):
		return "Еда"
	case strings.Contains(dd, "pharm"),
		strings.Contains(dd, "clinic"):
		return "Медицина"
	case strings.Contains(dd, "taxi"),
		strings.Contains(dd, "bus"),
		strings.Contains(dd, "fuel"),
		strings.Contains(dd, "petrol"):
		return "Транспорт"
	case strings.Contains(dd, "rent"),
		strings.Contains(dd, "kommun"),
		strings.Contains(dd, "utility"):
		return "Жильё/Коммуналка"
	default:
		return "Прочее"
	}
}
