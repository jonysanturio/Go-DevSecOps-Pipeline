# ETAPA 1: EL TALLER (Builder)
# Usamos la imagen oficial de Go para compilar
FROM golang:1.25-alpine AS builder

WORKDIR /app

# 1. Copiamos los archivos de dependencias primero (para aprovechar caché)
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# CGO_ENABLED=0: Desactiva dependencias de C (necesario para la imagen 'scratch')
# -o main: Nombre del ejecutable
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/api

# ETAPA 2: EL PRODUCTO FINAL (Runner)
# Usamos 'scratch': Una imagen VACÍA (0 MB). Máxima seguridad.
FROM scratch

COPY --from=builder /app/main /main

EXPOSE 8080


CMD ["/main"]