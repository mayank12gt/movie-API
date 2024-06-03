package main

import (
	"database/sql"
	"errors"
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
			return c.JSON(http.StatusBadRequest, map[string]string{
				"message": "Bad Request, verify JSON body",
			})
		}

		validate := validator.New()

		if err := validate.Struct(movie); err != nil {
			errors := err.(validator.ValidationErrors)
			app.logger.Print(errors)

			return c.JSON(422, map[string]string{
				"message": err.Error(),
			})
		}

		app.logger.Print(movie)

		if err := app.models.Movies.Insert(&movie); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"message": "Internal Server Error",
			})
		}

		return c.JSON(200, movie)
	}
}

func (app *app) getMovieHandler() func(c echo.Context) error {
	return func(c echo.Context) error {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"message": "id must be integer",
			})
		}

		movie, err := app.models.Movies.Get(int64(id))
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
		return c.JSON(200, movie)

	}
}

func (app *app) deleteMovieHandler() func(c echo.Context) error {
	return func(c echo.Context) error {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"message": "id must be integer",
			})
		}

		if err := app.models.Movies.Delete(int64(id)); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"message": "Internal Server Error",
			})
		}
		return c.JSON(200, map[string]string{
			"message": "Movie deleted",
		})

	}
}

func (app *app) updateMovieHandler() func(c echo.Context) error {
	return func(c echo.Context) error {

		type Req struct {
			Title   string   `json:"title" `
			Year    int32    `json:"year,omitempty" `
			Runtime int32    `json:"runtime,omitempty" `
			Genres  []string `json:"genres,omitempty"`
		}

		var request Req

		movieId, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"message": "id must be integer",
			})
		}

		err = c.Bind(&request)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"message": "Bad Request, verify the JSON body",
			})
		}

		movie, err := app.models.Movies.Get(int64(movieId))
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return c.JSON(404, map[string]string{
					"message": "no records found",
				})
			}
			return c.JSON(500, map[string]string{
				"message": "Internal Server Error",
			})
		}

		app.logger.Print(movie)

		movie.ID = int64(movieId)

		if request.Title != "" {
			movie.Title = request.Title
		}
		if request.Runtime != 0 {
			movie.Runtime = request.Runtime
		}
		if request.Year != 0 {
			movie.Year = request.Year
		}
		if len(request.Genres) != 0 {
			movie.Genres = request.Genres
		}
		app.logger.Print(movie)

		v := validator.New()

		if err := v.Struct(movie); err != nil {
			errors := err.(validator.ValidationErrors)
			app.logger.Print(errors)

			return c.JSON(422, map[string]string{
				"message": err.Error(),
			})
		}

		err = app.models.Movies.Update(movie)
		if err != nil {
			return c.JSON(500, map[string]string{
				"message": "Internal Server Error",
			})
		}

		return c.JSON(200, map[string]string{
			"message": "movie updated",
		})
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
				return c.JSON(http.StatusBadRequest, map[string]string{
					"message": "page must be integer",
				})
			}
		} else {
			input.Filters.Page = 1
		}

		if c.QueryParam("page_size") != "" {
			input.Filters.PageSize, err = strconv.Atoi(c.QueryParam("page_size"))
			if err != nil {
				return c.JSON(http.StatusBadRequest, map[string]string{
					"message": "page must be integer",
				})
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

			return c.JSON(422, map[string]string{
				"message": err.Error(),
			})
		}

		app.logger.Print(input)

		movies, meta, err := app.models.Movies.List(input.Title, input.Genres, input.Filters)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return c.JSON(404, map[string]string{
					"message": "records not found",
				})
			}
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

func (app *app) submitMovieRatingHandler() func(c echo.Context) error {
	return func(c echo.Context) error {

		movie_ID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"message": "id must be an integer",
			})
		}

		var input struct {
			Rating float64 `json:"rating" validate:"min=1"`
		}
		user := c.Get("user").(*data.User)

		if err := c.Bind(&input); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "Bad Request, verify the json body"})
		}

		rating := data.Rating{
			User_id:  user.ID,
			Movie_id: int64(movie_ID),
			Rating:   input.Rating,
		}

		if err := app.models.Ratings.AddRating(&rating); err != nil {
			return c.JSON(500, map[string]string{
				"message": "Internal Server Error",
			})
		}

		return c.JSON(200, rating)
	}
}

func (app *app) getMovieAverageRatingHandler() func(c echo.Context) error {
	return func(c echo.Context) error {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"message": "id must be integer",
			})
		}

		avearageRating, err := app.models.Movies.GetAverageRating(int64(id))
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return c.JSON(404, map[string]string{
					"message": "records not found",
				})
			}
			return c.JSON(500, map[string]string{
				"message": "Internal Server Error"})
		}

		return c.JSON(200, avearageRating)
	}
}

func (app *app) updateMovieRatingHandler() func(c echo.Context) error {
	return func(c echo.Context) error {
		movie_ID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"message": "id must be integer",
			})
		}

		var input struct {
			Rating float64 `json:"rating" validate:"min=1"`
		}
		user := c.Get("user").(*data.User)

		if err := c.Bind(&input); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "Bad Request, verify the json body"})
		}

		rating := data.Rating{
			User_id:  user.ID,
			Movie_id: int64(movie_ID),
			Rating:   input.Rating,
		}
		app.logger.Print(rating)
		if err := app.models.Ratings.UpdateRating(&rating); err != nil {
			return c.JSON(500, map[string]string{
				"message": "Internal Server Error",
			})
		}

		return c.JSON(200, rating)
	}
}

func (app *app) deleteMovieRatingHandler() func(c echo.Context) error {
	return func(c echo.Context) error {
		movie_ID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"message": "id must be integer",
			})
		}

		user := c.Get("user").(*data.User)

		rating := data.Rating{
			User_id:  user.ID,
			Movie_id: int64(movie_ID),
		}
		if err := app.models.Ratings.DeleteRating(&rating); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return c.JSON(404, map[string]string{
					"message": "records not found",
				})
			}
			return c.JSON(500, map[string]string{
				"message": "internal server error",
			})
		}

		return c.JSON(200, map[string]string{
			"message": "Rating Deleted",
		})
	}
}

func (app *app) getMovieRatingHandler() func(c echo.Context) error {
	return func(c echo.Context) error {
		movie_ID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"message": "id must be integer",
			})
		}

		user := c.Get("user").(*data.User)

		rating := data.Rating{
			User_id:  user.ID,
			Movie_id: int64(movie_ID),
		}

		if err := app.models.Ratings.GetRating(&rating); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return c.JSON(404, map[string]string{
					"message": "no records found",
				})
			}
			return c.JSON(500, err.Error())
		}

		return c.JSON(200, rating)
	}
}
