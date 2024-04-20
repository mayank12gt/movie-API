package main

import (
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/mayank12gt/movie-webapp/internal/data"
)

func (app *app) registerUserHandler() func(c echo.Context) error {
	return func(c echo.Context) error {

		var input struct {
			Name     string `json:"name"`
			Email    string `json:"email"`
			Password string `json:"password"`
		}

		if err := c.Bind(&input); err != nil {
			return c.JSON(http.StatusBadRequest, "Bad Request")
		}

		user := &data.User{
			Name:      input.Name,
			Email:     input.Email,
			Activated: false,
		}

		if err := user.Password.Set(input.Password); err != nil {
			return c.JSON(http.StatusBadRequest, err.Error())
		}

		v := validator.New()
		if err := v.Struct(user); err != nil {
			errors := err.(validator.ValidationErrors)
			app.logger.Print(errors)

			return c.JSON(422, errors.Error())
		}

		app.logger.Print(user)

		if err := app.models.Users.Insert(user); err != nil {
			return c.JSON(http.StatusInternalServerError, "Could not insert data")
		}

		err := app.models.Permissions.AddForUser(user.ID, "movies:read")
		if err != nil {
			return c.JSON(400, err.Error())
		}

		token, err := app.models.Tokens.New(user.ID, 3*24*time.Hour, data.ScopeActivation)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, err.Error())
		}

		app.background(func() {

			data := map[string]interface{}{
				"activationToken": token.Plaintext,
				"userId":          user.ID,
			}

			err := app.mailer.Send(user.Email, "user_welcome.tmpl", data)

			if err != nil {
				app.logger.Print(err)
			}
		})

		return c.JSON(200, user)

	}
}

func (app *app) activateUserHandler() func(c echo.Context) error {
	return func(c echo.Context) error {
		var input struct {
			TokenPlainText string `json:"token" validate:"required,min=26"`
		}

		if err := c.Bind(&input); err != nil {
			return c.JSON(http.StatusBadRequest, "Bad Request")
		}
		validate := validator.New()

		if err := validate.Struct(input); err != nil {
			errors := err.(validator.ValidationErrors)
			app.logger.Print(errors)

			return c.JSON(422, errors.Error())
		}

		user, err := app.models.Users.GetForToken(data.ScopeActivation, input.TokenPlainText)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, err.Error())
		}
		user.Activated = true

		err = app.models.Users.Update(user)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, err.Error())
		}

		err = app.models.Tokens.DeleteAllForUser(user.ID, data.ScopeActivation)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, err.Error())
		}

		return c.JSON(200, "Token Ok")
	}
}
