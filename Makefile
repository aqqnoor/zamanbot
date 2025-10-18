BASE_URL ?= https://openai-hub.neuraldeep.tech
KEY      ?= $(OPENAI_API_KEY)
export $(shell sed 's/=.*//' .env 2>/dev/null)
.PHONY: models ping embed transcribe go-run
models:
	curl -sS "$(BASE_URL)/v1/models" -H "Authorization: Bearer $(KEY)" | python -m json.tool

ping:
	curl -sS -X POST "$(BASE_URL)/v1/chat/completions" \
	  -H "Authorization: Bearer $(KEY)" \
	  -H "Content-Type: application/json" \
	  -d '{"model":"gpt-4o-mini","messages":[{"role":"user","content":"Скажи: pong"}],"temperature":0}' | python -m json.tool

embed:
	curl -sS -X POST "$(BASE_URL)/v1/embeddings" \
	  -H "Authorization: Bearer $(KEY)" \
	  -H "Content-Type: application/json" \
	  -d '{"model":"text-embedding-3-small","input":"Алға, мақсат!"}' | python -m json.tool
# Пример: make transcribe AUDIO=./sample.wav
transcribe:
	@if [ -z "$(AUDIO)" ]; then echo "Usage: make transcribe AUDIO=./file.wav"; exit 1; fi
	curl -sS -X POST "$(BASE_URL)/v1/audio/transcriptions" \
	  -H "Authorization: Bearer $(KEY)" \
	  -H "Content-Type: multipart/form-data" \
	  -F "model=whisper-1" \
	  -F "file=@$(AUDIO)" | python -m json.tool
go-run:
	cd backend && go run .
