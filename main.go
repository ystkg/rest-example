package main

import (
	"crypto/rand"
	"log"
	"os"
	"time"

	"github.com/ystkg/rest-example/handler"
)

func main() {
	dburl := os.Getenv("DBURL")
	if dburl == "" {
		log.Fatal("DBURL is empty")
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

	h := handler.NewHandler(dburl, jwtkey, location)

	e, err := handler.NewEcho(h)
	if err != nil {
		log.Fatal(err)
	}

	address := os.Getenv("ECHOADDRESS")
	if address == "" {
		address = ":1323"
	}
	e.Logger.Fatal(e.Start(address))
}
