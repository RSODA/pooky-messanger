package main

import (
	"context"
	"log"
	"pooky-messanger/pkg/config"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
)

func main() {
	ctx := context.Background()

	err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	r := gin.Default()

	httpCfg, err := config.NewHTTPConfig()
	if err != nil {
		log.Fatal(err)
	}

	dbCfg, err := config.NewPGConfig()
	if err != nil {
		log.Fatal(err)
	}

	dbCtx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()
	db, err := pgxpool.New(dbCtx, dbCfg.DSN())
	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping(dbCtx)
	if err != nil {
		log.Fatal(err)
	}

	sqlDB := stdlib.OpenDBFromPool(db)

	driver, err := postgres.WithInstance(sqlDB, &postgres.Config{})
	if err != nil {
		log.Fatal(err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations", // ← префикс file:// обязателен
		"postgres",
		driver,
	)

	if err != nil {
		log.Fatal(err)
	}

	if err = m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatal(err)
	}

	err = r.Run(httpCfg.Address())
	if err != nil {
		log.Fatal(err)
	}

}
