package main

import (
	"log"
	"os"
	"strings"

	"github.com/jgkawell/yarr/server"
	"github.com/jgkawell/yarr/storage"
)

var Version string = "0.0"
var GitHash string = "unknown"

type config struct {
	address  string
	username string
	password string
	database string
}

func main() {
	c := readConfig()

	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.SetOutput(os.Stdout)

	store, err := storage.New(c.database)
	if err != nil {
		log.Fatal("Failed to initialise database: ", err)
	}

	srv := server.NewServer(store, c.address)
	srv.Username = c.username
	srv.Password = c.password

	log.Printf("starting server at %s", srv.GetAddr())
	srv.Start()
}

func readConfig() config {
	c := config{
		address:  os.Getenv("YARR_ADDR"),
		database: os.Getenv("YARR_DB"),
	}
	if c.address == "" {
		c.address = "0.0.0.0:7070"
	}
	if c.database == "" {
		log.Fatal("Error YARR_DB is not set. Please set env var YARR_DB to postgres URI.")
	}
	auth := os.Getenv("YARR_AUTH")
	if auth != "" {
		parts := strings.SplitN(auth, ":", 2)
		if len(parts) != 2 {
			log.Fatal("auth was set but with wrong syntax (expected `username:password`)")
		}
		c.username = parts[0]
		c.password = parts[1]
	}
	return c
}
