package main

import (
	"context"
	"crypto/rand"
	"log"
	"os"
	"time"

	"github.com/ystkg/rest-example/handler"
	"github.com/ystkg/rest-example/repository"
	"github.com/ystkg/rest-example/service"
)

func main() {
	// リクエスト処理フロー
	// Echo -> Handler -> Service -> Repository -> Database

	// 初期化は逆順

	// Repository
	dburl := os.Getenv("DBURL")
	if dburl == "" {
		log.Fatal("DBURL is empty")
	}
	r, err := repository.NewRepository(dburl)
	if err != nil {
		log.Fatal(err)
	}
	if err := r.InitDb(context.Background()); err != nil {
		log.Fatal(err)
	}

	// Service
	s := service.NewService(r)

	// Handler
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
	validityMin := 120 // JWTのexp
	location, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		log.Fatal(err)
	}
	indent := "  " // レスポンスのJSONのインデント
	timeoutSec := 60
	h := handler.NewHandler(s, jwtkey, validityMin, location, indent, timeoutSec)

	// Echo
	address := os.Getenv("ECHOADDRESS")
	if address == "" {
		address = ":1323"
	}
	e := handler.NewEcho(h)
	e.Logger.Fatal(e.Start(address))
}
