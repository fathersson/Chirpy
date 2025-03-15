package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"
)

func main() {
	//Cfg := config.NewConfig()
	//подключение к базе
	/*err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	var env domain.Env
	err = cleanenv.ReadEnv(&env)
	if err != nil {
		log.Fatal(err)
	}*/
	/*
		db, err := sql.Open("postgres", env.DbUrl)
		if err != nil {
			log.Fatal(err)
		}
		err = db.Ping()
		if err != nil {
			log.Fatal("Error connecting to the database:", err)
		}
		defer db.Close()
	*/

	//dbQueries := database.New(db)
	//Cfg.db = dbQueries
	//Cfg.platform = os.Getenv("PLATFORM")
	//Cfg.tokenSecret = os.Getenv("TOKEN_SECRET")
	//Cfg.polkaKey = os.Getenv("POLKA_KEY")

	port := "8080"
	mux := http.NewServeMux()
	serv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	mux.Handle("/app/", middlewareMetricsInc(http.StripPrefix("/app/", http.FileServer(http.Dir(".")))))
	mux.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	mux.HandleFunc("GET /admin/metrics", metricsHandler)
	mux.HandleFunc("POST /admin/reset", apiCfg.resetHandler)
	mux.HandleFunc("POST /api/chirps", apiCfg.chirpHandler)
	mux.HandleFunc("POST /api/users", apiCfg.createUserHandler)
	mux.HandleFunc("GET /api/chirps", apiCfg.getChirpsHandler)
	mux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.getOneChirpHandler)
	mux.HandleFunc("POST /api/login", apiCfg.loginHandler)
	mux.HandleFunc("POST /api/refresh", apiCfg.refreshHandler)
	mux.HandleFunc("POST /api/revoke", apiCfg.revokeHandler)
	mux.HandleFunc("PUT /api/users", apiCfg.putUsersHandler)
	mux.HandleFunc("DELETE /api/chirps/{chirpID}", apiCfg.deleteChirpHandler)
	mux.HandleFunc("POST /api/polka/webhooks", apiCfg.webhooksHandler)

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
