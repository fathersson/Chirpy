package config

import (
	"Chirpy/domain"
	"database/sql"
	"fmt"
	"log"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

/*type Handler struct {
	fileserverHits atomic.Int32
	db             *database.Queries
	platform       string
	tokenSecret    string
	polkaKey       string
}

// Конструктор для Config
func NewHandlerConfig() *Handler {
	return &Handler{
		fileserverHits: atomic.Int32{},   // Инициализация атомарной переменной
		db:             database.New(db), // Устанавливаем подключение к базе данных
		platform:       platform,         // Устанавливаем платформу
		tokenSecret:    tokenSecret,      // Устанавливаем секрет для токенов
		polkaKey:       polkaKey,         // Устанавливаем ключ для Polka
	}
}

type Service struct {
	fileserverHits atomic.Int32
	db             *database.Queries
	platform       string
	tokenSecret    string
	polkaKey       string
}

func NewServiceConfig(db *database.Queries, platform, tokenSecret, polkaKey string) *Service {
	return &Service{
		fileserverHits: atomic.Int32{}, // Инициализация атомарной переменной
		db:             db,             // Устанавливаем подключение к базе данных
		platform:       platform,       // Устанавливаем платформу
		tokenSecret:    tokenSecret,    // Устанавливаем секрет для токенов
		polkaKey:       polkaKey,       // Устанавливаем ключ для Polka
	}
}*/

/*type Service struct {
	fileserverHits atomic.Int32
	db             *database.Queries
	Platform       string
	TokenSecret    string
	PolkaKey       string
}*/

// Интерфейс для базы данных
type Database interface {
	Connect() error
}

func Connect() error {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	var env domain.Env
	err = cleanenv.ReadEnv(&env)
	if err != nil {
		log.Fatal(err)
	}
	// Открываем подключение к базе данных
	conn, err := sql.Open("postgres", env.DbUrl)
	if err != nil {
		return fmt.Errorf("failed to open database connection: %w", err)
	}
	defer conn.Close()

	// Проверяем подключение
	err = conn.Ping()
	if err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	fmt.Println("Connected to PostgreSQL database!")
	return nil
}

// Структура сервиса, которая использует интерфейс Database
type Service struct {
	db Database
}

// Конструктор сервиса, который принимает интерфейс Database
func NewService(db Database) *Service {
	return &Service{db: db}
}

//dbQueries := database.New(db)
//Cfg.db = dbQueries
//Cfg.platform = os.Getenv("PLATFORM")
//Cfg.tokenSecret = os.Getenv("TOKEN_SECRET")
//Cfg.polkaKey = os.Getenv("POLKA_KEY")
