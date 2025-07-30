FROM golang:1.24.4-alpine
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o ./chat-server ./cmd/chat-server/main.go
EXPOSE 8080
ENTRYPOINT [ "./chat-server" ]