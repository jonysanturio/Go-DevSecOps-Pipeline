package main // Es el paquete principal, se encarga de decirle al archivo que se ejecute y no es una libreria

import (
	"database/sql"
	"fmt" // Es una libreria estandar: "fmt" sirve para fomentar texto osea que aparezca en pantalla
	"log"
	"net/http" // Libreria para crear servidores web
	"os"
	"time"

	_ "github.com/lib/pq" // Importo el drive de Postgres
)

// Definimos el "Molde" de nuestro producto
// Las etiquetas `json:"..."` le dicen a Go cómo leer/escribir esto en internet.


func main(){
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	// 1. Cadena de Conexion 
	connStr := fmt.Sprintf("postgres://%s:%s@db:5432/%s?sslmode=disable", dbUser, dbPass, dbName)
	// 2. Abrir la conexión (Esto solo valida los parámetros)
	db, err := sql.Open("postgres", connStr)
	if err != nil{
		log.Fatal(err)
	}
	// 3. Reintentos de conexión (Wait-for-it)
	for i := 0; i < 10; i++ {
		err := db.Ping() // Ping intenta conectar de verdad
		if err != nil{
			break // Salgo del bucle
		}
		fmt.Println("Esperando datos.....")
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		log.Fatal("No se pudo conectar a la DB despues de varios intentos")
	}
	fmt.Println("¡Conexion exitosa!")
	// Llamo la funcion create table adentro de func main()
	createTable(db)
	//Llamo Rutas
	http.HandleFunc("/login", loginHandler)
	//Llamo ruta de Middleware
	http.HandleFunc("/products", func(w http.ResponseWriter, r *http.Request){
		if r.Method == http.MethodPost{
			createProductHandler(w, r, db)
		} else {
			getProductsHandler(w, r, db)
		}
	})
	// 5. Imprimir en consola para saber si anda el servidor
	fmt.Printf("Servidor corriendo...")
	// 6. Arracar el servidor (sabiendo que esta escuchando las interfaces)
	// ahora en docker si o si es fundamental poner el 0.0.0.0 y no "localhost"
	http.ListenAndServe("0.0.0.0:8080", nil)
}
