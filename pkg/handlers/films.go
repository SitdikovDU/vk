package handlers

import (
	"encoding/json"
	"errors"
	"filmlibrary/pkg/errs"
	"filmlibrary/pkg/items"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"text/template"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type FilmsHandler struct {
	Tmpl      *template.Template
	FilmsRepo items.ItemRepo
	Logger    *zap.SugaredLogger
}

// @Summary Create film
// @Description Create new film
// @Security ApiKeyAuth
// @Tags films
// @Accept json
// @Produce json
// @Param  actor body films.Film true "film data"
// @Success 201 {object} Response
// @Failed 400 {object} ErrorResponse
// @Failed 500 {object} ErrorResponse
// @Router /api/films [post]
func (h *FilmsHandler) CreateFilm(w http.ResponseWriter, r *http.Request) {
	var film items.Film

	err := json.NewDecoder(r.Body).Decode(&film)
	if err != nil {
		newErr := errors.New(errs.JSONerror)
		writeError(h.Logger, w, http.StatusBadRequest, newErr)
		return
	}

	film.ID, err = h.FilmsRepo.CreateFilm(film)
	if err != nil {
		writeError(h.Logger, w, http.StatusInternalServerError, err)
		return
	}

	writeResponse(h.Logger, w, http.StatusCreated, film.ID)
}

// @Summary Get films
// @Description get films sorted by parameters
// @Tags films
// @Produce json
// @Param field query string false "sorting field"
// @Param order query int false "desc or asc"
// @Success 200 {object} Response
// @Failed 400 {object} ErrorResponse
// @Failed 500 {object} ErrorResponse
// @Router /api/films [get]
func (h *FilmsHandler) GetFilms(w http.ResponseWriter, r *http.Request) {
	field, order, err := parseOrderBy(r)
	if err != nil {
		writeError(h.Logger, w, http.StatusBadRequest, err)
		return
	}

	films, err := h.FilmsRepo.GetFilms(field, order)
	if err != nil {
		writeError(h.Logger, w, http.StatusInternalServerError, err)
		return
	}

	writeResponse(h.Logger, w, http.StatusOK, films)
}

// @Summary Search film
// @Description get films
// @Tags films
// @Produce json
// @Param query query string true "search query"
// @Success 200 {object} Response
// @Failed 400 {object} ErrorResponse
// @Failed 500 {object} ErrorResponse
// @Router /api/films/search [get]
func (h *FilmsHandler) SearchFilm(w http.ResponseWriter, r *http.Request) {
	searchQuery := r.URL.Query().Get("query")
	if searchQuery == "" {
		myErr := errors.New(errs.EmptySearchError)
		writeError(h.Logger, w, http.StatusBadRequest, myErr)
		return
	}

	films, err := h.FilmsRepo.SearchFilm(searchQuery)
	if err != nil {
		writeError(h.Logger, w, http.StatusInternalServerError, err)
		return
	}

	writeResponse(h.Logger, w, http.StatusOK, films)
}

// @Summary Update film
// @Description update full information about film
// @Security ApiKeyAuth
// @Tags films
// @Accept json
// @Produce json
// @Param id path int true "film id"
// @Param  actor body films.Film true "film data"
// @Success 200 {object} Response
// @Failed 400 {object} ErrorResponse
// @Failed 500 {object} ErrorResponse
// @Router /api/films/{id} [post]
func (h *FilmsHandler) UpdateFilm(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	id, err := strconv.ParseUint(vars["FILM_ID"], 10, 32)
	if err != nil {
		writeError(h.Logger, w, http.StatusBadRequest, err)
		return
	}

	var film items.Film
	err = json.NewDecoder(r.Body).Decode(&film)
	if err != nil {
		newErr := errors.New(errs.JSONerror)
		writeError(h.Logger, w, http.StatusBadRequest, newErr)
		return
	}
	film.ID = uint32(id)

	err = h.FilmsRepo.UpdateFilm(film)
	if err != nil {
		writeError(h.Logger, w, http.StatusInternalServerError, err)
		return
	}

	writeResponse(h.Logger, w, http.StatusOK, film.ID)
}

// @Summary Update film column
// @Description Can delete, change or add information
// @Security ApiKeyAuth
// @Tags films
// @Accept json
// @Produce json
// @Param id path int true "film id"
// @Param column path string true "change column"
// @Param  actor body films.Film true "film data"
// @Success 200 {object} Response
// @Failed 400 {object} ErrorResponse
// @Failed 500 {object} ErrorResponse
// @Router /api/films/{id}/{column} [post]
func (h *FilmsHandler) UpdateColumnFilm(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	columnName := vars["COLUMN_NAME"]

	id, err := strconv.ParseUint(vars["FILM_ID"], 10, 32)
	if err != nil {
		writeError(h.Logger, w, http.StatusBadRequest, err)
		return
	}

	var film items.Film
	t := reflect.TypeOf(film)

	columnExist := false
	for i := 0; i < t.NumField(); i++ {
		if strings.EqualFold(t.Field(i).Name, columnName) {
			columnExist = true
			columnName = t.Field(i).Name
		}
	}

	if !columnExist {
		newErr := errors.New(errs.WrongColumnError)
		writeError(h.Logger, w, http.StatusBadRequest, newErr)
		return
	}

	err = json.NewDecoder(r.Body).Decode(&film)
	if err != nil {
		writeError(h.Logger, w, http.StatusBadRequest, err)
		return
	}
	film.ID = uint32(id)

	if columnName == "Actors" {
		err := h.FilmsRepo.DeleteActors(film.ID)
		if err != nil {
			writeError(h.Logger, w, http.StatusInternalServerError, err)
			return
		}
		err = h.FilmsRepo.InsertActors(film.ID, film.Actors)
		if err != nil {
			writeError(h.Logger, w, http.StatusInternalServerError, err)
			return
		}
	} else {
		err = h.FilmsRepo.UpdateColumnFilm(film, columnName)
		if err != nil {
			writeError(h.Logger, w, http.StatusInternalServerError, err)
			return
		}
	}

	writeResponse(h.Logger, w, http.StatusOK, film.ID)
}
