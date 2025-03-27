package main

import (
	"Chirpy/internal/config"
	"Chirpy/internal/database"
	"Chirpy/internal/handler"
	"context"
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	Cfg := config.NewConfig()
	//подключение к базе
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	//var env config.Config
	err = cleanenv.ReadEnv(&Cfg)
	if err != nil {
		log.Fatal(err)
	}
	db, err := sql.Open("postgres", Cfg.DBURL)
	if err != nil {
		slog.Error("failed to open database connection", "err", err)
	}
	defer db.Close()

	// Проверяем подключение
	err = db.Ping()
	if err != nil {
		slog.Error("failed to ping database:", "err", err)
	}

	fmt.Println("Connected to PostgreSQL database!")

	dbQueries := database.New(db) //repository
	//service := handler.NewService(dbQueries)
	//Service := handler.NewService()
	//var b handler.ServiceStr
	//service := handler.NewService(b.Service)
	var service handler.ServiceInt
	Handler := handler.NewHandler(dbQueries, service)

	port := "8080"
	mux := http.NewServeMux()
	serv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	mux.Handle("/app/", handler.MiddlewareMetricsInc(Cfg, http.StripPrefix("/app/", http.FileServer(http.Dir(".")))))
	mux.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	mux.HandleFunc("GET /admin/metrics", Handler.MetricsHandler)
	mux.HandleFunc("POST /admin/reset", Handler.ResetHandler)
	mux.HandleFunc("POST /api/chirps", Handler.ChirpHandler)
	mux.HandleFunc("POST /api/users", Handler.CreateUserHandler)
	mux.HandleFunc("GET /api/chirps", Handler.GetChirpsHandler)
	mux.HandleFunc("GET /api/chirps/{chirpID}", Handler.GetOneChirpHandler)
	mux.HandleFunc("POST /api/login", Handler.LoginHandler)
	mux.HandleFunc("POST /api/refresh", Handler.RefreshHandler)
	mux.HandleFunc("POST /api/revoke", Handler.RevokeHandler)
	mux.HandleFunc("PUT /api/users", Handler.PutUsersHandler)
	mux.HandleFunc("DELETE /api/chirps/{chirpID}", Handler.DeleteChirpHandler)
	mux.HandleFunc("POST /api/polka/webhooks", Handler.WebhooksHandler)

	// Канал для получения сигналов завершения
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, syscall.SIGINT, syscall.SIGTERM)

	// запускает HTTP-сервер и начинает слушать указанный адрес и порт
	go func() {
		log.Fatal(serv.ListenAndServe())
	}()

	// Ожидаем получения сигнала завершения
	sig := <-signalChannel
	log.Printf("Received signal %s, shutting down gracefully...", sig)

	// Создаем контекст с тайм-аутом для graceful shutdown (например, 10 секунд)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Останавливаем сервер, завершая текущие запросы
	err = serv.Shutdown(ctx)
	if err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}
}
