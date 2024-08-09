package main

import (
	"context"
	"crypto/rand"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
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

	// Echo(Graceful Shutdown)
	address := os.Getenv("ECHOADDRESS")
	if address == "" {
		address = ":1323"
	}
	e := handler.NewEcho(h)
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	go func() {
		if err := e.Start(address); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatal("shutting down the server")
		}
	}()
	<-ctx.Done()
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSec)*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
}
