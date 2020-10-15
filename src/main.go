package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/rubbenpad/gofood/app"
	"github.com/rubbenpad/gofood/routes"
	"github.com/rubbenpad/gofood/store"
)

func init() {
	if err := godotenv.Load("../.env"); err != nil {
		log.Print("No .env file found")
	}
}

func main() {

	// This allows sync dgraph schema
	setupdb, _ := os.LookupEnv("SETUP_DB")
	fmt.Println(setupdb)
	if setupdb == "yes" {
		dgraph := store.New()
		dgraph.Setup()
	}

	// Start app and pass it as parameter to api's
	ap := app.New()
	routes.LoadDataAPI(ap)

	log.Fatal(http.ListenAndServe(":3000", ap))
}
