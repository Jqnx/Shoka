package main

import (
	"fmt"
	"log"
	"net/http"

	"Shoka/internal/api"
	"Shoka/internal/db"
	"Shoka/internal/db/sqlc"
)

func main() {
	db, err := db.New()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	fmt.Print("Connected to database.")

	queries := sqlc.New(db)

	api := api.New(queries)
	http.ListenAndServe(":8080", api.Router)
}
