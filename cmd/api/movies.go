package main

import (
	"errors"
	"fmt"
	"net/http"
	"secondBook/internal/data"
	"secondBook/internal/validator"
)

func (app *application) createMovieHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title   string       `json:"title"`
		Year    int32        `json:"year"`
		Runtime data.RunTime `json:"runtime"`
		Genres  []string     `json:"genres"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.appBadResp(w, r, err)
		//app.errResp(w, r, http.StatusBadRequest, err.Error())
		return
	}

	movie := &data.Movie{
		Title:   input.Title,
		Year:    input.Year,
		Runtime: input.Runtime,
		Genres:  input.Genres,
	}

	v := validator.New()
	if data.ValidateMovie(v, movie); !v.Valid() {
		app.failValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Movies.Insert(movie)
	if err != nil {
		app.serverErrResp(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprint("/v1/movies/%d", movie.Id))

	err = app.writeJSON(w, http.StatusCreated, envelop{"movie": movie}, headers)
	if err != nil {
		app.serverErrResp(w, r, err)
	}

	//	fmt.Fprintf(w, "%+v\n", input)

}
func (app *application) showMovieHandler(w http.ResponseWriter, r *http.Request) {

	id, err := app.retriveId(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	movie, err := app.models.Movies.Read(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrResp(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelop{"movie": movie}, nil)
	if err != nil {

		app.serverErrResp(w, r, err)
	}

}
func (app *application) updateHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.retriveId(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	movie, err := app.models.Movies.Read(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrResp(w, r, err)

		}
		return
	}

	var input struct {
		Title   *string       `json:"title"`
		Year    *int32        `json:"year"`
		Runtime *data.RunTime `json:"runtime"`
		Genres  []string      `json:"genres"`
	}
	err = app.readJSON(w, r, &input)
	if err != nil {
		app.appBadResp(w, r, err)
		return
	}
	if input.Title != nil {
		movie.Title = *input.Title
	}
	if input.Year != nil {
		movie.Year = *input.Year
	}
	if input.Runtime != nil {
		movie.Runtime = *input.Runtime
	}
	if input.Genres != nil {
		movie.Genres = input.Genres
	}

	v := validator.New()
	if data.ValidateMovie(v, movie); !v.Valid() {
		app.failValidationResponse(w, r, v.Errors)
		return
	}
	err = app.models.Movies.Update(movie)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrResp(w, r, err)
		}
		return
	}
	err = app.writeJSON(w, http.StatusOK, envelop{"movie": movie}, nil)
	if err != nil {
		app.serverErrResp(w, r, err)
	}

}
func (app *application) deleteHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.retriveId(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	err = app.models.Movies.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrResp(w, r, err)
		}
		return
	}
	err = app.writeJSON(w, http.StatusOK, envelop{"message": "movie successfully deleted"}, nil)
	if err != nil {
		app.serverErrResp(w, r, err)
	}

}
func (app *application) listMovies(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title  string
		Genres []string
		data.Filters
	}

	v := validator.New()
	qs := r.URL.Query()
	input.Title = app.readString(qs, "title", "")
	input.Genres = app.readCsv(qs, "genres", []string{})
	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 20, v)
	input.Filters.Sort = app.readString(qs, "sort", "id")
	input.Filters.SortSafe = []string{"id", "title", "year", "runtime", "-id", "-title", "-year", "-runtime"}
	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failValidationResponse(w, r, v.Errors)
		return
	}
	movies, metadata, err := app.models.Movies.GetAll(input.Title, input.Genres, input.Filters)
	if err != nil {
		app.serverErrResp(w, r, err)
		return
	}
	err = app.writeJSON(w, http.StatusOK, envelop{"movies": movies, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrResp(w, r, err)
	}

	//fmt.Fprintf(w, "%+v\n", input)

}