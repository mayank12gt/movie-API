package main

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/mayank12gt/movie-webapp/internal/data"
)

type Response struct {
	// Metadata map[string]interface{} `json:"metadata"`
	MetaData data.Metadata `json:"metadata"`
	Movies   []*data.Movie `json:"movies"`
}

func (app *app) createMovieHandler() func(c echo.Context) error {
	return func(c echo.Context) error {

		var movie data.Movie

		if err := c.Bind(&movie); err != nil {
			return c.JSON(http.StatusBadRequest, "Bad Request")
		}

		validate := validator.New()

		if err := validate.Struct(movie); err != nil {
			errors := err.(validator.ValidationErrors)
			app.logger.Print(errors)

			return c.JSON(422, errors.Error())
		}

		app.logger.Print(movie)

		if err := app.models.Movies.Insert(&movie); err != nil {
			return c.JSON(http.StatusInternalServerError, "Could not insert data")
		}

		return c.JSON(200, movie)
	}
}

func (app *app) getMovieHandler() func(c echo.Context) error {
	return func(c echo.Context) error {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return c.JSON(http.StatusBadRequest, err.Error())
		}

		movie, err := app.models.Movies.Get(int64(id))
		if err != nil {
			return c.JSON(http.StatusNotFound, err.Error())
		}
		return c.JSON(200, movie)

	}
}

func (app *app) deleteMovieHandler() func(c echo.Context) error {
	return func(c echo.Context) error {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return c.JSON(http.StatusBadRequest, err.Error())
		}

		if err := app.models.Movies.Delete(int64(id)); err != nil {
			return c.JSON(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(200, "Movie Deleted")

	}
}

func (app *app) updateMovieHandler() func(c echo.Context) error {
	return func(c echo.Context) error {
		return nil
	}
}

func (app *app) listMovieHandler() func(c echo.Context) error {
	return func(c echo.Context) error {

		var input struct {
			Title  string
			Genres []string
			data.Filters
		}

		input.Title = c.QueryParam("title")
		if c.QueryParam("genres") != "" {
			input.Genres = strings.Split(c.QueryParam("genres"), ",")
		} else {
			input.Genres = []string{}
		}

		var err error
		if c.QueryParam("page") != "" {
			input.Filters.Page, err = strconv.Atoi(c.QueryParam("page"))
			if err != nil {
				return c.JSON(http.StatusBadRequest, err.Error())
			}
		} else {
			input.Filters.Page = 1
		}

		if c.QueryParam("page_size") != "" {
			input.Filters.PageSize, err = strconv.Atoi(c.QueryParam("page_size"))
			if err != nil {
				return c.JSON(http.StatusBadRequest, err.Error())
			}
		} else {
			input.Filters.PageSize = 20
		}

		input.Filters.Sort = c.QueryParam("sort")
		if input.Filters.Sort == "" {
			input.Filters.Sort = "id"
		}
		// if input.Filters.Page == 0 {
		// 	input.Filters.Page = 1
		// }
		// if input.Genres == nil {
		// 	input.Genres = []string{}
		// }
		if input.Title == "" {
			input.Title = ""
		}
		// if input.Filters.PageSize == 0 {
		// 	input.Filters.PageSize = 20
		// }

		validate := validator.New()

		if err := validate.Struct(input); err != nil {
			errors := err.(validator.ValidationErrors)
			app.logger.Print(errors)

			return c.JSON(422, errors.Error())
		}

		app.logger.Print(input)

		movies, meta, err := app.models.Movies.List(input.Title, input.Genres, input.Filters)
		if err != nil {
			return c.JSON(http.StatusNotFound, err.Error())
		}

		res := Response{
			MetaData: meta,
			Movies:   movies,
		}
		return c.JSON(200, res)

		//return c.JSON(200, input)

	}
}
