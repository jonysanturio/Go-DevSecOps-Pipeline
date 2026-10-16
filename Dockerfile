FROM golang:1.27.2-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download


COPY . .


RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/api


FROM scratch

COPY --from=builder /app/main /main

USER 65532:65532

EXPOSE 8080


CMD ["/main"]