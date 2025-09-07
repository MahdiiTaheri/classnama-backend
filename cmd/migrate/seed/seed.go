package main

import (
	"log"

	"github.com/MahdiiTaheri/classnama-backend/internal/db"
	"github.com/MahdiiTaheri/classnama-backend/internal/env"
	"github.com/MahdiiTaheri/classnama-backend/internal/store"
)

func main() {
	addr := env.GetString("DB_ADDR", "postgres://admin:adminpassword@localhost/socialapp?sslmode=disable")
	conn, err := db.New(addr, 3, 3, "15m")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	storage := store.NewStorage(conn)

	// Seed database with initial data
	db.Seed(storage)
	log.Println("Database seeding finished successfully!")
}
