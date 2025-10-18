# build
FROM golang:1.22 as builder
WORKDIR /app
COPY . .
RUN go mod tidy && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server .

# run
FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=builder /app/server /app/server
COPY ./data /app/data
ENV PORT=8080 DATA_DIR=/app/data
EXPOSE 8080
USER 65532:65532
ENTRYPOINT ["/app/server"]
