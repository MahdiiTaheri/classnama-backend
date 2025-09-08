package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/MahdiiTaheri/classnama-backend/docs"
	"github.com/MahdiiTaheri/classnama-backend/internal/auth"
	"github.com/MahdiiTaheri/classnama-backend/internal/ratelimiter"
	"github.com/MahdiiTaheri/classnama-backend/internal/store"
	"github.com/MahdiiTaheri/classnama-backend/internal/store/cache"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger/v2"
	"go.uber.org/zap"
)

type application struct {
	config        config
	logger        *zap.SugaredLogger
	store         store.Storage
	cacheStorage  cache.Storage
	authenticator auth.Authenticator
	ratelimiter   ratelimiter.Limiter
}

type config struct {
	addr        string
	env         string
	apiURL      string
	db          dbConfig
	auth        authConfig
	redisCfg    redisCfg
	ratelimiter ratelimiter.Config
}

type redisCfg struct {
	addr    string
	pw      string
	db      int
	enabled bool
}

type dbConfig struct {
	addr         string
	maxOpenConns int
	maxIdleConns int
	maxIdleTime  string
}

type authConfig struct {
	basic basicConfig
	token tokenConfig
}

type tokenConfig struct {
	secret string
	exp    time.Duration
	iss    string
}

type basicConfig struct {
	user string
	pass string
}

func (app *application) mount() http.Handler {
	r := chi.NewRouter()

	// middlewares
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(app.RateLimiterMiddleware)

	r.Route("/v1", func(r chi.Router) {
		r.Get("/health", app.healthCheckHandler)

		docsURL := fmt.Sprintf("%s/swagger/doc.json", app.config.addr)
		r.Get("/swagger/*", httpSwagger.Handler(httpSwagger.URL(docsURL)))

		r.Route("/execs", func(r chi.Router) {
			// PUBLIC
			r.Post("/register", app.registerExecHandler)
			r.Post("/login", app.loginExecHandler)

			// PROTECTED
			r.Group(func(r chi.Router) {
				r.Use(app.AuthTokenMiddleware)
				r.Use(app.requireRole("admin", "manager")) // only execs can access
				r.Get("/", app.getExecsHandler)

				r.Route("/{execID}", func(r chi.Router) {
					r.Use(app.execsContextMiddleware) // ONLY for routes with execID
					r.Get("/", app.getExecHandler)
					r.Patch("/", app.updateExecHandler)
					r.Delete("/", app.deleteExecHandler)
				})
			})
		})

		r.Route("/teachers", func(r chi.Router) {
			// PUBLIC LOGIN
			r.Post("/login", app.loginTeacherHandler)

			// PROTECTED: Only execs can manage teachers
			r.Group(func(r chi.Router) {
				r.Use(app.AuthTokenMiddleware)
				r.Use(app.requireRole("manager", "admin")) // only execs can access
				r.Post("/", app.registerTeacherHandler)
				r.Get("/", app.getTeachersHandler)

				r.Route("/{teacherID}", func(r chi.Router) {
					r.Use(app.teachersContextMiddleware)
					r.Get("/", app.getTeacherHandler)
					r.Get("/students", app.getStudentsByTeacherHandler)
					r.Patch("/", app.updateTeacherHandler)
					r.Delete("/", app.deleteTeacherHandler)
				})
			})
		})

		r.Route("/students", func(r chi.Router) {
			// PUBLIC LOGIN
			r.Post("/login", app.loginStudentHandler)

			// PROTECTED: Only execs can manage students
			r.Group(func(r chi.Router) {
				r.Use(app.AuthTokenMiddleware)
				r.Use(app.requireRole("admin", "manager")) // only execs can access
				r.Post("/", app.registerStudentHandler)
				r.Get("/", app.getStudentsHandler)

				r.Route("/{studentID}", func(r chi.Router) {
					r.Use(app.studentsContextMiddleware)
					r.Get("/", app.getStudentHandler)
					r.Patch("/", app.updateStudentHandler)
					r.Delete("/", app.deleteStudentHandler)
				})
			})
		})

	})

	return r
}

func (app *application) run(mux http.Handler) error {

	//Docs
	docs.SwaggerInfo.Version = version
	docs.SwaggerInfo.Host = app.config.apiURL
	docs.SwaggerInfo.BasePath = "/v1"

	srv := &http.Server{
		Addr:         app.config.addr,
		Handler:      mux,
		WriteTimeout: 30 * time.Second,
		ReadTimeout:  10 * time.Second,
		IdleTimeout:  time.Minute,
	}

	shutdown := make(chan error)

	// graceful shutdown
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		s := <-quit

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		app.logger.Infow("signal caught", "signal", s.String())
		shutdown <- srv.Shutdown(ctx)
	}()

	app.logger.Infow("server started", "addr", app.config.addr, "env", app.config.env)

	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	err = <-shutdown
	if err != nil {
		return err
	}

	app.logger.Infow("server stopped", "addr", app.config.addr, "env", app.config.env)
	return nil
}
