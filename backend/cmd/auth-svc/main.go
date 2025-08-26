package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/spf13/viper"

	"yunai/internal/repository"
	"yunai/internal/service"
	"yunai/pkg/auth"
	"yunai/pkg/database"
	"yunai/pkg/logger"
	"yunai/pkg/redis"
)

func main() {
	// 初始化配置
	initConfig()

	// 初始化日志
	logger := logger.New()

	// 初始化数据库
	db, err := database.NewPostgreSQL()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// 初始化 Redis
	rdb, err := redis.NewClient()
	if err != nil {
		log.Fatal("Failed to connect to Redis:", err)
	}
	defer rdb.Close()

	// 初始化 JWT 管理器
	jwtManager := auth.NewManager(
		viper.GetString("jwt.secret"),
		viper.GetDuration("jwt.access_token_duration"),
		viper.GetDuration("jwt.refresh_token_duration"),
	)

	// 初始化仓库层
	userRepo := repository.NewUserRepository(db)
	authRepo := repository.NewAuthRepository(rdb)

	// 初始化服务层
	authService := service.NewAuthService(userRepo, authRepo, jwtManager, logger)

	// 初始化传输层
	httpAuthHandler := authHandler.NewHandler(authService, logger)

	// 创建路由
	r := chi.NewRouter()

	// 中间件
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.Timeout(60 * time.Second))

	// CORS 配置
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// 健康检查
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// API 路由
	r.Route("/api/v1", func(r chi.Router) {
		r.Mount("/auth", httpAuthHandler.Routes())
	})

	// 启动服务器
	port := viper.GetString("server.port")
	if port == "" {
		port = "8001"
	}

	server := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	// 优雅关闭
	go func() {
		logger.Info(fmt.Sprintf("Auth service starting on port %s", port))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Server failed to start:", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// 创建超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 关闭服务器
	if err := server.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown:", err)
	}

	logger.Info("Server exited")
}

func initConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath(".")

	// 环境变量
	viper.AutomaticEnv()

	// 默认配置
	viper.SetDefault("server.port", "8001")
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.name", "yunai")
	viper.SetDefault("database.user", "postgres")
	viper.SetDefault("database.password", "password")
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.db", 0)
	viper.SetDefault("jwt.secret", "yunai-secret-key")
	viper.SetDefault("jwt.access_token_duration", "1h")
	viper.SetDefault("jwt.refresh_token_duration", "720h")

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Config file not found, using defaults: %v", err)
	}
}
