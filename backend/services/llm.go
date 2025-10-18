package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type LLM struct {
	baseURL string
	apiKey  string
	client  *http.Client
}

func NewLLM(baseURL, apiKey string) *LLM {
	return &LLM{
		baseURL: baseURL,
		apiKey:  apiKey,
		client:  &http.Client{Timeout: 60 * time.Second},
	}
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	Temperature float32       `json:"temperature,omitempty"`
}

type chatResponse struct {
	Choices []struct {
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func (l *LLM) Chat(model string, messages []ChatMessage, temperature float32) (string, error) {
	url := l.baseURL + "/v1/chat/completions"
	body, _ := json.Marshal(chatRequest{
		Model:       model,
		Messages:    messages,
		Temperature: temperature,
	})
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+l.apiKey)
	req.Header.Set("Content-Type", "application/json")

	res, err := l.client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		b, _ := io.ReadAll(res.Body)
		return "", fmt.Errorf("bad status %d: %s", res.StatusCode, string(b))
	}

	var cr chatResponse
	if err := json.NewDecoder(res.Body).Decode(&cr); err != nil {
		return "", err
	}
	if len(cr.Choices) == 0 {
		return "", fmt.Errorf("empty choices")
	}
	return cr.Choices[0].Message.Content, nil
}
