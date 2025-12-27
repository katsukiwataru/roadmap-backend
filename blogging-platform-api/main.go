package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	if err := connectDB(); err != nil {
		log.Fatalf("DB connection failed: %v", err)
	}
	defer db.Close()
	log.Println("DB connected!")

	if err := createTables(); err != nil {
		log.Fatalf("Failed to create tables: %v", err)
	}
	log.Println("Tables ready!")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, Blog API! DB is connected.")
	})

	http.HandleFunc("/posts", handlePosts)
	http.HandleFunc("/posts/", handlePost)

	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
