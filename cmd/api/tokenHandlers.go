package main

import (
	"database/sql"
	"errors"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/mayank12gt/movie-webapp/internal/data"
)

func (app *app) createAuthenticationTokenHandler() func(c echo.Context) error {
	return func(c echo.Context) error {

		var input struct {
			Email    string `json:"email" validate:"required,email"`
			Password string `json:"password" validate:"required,min=8,max=72"`
		}

		if err := c.Bind(&input); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"message": "Bad Request, verify json body",
			})
		}

		v := validator.New()
		if err := v.Struct(input); err != nil {
			errors := err.(validator.ValidationErrors)
			app.logger.Print(errors)

			return c.JSON(422, map[string]string{
				"message": err.Error(),
			})
		}

		user, err := app.models.Users.GetByEmail(input.Email)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return c.JSON(404, map[string]string{
					"message": "records not found",
				})
			}
			return c.JSON(500, map[string]string{
				"message": "Internal Server Error",
			})
		}
		if !user.Activated {
			return c.JSON(422, map[string]string{
				"message": "User not activated",
			})
		}

		match, err := user.Password.Compare(input.Password)
		if err != nil {
			return c.JSON(500, map[string]string{
				"message": "Internal Server Error",
			})
		}
		if !match {
			return c.JSON(422, map[string]string{
				"message": "password is incorrect",
			})
		}

		token, err := app.models.Tokens.New(user.ID, 24*time.Hour, data.ScopeAuthentication)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"message": "Internal Server Error",
			})
		}

		cookie := http.Cookie{
			Name:     "token",
			Value:    token.Plaintext,
			Secure:   false,
			Expires:  time.Now().Add(time.Hour * 24),
			HttpOnly: true,
			Path:     "/",
		} //Creates the cookie to be passed.

		c.SetCookie(&cookie)
		return c.JSON(200, token)
	}
}

func (app *app) signOutHandler() func(c echo.Context) error {
	return func(c echo.Context) error {

		user := c.Get("user").(*data.User)

		err := app.models.Tokens.DeleteAllForUser(user.ID, data.ScopeAuthentication)
		if err != nil {
			return c.JSON(500, map[string]string{
				"message": "Internal Server Error",
			})
		}

		cookie := http.Cookie{
			Name:     "token",
			Value:    "",
			Secure:   false,
			Expires:  time.Now().Add(-time.Hour),
			HttpOnly: true,
			Path:     "/",
		}

		c.SetCookie(&cookie)
		return c.JSON(200, map[string]string{
			"message": "User Signed Out",
		})
	}
}
