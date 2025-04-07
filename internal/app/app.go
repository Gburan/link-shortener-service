package app

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"

	"link-shortener-service/internal/config"
	"link-shortener-service/internal/handler/expander_url"
	"link-shortener-service/internal/handler/shorter_url"
	"link-shortener-service/internal/infastracture/repository/inmemory"
	"link-shortener-service/internal/infastracture/repository/postgres"
	"link-shortener-service/internal/middleware"
	"link-shortener-service/internal/usecase/contract/repository"
	usecase_expander_url "link-shortener-service/internal/usecase/expander_url"
	usecase_shorter_url "link-shortener-service/internal/usecase/shorter_url"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

type App struct {
	server http.Server
	config config.Config
	pool   *pgxpool.Pool
}

func (a *App) Run() error {
	return a.server.ListenAndServe()
}

func (a *App) Stop() error {
	log.Println("Gracefully shutdown...")
	return a.server.Shutdown(context.Background())
}

func NewApp(ctx context.Context, cfg config.Config) (*App, error) {
	a := &App{
		config: cfg,
	}

	err := a.setup(ctx)
	if err != nil {
		return nil, err
	}

	return a, nil
}

func (a *App) setup(ctx context.Context) error {
	funcs := []func(context.Context) error{
		a.newPool,
		a.setupHttpServer,
		a.runMigrationsDB,
	}

	for _, f := range funcs {
		if err := f(ctx); err != nil {
			return err
		}
	}

	return nil
}

func (a *App) setupHttpServer(_ context.Context) error {
	var rep repository.URLRepository
	switch a.config.AppSettings.Storage {
	case "db":
		rep = postgres.NewDBRepository(a.pool)
	case "map":
		rep = inmemory.NewMapRepository()
	default:
		return errors.New(
			fmt.Sprintf("got unknown storage type from config: %s", a.config.AppSettings.Storage),
		)
	}
	valid := validator.New(validator.WithRequiredStructEnabled())

	shorterUseCase := usecase_shorter_url.NewUsecase(
		rep,
		a.config.AppSettings.FirstURLPart,
		a.config.AppSettings.URLLength,
	)
	shorter := shorter_url.NewUrlHandler(shorterUseCase, valid)

	expanderUseCase := usecase_expander_url.NewUsecase(rep)
	expander := expander_url.NewUrlHandler(expanderUseCase, valid)

	r := mux.NewRouter()
	r.HandleFunc("/", expander.ExpanderURL).Methods("GET")
	r.HandleFunc("/", shorter.ShorterURL).Methods("POST")

	h := middleware.LoggerMiddleware(r)
	h = middleware.PanicMiddleware(h)

	log.Println("starting on:" + a.config.Server.Address)
	a.server = http.Server{
		Addr:    a.config.Server.Address,
		Handler: h,
	}

	return nil
}

func (a *App) newPool(ctx context.Context) error {
	pool, err := pgxpool.New(ctx, a.config.DB.Conn)
	if err != nil {
		return err
	}
	a.pool = pool

	return nil
}

func (a *App) runMigrationsDB(_ context.Context) error {
	dsn := flag.String("dsn", a.config.DB.Conn, "PostgreSQL")

	sql, err := goose.OpenDBWithDriver("postgres", *dsn)
	if err != nil {
		return err
	}

	if err = goose.Up(sql, a.config.DB.MigrationsDir); err != nil {
		return err
	}

	return nil
}
