package main

import (
	"advanced-blog-management-system/internal/handler"
	"advanced-blog-management-system/internal/middleware"
	"advanced-blog-management-system/internal/repository"
	"advanced-blog-management-system/internal/service"
	"advanced-blog-management-system/pkg/database"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"strconv"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
)

type config struct {
	dbHost     string
	dbPort     int
	dbUser     string
	dbPassword string
	dbName     string
	dbSSLMode  string

	jwtSecret string

	serverHost string
	serverPort string
}

func loadConfig() config {
	dbPort, err := strconv.Atoi(getEnv("DB_PORT", "5432"))
	if err != nil {
		dbPort = 5432
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET environment variable is required")
	}

	return config{
		dbHost:     getEnv("DB_HOST", "localhost"),
		dbPort:     dbPort,
		dbUser:     getEnv("DB_USER", "postgres"),
		dbPassword: getEnv("DB_PASSWORD", "postgres"),
		dbName:     getEnv("DB_NAME", "blog_db"),
		dbSSLMode:  getEnv("DB_SSLMODE", "disable"),

		jwtSecret: jwtSecret,

		serverHost: getEnv("SERVER_HOST", "0.0.0.0"),
		serverPort: getEnv("SERVER_PORT", "8080"),
	}
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

// readMigrations читает все .sql файлы из папки migrations/ в порядке имён
// и возвращает их содержимое как срез строк для database.RunMigrations
func readMigrations() ([]string, error) {
	entries, err := os.ReadDir("migrations")
	if err != nil {
		return nil, err
	}

	var names []string
	for _, e := range entries {
		if !e.IsDir() && filepath.Ext(e.Name()) == ".sql" {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)

	migrations := make([]string, 0, len(names))
	for _, name := range names {
		content, err := os.ReadFile(filepath.Join("migrations", name))
		if err != nil {
			return nil, err
		}
		migrations = append(migrations, string(content))
	}
	return migrations, nil
}

const scheduledPostsCheckInterval = 30 * time.Second

// runScheduler периодически публикует черновики, время публикации которых наступило
// Работает в отдельной горутине до отмены ctx (вызывается при graceful shutdown)
// ticker.C и ctx.Done() — оба каналы, select между ними — стандартный
// Go-паттерн ожидания "что наступит раньше"
func runScheduler(ctx context.Context, postService *service.PostService) {
	ticker := time.NewTicker(scheduledPostsCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			published, err := postService.PublishScheduledPosts(ctx)
			if err != nil {
				log.Printf("scheduler: failed to publish scheduled posts: %v", err)
				continue
			}
			if published > 0 {
				log.Printf("scheduler: published %d scheduled post(s)", published)
			}
		case <-ctx.Done():
			log.Println("scheduler: stopped")
			return
		}
	}
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, relying on environment variables")
	}

	cfg := loadConfig()
	log.Println("configuration loaded")

	db, err := database.NewPostgresDB(database.Config{
		Host:     cfg.dbHost,
		Port:     cfg.dbPort,
		User:     cfg.dbUser,
		Password: cfg.dbPassword,
		DBName:   cfg.dbName,
		SSLMode:  cfg.dbSSLMode,
	})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	log.Println("connected to database")

	migrations, err := readMigrations()
	if err != nil {
		log.Fatalf("failed to read migrations: %v", err)
	}
	if err := database.RunMigrations(db, migrations); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}
	log.Printf("applied %d migration(s)", len(migrations))

	userRepo := repository.NewUserRepository(db)
	postRepo := repository.NewPostRepository(db)
	commentRepo := repository.NewCommentRepository(db)

	userService := service.NewUserService(userRepo)
	postService := service.NewPostService(postRepo, userRepo, commentRepo)
	commentService := service.NewCommentService(commentRepo, postRepo, userRepo)

	authHandler := handler.NewAuthHandler(userService, cfg.jwtSecret)
	postHandler := handler.NewPostHandler(postService)
	commentHandler := handler.NewCommentHandler(commentService)

	router := setupRouter(authHandler, postHandler, commentHandler)

	schedulerCtx, stopScheduler := context.WithCancel(context.Background())
	go runScheduler(schedulerCtx, postService)

	srv := &http.Server{
		Addr:         cfg.serverHost + ":" + cfg.serverPort,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("server starting on http://%s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	log.Println("shutdown signal received")

	stopScheduler()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("forced server shutdown: %v", err)
	}

	if err := db.Close(); err != nil {
		log.Printf("failed to close database connection: %v", err)
	}

	log.Println("server stopped cleanly")
}

func setupRouter(
	authHandler *handler.AuthHandler,
	postHandler *handler.PostHandler,
	commentHandler *handler.CommentHandler,
) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.LoggingMiddleware)
	r.Use(middleware.RecoveryMiddleware)
	r.Use(middleware.CORSMiddleware)

	r.Route("/api", func(api chi.Router) {
		// Публичные эндпоинты
		api.Get("/health", handler.HealthCheckHandler)
		api.Post("/register", authHandler.RegisterHandler)
		api.Post("/login", authHandler.LoginHandler)

		api.Get("/posts", postHandler.GetAllPosts)
		api.Get("/posts/{id}", postHandler.GetPost)
		api.Get("/users/{authorID}/posts", postHandler.GetPostsByAuthor)
		api.Get("/posts/{id}/comments", commentHandler.GetCommentsByPostID)

		// Защищенные эндпоинты
		api.Group(func(protected chi.Router) {
			protected.Use(middleware.AuthMiddleware)

			protected.Post("/posts", postHandler.CreatePost)
			protected.Put("/posts/{id}", postHandler.UpdatePost)
			protected.Delete("/posts/{id}", postHandler.DeletePost)

			protected.Post("/posts/{id}/comments", commentHandler.CreateComment)
			protected.Put("/comments/{id}", commentHandler.UpdateComment)
			protected.Delete("/comments/{id}", commentHandler.DeleteComment)
		})
	})

	return r
}
