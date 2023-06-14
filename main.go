package main

import (
	"crypto/rand"
	"log"
	"os"
	"time"

	"github.com/ystkg/rest-example/handler"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	dburl := os.Getenv("DBURL")
	if dburl == "" {
		log.Fatal("DBURL is empty")
	}
	db, err := gorm.Open(postgres.Open(dburl), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	jwtkeyStr := os.Getenv("JWTKEY")
	var jwtkey []byte
	if jwtkeyStr != "" {
		jwtkey = []byte(jwtkeyStr)
	} else {
		jwtkey = make([]byte, 32)
		_, err := rand.Read(jwtkey)
		if err != nil {
			log.Fatal(err)
		}
	}

	location, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		log.Fatal(err)
	}

	h := handler.NewHandler(db, jwtkey, location)
	if err := h.InitDB(); err != nil {
		log.Fatal(err)
	}

	e := handler.NewEcho(h)

	address := os.Getenv("ECHOADDRESS")
	if address == "" {
		address = ":1323"
	}
	e.Logger.Fatal(e.Start(address))
}
