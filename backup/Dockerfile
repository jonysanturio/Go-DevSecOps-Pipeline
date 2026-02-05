# 1. Imagen base: se usa la version oficial y que este vinculada a Go (Alpine)
FROM golang:1.22-alpine

# 2. Directorio de trabajo: se crea una carpeta dentro del contenedor
WORKDIR /app

# 3. Copia Inicial: se copia los archivo de la PC al contenedor
COPY . .

# 4. Comando Por Defecto: que pasa si solo se corre el contenedor sin de instrucciones
CMD ["go", "run", "."]

