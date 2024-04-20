package main

import (
	"context"
	"database/sql"
	"flag"
	"log"
	"os"
	"sync"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/mayank12gt/movie-webapp/internal/data"
	"github.com/mayank12gt/movie-webapp/internal/mailer"
)

const version = "1.0.0"

type config struct {
	port int
	env  string
	db   struct {
		dsn string
	}

	smtp struct {
		host     string
		port     int
		username string
		password string
		sender   string
	}
}

type app struct {
	config config
	logger *log.Logger
	models data.Models
	mailer mailer.Mailer
	wg     sync.WaitGroup
}

func main() {

	godotenv.Load()
	var cfg config

	flag.IntVar(&cfg.port, "port", 4000, "API Server Port")
	flag.StringVar(&cfg.env, "env", "dev", "Env(dev|staging|prod)")

	flag.StringVar(&cfg.db.dsn, "db-dsn", os.Getenv("DSN"), "PostgreSQL DSN")

	flag.StringVar(&cfg.smtp.host, "smtp-host", "sandbox.smtp.mailtrap.io", "SMTP host")
	flag.IntVar(&cfg.smtp.port, "smtp-port", 25, "SMTP port")
	flag.StringVar(&cfg.smtp.username, "smtp-username", "a5e17edb57d773", "SMTP username")
	flag.StringVar(&cfg.smtp.password, "smtp-password", "1de789a2d550ae", "SMTP password")
	flag.StringVar(&cfg.smtp.sender, "smtp-sender", "MoviesDB <no-reply@moviesdb.net>", "SMTP sender")

	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	flag.Parse()

	db, err := openDB(cfg)
	if err != nil {
		logger.Fatal(err)
	}
	defer db.Close()
	logger.Printf("DB connected")

	app := &app{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
		mailer: mailer.New(cfg.smtp.host, cfg.smtp.port, cfg.smtp.username, cfg.smtp.password, cfg.smtp.sender),
	}

	// server := &http.Server{
	// 	Addr: fmt.Sprintf(":%d",cfg.port),
	// 	Handler: app.routes(),
	// }

	// server := echo.New()

	// server.POST("/movies", app.createMovieHandler())
	// server.GET("/movies", app.listMovieHandler())
	// server.GET("/movies/:id", app.getMovieHandler())
	// server.DELETE("/movies/:id", app.deleteMovieHandler())

	// server.GET("/", func(c echo.Context) error {
	// 	return c.JSON(200, "hello")
	// })

	// app.logger.Print(app.config.port)
	// server.Logger.Fatal(server.Start(":3000"))

	err = app.serve()
	if err != nil {
		logger.Printf(err.Error())
	}

}

func openDB(cfg config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	return db, nil

}
