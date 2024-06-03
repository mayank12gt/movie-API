package main

import (
	"github.com/labstack/echo/v4"
	"github.com/mayank12gt/movie-webapp/internal/data"
)

func (app *app) authenticate(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {

		app.logger.Print("auth middleware")
		authToken, err := c.Cookie("token")

		if err != nil {
			return c.JSON(400, map[string]string{
				"error": "user is not authenticated",
			})
		}

		user, err := app.models.Users.GetForToken(data.ScopeAuthentication, authToken.Value)
		if err != nil {
			return c.JSON(500, map[string]string{
				"message": "Internal server error",
			})
		}

		c.Set("user", user)

		return next(c)

	}
}

func (app *app) checkPermission(code string, next echo.HandlerFunc) echo.HandlerFunc {
	fn := func(c echo.Context) error {
		app.logger.Print("PermissionsMiddleware")
		user := c.Get("user").(*data.User)

		permissions, err := app.models.Permissions.GetAllForUser(user.ID)
		if err != nil {
			return c.JSON(500, map[string]string{
				"message": "internal server error",
			})
		}
		if !permissions.Include(code) {
			return c.JSON(400, map[string]string{
				"message": "user does not have the required permission",
			})
		}

		return next(c)

	}
	return app.authenticate(fn)
}

// func (app *app) requireActivated(next echo.HandlerFunc) echo.HandlerFunc {
// 	return func(c echo.Context) error {

// 		app.logger.Print("activated middleware")
// 		user := c.Get("user").(data.User)

// 		if !user.Activated{
// 			return c.JSON(400,"Please activate your account")
// 		}

// 		return next(c)

// 	}
//}
