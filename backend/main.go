package main

import (
	"fmt"
	"log"
	"os"

	"zmb-assistant/services"

	"github.com/joho/godotenv"
)

func mustGetEnv(k string) string {
	v := os.Getenv(k)
	if v == "" {
		log.Fatalf("%s is empty; set it in .env or env", k)
	}
	return v
}

func main() {
	_ = godotenv.Load(".env", "../.env")
	baseURL := os.Getenv("OPENAI_BASE_URL")
	if baseURL == "" {
		baseURL = "https://openai-hub.neuraldeep.tech"
	}
	apiKey := mustGetEnv("OPENAI_API_KEY")

	llm := services.NewLLM(baseURL, apiKey)
	reply, err := llm.Chat("gpt-4o-mini",
		[]services.ChatMessage{{Role: "user", Content: "Ответь одним словом: pong"}}, 0)
	if err != nil {
		log.Fatalf("LLM error: %v", err)
	}
	fmt.Println("LLM reply:", reply)

	startServer()
}
