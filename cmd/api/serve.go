package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func (app *app) serve() error {

	server := echo.New()
	server.Use(middleware.CORS())
	app.registerHandlers(server)

	shutdownErr := make(chan error)

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

		s := <-quit

		app.logger.Print(s.String())
		app.logger.Print("Shutting Down")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			shutdownErr <- err
		}
		app.logger.Print("Completing Background Tasks")

		app.wg.Wait()
		shutdownErr <- nil

	}()

	app.logger.Print(app.config.port)
	err := server.Start(":3000")

	if !errors.Is(err, http.ErrServerClosed) {
		app.logger.Print("here")
		app.logger.Print(err.Error())
		return err
	}

	err = <-shutdownErr
	if err != nil {
		return err
	}
	app.logger.Printf("server stopped")

	return nil
}

func (app *app) registerHandlers(server *echo.Echo) {
	server.POST("/movies", app.checkPermission("movies:write", app.createMovieHandler()))
	server.GET("/movies", app.listMovieHandler())
	server.GET("/movies/:id", app.getMovieHandler())
	server.DELETE("/movies/:id", app.checkPermission("movies:write", app.deleteMovieHandler()))
	server.PUT("/movies/:id", app.checkPermission("movies:write", app.updateMovieHandler()))

	server.POST("/movies/:id/ratings", app.submitMovieRatingHandler(), app.authenticate)
	server.GET("/movies/:id/ratings", app.getMovieAverageRatingHandler())
	server.GET("/movies/:id/rating", app.getMovieRatingHandler(), app.authenticate)
	server.PUT("/movies/:id/ratings", app.updateMovieRatingHandler(), app.authenticate)
	server.DELETE("/movies/:id/ratings", app.deleteMovieRatingHandler(), app.authenticate)

	server.POST("/users", app.registerUserHandler())
	server.POST("/users/activate", app.activateUserHandler())
	server.POST("/users/authenticate", app.createAuthenticationTokenHandler())
	server.POST("/users/signOut", app.signOutHandler(), app.authenticate)

	server.GET("/", func(c echo.Context) error {
		//time.Sleep(4 * time.Second)
		return c.JSON(200, map[string]string{
			"message": "Up and Running",
		})
	})

}
