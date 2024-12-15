package main

import (
	"fmt"
	"os"
	"pandagame/internal/config"

	"github.com/surrealdb/surrealdb.go"
	"github.com/surrealdb/surrealdb.go/pkg/models"
)

type Demo struct {
	ID  *models.RecordID `json:"id"`
	Foo string           `json:"foo"`
}

func main() {
	// This is just a playaround script to debug surreal sdk quirks
	os.Setenv("DB_DB", "pandaDB")
	os.Setenv("DB_NS", "pandaNS")
	os.Setenv("DB_ADMIN", "pandaAdmin")
	os.Setenv("DB_PASS", "strongpassword")
	os.Setenv("DB_ADDR", "ws://localhost:8000")
	config.LoadAppConfig()
	db, err := config.AdminSurreal()
	if err != nil {
		fmt.Println("db connect error", err.Error())
		os.Exit(1)
	}
	d := Demo{
		ID: &models.RecordID{
			ID:    "helloworld",
			Table: "demo",
		},
		Foo: "loobar",
	}
	emo, err := surrealdb.Upsert[[]Demo](db, models.Table("demo"), &d)
	if err != nil {
		fmt.Println("upsert error", err.Error())
		os.Exit(1)
	}
	fmt.Printf("%#v, %T\n", emo, emo)

	result, err := surrealdb.Select[Demo](db, *d.ID)
	if err != nil {
		fmt.Println("db select error", err.Error())
		os.Exit(1)
	}
	fmt.Printf("%#v, %T\n", result, result)

	del, err := surrealdb.Delete[Demo](db, *d.ID)
	if err != nil {
		fmt.Println("delete error", err.Error())
		os.Exit(1)
	}
	fmt.Printf("%#v, %T\n", del, del)
	// db.Let("name", "gopher")
	// db.Let("password", "gogogo")
	// token, err := db.SignUp(&surrealdb.Auth{
	// 	Access: "player",
	// })
	// if err != nil {
	// 	fmt.Println("signup error", err.Error())
	// }
	// fmt.Println(token)
}
