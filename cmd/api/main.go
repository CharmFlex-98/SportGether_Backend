package main // "main" indicating this is main app, not library. Other files cannot import this.
import (
	"context"
	"database/sql"
	"flag"
	"log/slog"
	"os"
	"time"
)

type config struct {
	port     int
	env      string
	dbConfig struct {
		dsn               string
		maxOpenConnection int
		maxIdleConnection int
		maxIdleTime       time.Duration
	}
}

type Application struct {
	config   config
	logger   *slog.Logger
	database *sql.DB
}

func main() {
	config := config{}
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	flag.IntVar(&config.port, "port", 3000, "Port for listening")
	flag.StringVar(&config.env, "environment", "dev", "environment dev|prd")

	flag.Parse()

	db, err := config.openDatabase()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	app := Application{
		config:   config,
		logger:   logger,
		database: db,
	}

	app.serve()
}

func (cfg config) openDatabase() (*sql.DB, error) {
	db, err := sql.Open("postgres", os.Getenv("DATABASE_UR:"))
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.dbConfig.maxOpenConnection)
	db.SetMaxIdleConns(cfg.dbConfig.maxIdleConnection)
	db.SetConnMaxIdleTime(cfg.dbConfig.maxIdleTime)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	return db, nil
}
