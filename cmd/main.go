package main

import (
	"fmt"
	"strconv"

	"log"
	"net/http"
	"os"
	"time"

	"banner/internal/banner/controller"
	"banner/internal/banner/repository"
	"banner/internal/banner/service"
	db "banner/internal/infrastructure/db"
	"banner/internal/responder"
	"banner/internal/router"

	"github.com/jackc/pgx/v5/stdlib"
	goose "github.com/pressly/goose/v3"
)

func main() {
	Run()
}

func Run() {
	pool := db.NewPostgresClient()
	defer pool.Close()

	sqldb := stdlib.OpenDBFromPool(pool)
	defer sqldb.Close()
	if err := goose.Up(sqldb, "internal/infrastructure/migrations"); err != nil {
		log.Fatalf("goose up: %v", err)
	}

	repo := repository.NewPostgresRepository(pool)
	numShards, flushInterval := getConfig()
	service := service.NewBannerService(repo, numShards, flushInterval)
	defer service.Close()

	rsp := responder.NewResponder()
	ctrl := controller.NewBannerController(service, rsp)

	router := router.NewRouter(ctrl)

	host, port := os.Getenv("HOST"), os.Getenv("PORT")
	server := http.Server{Addr: fmt.Sprintf("%s:%s", host, port), Handler: router}
	log.Printf("server start on %s", server.Addr)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal("server error: ", err)
	}
}

func getConfig() (numShards int, flushInterval time.Duration) {
	numShards, err := strconv.Atoi(os.Getenv("NUM_SHARDS"))
	if err != nil {
		log.Fatalf("NUM_SHARDS error: %v", err)
	}

	interval, err := strconv.Atoi(os.Getenv("FLUSH_INTERVAL"))
	if err != nil {
		log.Fatalf("FLUSH_INTERVAL error: %v", err)
	}
	flushInterval = time.Duration(interval) * time.Second

	return numShards, flushInterval
}
