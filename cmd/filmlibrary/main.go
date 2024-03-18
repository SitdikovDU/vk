package main

import (
	"database/sql"
	"filmlibrary/pkg/config"
	"filmlibrary/pkg/explorer"
	"log"
	"net/http"

	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

// @title Film Library API
// @version 1.0
// @description Film Library application

// @host localhost:8080
// @BasePath /
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
func main() {
	zapLogger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("failed to create logger: %v", err)
		return
	}

	defer zapLogger.Sync()

	logger := zapLogger.Sugar()

	databaseConfig, err := config.NewDatabase()
	if err != nil {
		logger.Fatal("failed to init database config", zap.Error(err))
		return
	}

	/*connStr := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable",
	user,
	password,
	host,
	dbname)*/
	logger.Info(databaseConfig.DataSourceName, databaseConfig.DriverName)
	db, err := sql.Open(databaseConfig.DriverName, databaseConfig.DataSourceName)
	if err != nil {
		logger.Error("failed to connect to database", zap.Error(err))
		return
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		logger.Error("ping: failed to connect to database", zap.Error(err))
		return
	}
	logger.Info("Successfully connected to database!")

	serverConfig, err := config.NewServer()
	if err != nil {
		logger.Fatal("failed to init server config", zap.Error(err))
		return
	}

	logger.Infow("starting server",
		"type", "START",
		"addr", serverConfig.Addr,
	)

	handler, err := explorer.NewExplorer(db, logger)
	if err != nil {
		logger.Error("failed", zap.Error(err))
		return
	}

	if err := http.ListenAndServe(serverConfig.Addr, handler); err != nil {
		logger.Error("failed to start server", zap.Error(err))
		return
	}
}
