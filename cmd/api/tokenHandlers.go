package main

import (
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
			return c.JSON(http.StatusBadRequest, "Bad Request")
		}

		v := validator.New()
		if err := v.Struct(input); err != nil {
			errors := err.(validator.ValidationErrors)
			app.logger.Print(errors)

			return c.JSON(422, errors.Error())
		}

		user, err := app.models.Users.GetByEmail(input.Email)
		if err != nil {
			return c.JSON(400, err.Error())
		}
		if !user.Activated {
			return c.JSON(400, "Acoount not activated")
		}

		match, err := user.Password.Compare(input.Password)
		if err != nil {
			return c.JSON(400, err.Error())
		}
		if !match {
			return c.JSON(400, "Incorrect Password")
		}

		token, err := app.models.Tokens.New(user.ID, 24*time.Hour, data.ScopeAuthentication)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, err.Error())
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
