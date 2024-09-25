package main

import (
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

// use godot package to load/read the .env file and
// return the value of the key
func goDotEnvVariable(key string) string {

	// load .env file
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	return os.Getenv(key)
}

func main() {
	godotenv.Load()
	mux := http.NewServeMux()
	port := goDotEnvVariable("PORT")
	server := http.Server{
		Handler: mux,
		Addr:    ":" + port,
	}
	mux.Handle("/app/", http.StripPrefix("/app", http.FileServer(http.Dir("."))))
	mux.HandleFunc("/healthz", handleHealthz)
	log.Printf("Server run succesffuly on port: %s\n", port)
	log.Fatal(server.ListenAndServe())
}

func handleHealthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
