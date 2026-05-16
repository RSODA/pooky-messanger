package main

import (
	"context"
	"log"
	"pooky-messanger/internal/auth"
	"pooky-messanger/internal/chat"
	"pooky-messanger/internal/user"
	"pooky-messanger/internal/ws"
	"pooky-messanger/pkg/config"
	"pooky-messanger/pkg/jwt"
	"pooky-messanger/pkg/middleware"
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
		"file://migrations",
		"postgres",
		driver,
	)

	if err != nil {
		log.Fatal(err)
	}

	if err = m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatal(err)
	}

	// token

	tokenCfg, err := config.NewTokenService()
	if err != nil {
		log.Fatal(err)
	}

	token := jwt.NewGetTokenRequest(tokenCfg.GetSecret(), tokenCfg.GetDuration())

	mid := middleware.NewMiddleware(token)

	// auth

	repoAuth := auth.NewAuthRepository(db)
	serviceAuth := auth.NewAuthService(repoAuth, token)
	handlerAuth := auth.NewAuthHandler(serviceAuth)

	v1auth := r.Group("/api/v1/auth")
	v1auth.POST("/login", handlerAuth.Login)
	v1auth.POST("/register", handlerAuth.Register)

	// user

	repoUser := user.NewUserRepository(db)
	serviceUser := user.NewUserService(repoUser)
	handlerUser := user.NewUserHandler(serviceUser)

	v1user := r.Group("/api/v1/users")
	v1user.GET("/me", mid.AuthMiddleware(), handlerUser.GetMe)
	v1user.GET("/:username", mid.AuthMiddleware(), handlerUser.GetUser)
	v1user.PUT("/me", mid.AuthMiddleware(), handlerUser.UpdateMe)

	// ws

	hub := ws.NewHub()
	r.GET("/api/v1/ws", mid.AuthMiddleware(), ws.WebSocketHandler(hub))

	go hub.Run()

	// chat

	repoChat := chat.NewRepository(db)
	serviceChat := chat.NewService(repoChat, hub)
	handlerChat := chat.NewChatHandler(serviceChat)

	v1chat := r.Group("/api/v1/conversations")
	v1chat.POST("", mid.AuthMiddleware(), handlerChat.CreateConversation)
	v1chat.GET("", mid.AuthMiddleware(), handlerChat.GetConversations)
	v1chat.GET("/:id/messages", mid.AuthMiddleware(), handlerChat.GetMessages)
	v1chat.POST("/:id/messages", mid.AuthMiddleware(), handlerChat.SendMessage)

	// start http

	err = r.Run(httpCfg.Address())
	if err != nil {
		log.Fatal(err)
	}
}
