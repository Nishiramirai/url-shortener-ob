package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"url-shortener-ob/internal/config"
	"url-shortener-ob/internal/handler"
	"url-shortener-ob/internal/service"
	"url-shortener-ob/internal/storage/memory"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.MustLoad()
	// TODO: убрать эту хуйню
	log.Printf("%v\n", cfg)

	// TODO: init logger

	// TODO: Сделать нормально с выбором хранилища
	repo := memory.New()

	// TODO: init service
	urlService := service.New(repo)

	// TODO: init handler
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	h := handler.New(urlService)
	h.RegisterRoutes(router)

	srv := &http.Server{
		Addr:         cfg.HTTPServer.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	go func() {
		// TODO: нормальный логгер
		log.Printf("HTTP server is running on %s", cfg.HTTPServer.Address)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			// TODO: нормальный логгер, фатал хуйня
			log.Fatalf("Listen and serve error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// TODO: нормальный логгер
	log.Println("Shutting down server gracefully...")

	ctxShutdown, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutdown()

	if err := srv.Shutdown(ctxShutdown); err != nil {
		// TODO: нормальный логгер
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	// TODO: нормальный логгер
	log.Println("Server exited properly")
}
