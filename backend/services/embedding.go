package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type EmbeddingClient struct {
	baseURL string
	apiKey  string
	client  *http.Client
}

func NewEmbedding(baseURL, apiKey string) *EmbeddingClient {
	return &EmbeddingClient{
		baseURL: baseURL,
		apiKey:  apiKey,
		client:  &http.Client{Timeout: 60 * time.Second},
	}
}

type embeddingRequest struct {
	Model string      `json:"model"`
	Input interface{} `json:"input"`
}

type embeddingResponse struct {
	Data []struct {
		Embedding []float32 `json:"embedding"`
	} `json:"data"`
}

func (e *EmbeddingClient) Embed(input interface{}) ([]float32, error) {
	url := e.baseURL + "/v1/embeddings"
	payload, _ := json.Marshal(embeddingRequest{
		Model: "text-embedding-3-small",
		Input: input,
	})
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+e.apiKey)
	req.Header.Set("Content-Type", "application/json")

	res, err := e.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, fmt.Errorf("embeddings status %d", res.StatusCode)
	}

	var er embeddingResponse
	if err := json.NewDecoder(res.Body).Decode(&er); err != nil {
		return nil, err
	}
	if len(er.Data) == 0 {
		return nil, fmt.Errorf("empty embedding data")
	}
	return er.Data[0].Embedding, nil
}
