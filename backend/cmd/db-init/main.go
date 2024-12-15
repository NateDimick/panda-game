package main

import (
	"pandagame/internal/config"
)

func main() {
	// create tables in surreal db
	config.LoadAppConfig()
	db, err := config.AdminSurreal()
	if err != nil {
		panic(err)
	}
	surrealdb.
}
