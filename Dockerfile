# Stage 1: Сборка приложения
FROM golang:latest AS build
COPY go.mod . 

RUN go mod download
COPY . .
RUN go build -o app ./cmd/filmlibrary/main.go


# CMD для запуска приложения
CMD ["./app"]
