package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"log"
	"log/slog"
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
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	// リクエスト処理フロー
	// Echo -> Handler -> Service -> Repository -> Database

	// 初期化は逆順

	// Repository
	dburl := os.Getenv("DBURL")
	if dburl == "" {
		log.Fatal("DBURL is empty")
	}
	sqlDB, err := sql.Open("pgx", dburl)
	if err != nil {
		log.Fatal(err)
	}
	defer sqlDB.Close()
	r, err := repository.NewRepository(logger, sqlDB)
	if err != nil {
		log.Fatal(err)
	}
	if err := r.InitDb(context.Background()); err != nil {
		log.Fatal(err)
	}

	// Service
	s := service.NewService(logger, r)

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
	location, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		log.Fatal(err)
	}
	const timeoutSec = 60
	h := handler.NewHandler(logger, s, &handler.HandlerConfig{
		JwtKey:      jwtkey,
		ValidityMin: 120, // JWTのexp
		Location:    location,
		Locale:      "en",
		Indent:      "  ", // レスポンスのJSONのインデント
		TimeoutSec:  timeoutSec,
	})

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
