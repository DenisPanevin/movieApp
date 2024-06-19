package main

import (
	"context"
	"database/sql"
	"flag"
	_ "github.com/lib/pq"
	"os"
	"secondBook/internal/data"
	"secondBook/internal/jsonLog"
	"secondBook/internal/mailer"
	"sync"
	"time"
)

const version = "1.0.0"

type config struct {
	port int
	env  string
	db   struct {
		dsn         string
		maxOpenCons int
		maxIdleCons int
		maxIdelTime string
	}
}

type application struct {
	config config
	logger *jsonLog.Logger
	models data.Models
	mailer mailer.Mailer
	wg     sync.WaitGroup
}

func main() {
	var cfg config

	//log.Print(os.Getenv("GREENLIGHT_DB_DSN"))
	//fmt.Sprintf("dbname=%s sslmode=disable", os.Getenv("GREENLIGHT_DB_DSN"))

	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "dev", "dev|prod|stage")
	flag.StringVar(&cfg.db.dsn, "db-dsn", "postgres://greenlight:pa55word@localhost/greenlight?sslmode=disable", "PostgreSQL DSN")

	flag.IntVar(&cfg.db.maxOpenCons, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleCons, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.StringVar(&cfg.db.maxIdelTime, "db-max-idle-time", "15m", "PostgreSQL max connection idle time")

	flag.Parse()

	logger := jsonLog.New(os.Stdout, jsonLog.LevelInfo)

	db, err := openDB(cfg)
	if err != nil {
		logger.PrintFatal(err, nil)
	}

	defer db.Close()
	logger.PrintInfo("database connection pool established", nil)

	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
		mailer: mailer.New("sandbox.smtp.mailtrap.io", 2525, "762e10140df732", "92e91085dd3d78", "ACME <no-reply@ACMEt.net>"),
	}

	err = app.Serve()
	logger.PrintFatal(err, nil)
}

func openDB(cfg config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(cfg.db.maxOpenCons)
	db.SetMaxIdleConns(cfg.db.maxIdleCons)
	duration, err := time.ParseDuration(cfg.db.maxIdelTime)
	if err != nil {
		return nil, err
	}
	db.SetConnMaxIdleTime(duration)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}
	return db, nil
}
