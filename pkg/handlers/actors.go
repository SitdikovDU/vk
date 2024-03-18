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

type ActorsHandler struct {
	Tmpl       *template.Template
	ActorsRepo items.ItemRepo
	Logger     *zap.SugaredLogger
}

// @Summary Create actor
// @Description Create new actor
// @Security ApiKeyAuth
// @Tags actors
// @Accept json
// @Produce json
// @Param  actor body actors.Actor true "actor data"
// @Success 201 {object} Response
// @Failed 400 {object} ErrorResponse
// @Failed 500 {object} ErrorResponse
// @Router /api/actors [post]
func (h *ActorsHandler) CreateActor(w http.ResponseWriter, r *http.Request) {
	var actor items.Actor

	err := json.NewDecoder(r.Body).Decode(&actor)
	if err != nil {
		newErr := errors.New(errs.JSONerror)
		writeError(h.Logger, w, http.StatusBadRequest, newErr)
		return
	}

	actor.ID, err = h.ActorsRepo.CreateActor(actor)
	if err != nil {
		writeError(h.Logger, w, http.StatusInternalServerError, err)
		return
	}

	writeResponse(h.Logger, w, http.StatusCreated, actor)
}

// @Summary Get actors
// @Description Get actor list
// @Tags actors
// @Produce json
// @Success 200 {object} Response
// @Failed 500 {object} ErrorResponse
// @Router /api/actors [get]
func (h *ActorsHandler) GetActors(w http.ResponseWriter, r *http.Request) {

	actors, err := h.ActorsRepo.GetActors()
	if err != nil {
		writeError(h.Logger, w, http.StatusInternalServerError, err)
		return
	}

	writeResponse(h.Logger, w, http.StatusOK, actors)
}

// @Summary Get actor
// @Description Get actor by id
// @Tags actors
// @Produce json
// @Param  id path int true "actor id"
// @Success 200 {object} Response
// @Failed 400 {object} ErrorResponse
// @Failed 500 {object} ErrorResponse
// @Router /api/actors/{id} [get]
func (h *ActorsHandler) GetActor(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	id, err := strconv.ParseUint(vars["ACTOR_ID"], 10, 32)
	if err != nil {
		writeError(h.Logger, w, http.StatusBadRequest, err)
		return
	}

	actor, err := h.ActorsRepo.GetActorByID(uint32(id))
	if err != nil {
		writeError(h.Logger, w, http.StatusInternalServerError, err)
		return
	}

	writeResponse(h.Logger, w, http.StatusOK, actor)
}

// @Summary Update actor
// @Description Update full actor data
// @Security ApiKeyAuth
// @Tags actors
// @Accept json
// @Produce json
// @Param  id path int true "actor id"
// @Param  actor body actors.Actor true "actor data"
// @Success 200 {object} Response
// @Failed 400 {object} ErrorResponse
// @Failed 500 {object} ErrorResponse
// @Router /api/actors/{id} [post]
func (h *ActorsHandler) UpdateActor(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	id, err := strconv.ParseUint(vars["ACTOR_ID"], 10, 32)
	if err != nil {
		writeError(h.Logger, w, http.StatusBadRequest, err)
		return
	}

	var actor items.Actor
	err = json.NewDecoder(r.Body).Decode(&actor)
	if err != nil {
		newErr := errors.New(errs.JSONerror)
		writeError(h.Logger, w, http.StatusBadRequest, newErr)
		return
	}
	actor.ID = uint32(id)

	err = h.ActorsRepo.UpdateActor(actor)
	if err != nil {
		writeError(h.Logger, w, http.StatusInternalServerError, err)
		return
	}

	writeResponse(h.Logger, w, http.StatusOK, actor)
}

// @Summary Update actor column
// @Description Update one column actor data
// @Security ApiKeyAuth
// @Tags actors
// @Accept json
// @Produce json
// @Param  id path int true "actor id"
// @Param  column path string true "column name"
// @Param  actor body actors.Actor true "actor data"
// @Success 200 {object} Response
// @Failed 400 {object} ErrorResponse
// @Failed 500 {object} ErrorResponse
// @Router /api/actors/{id}/{columnName} [post]
func (h *ActorsHandler) UpdateColumnActor(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	columnName := vars["COLUMN_NAME"]

	id, err := strconv.ParseUint(vars["ACTOR_ID"], 10, 32)
	if err != nil {
		writeError(h.Logger, w, http.StatusBadRequest, err)
		return
	}

	var actor items.Actor
	t := reflect.TypeOf(actor)

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

	err = json.NewDecoder(r.Body).Decode(&actor)
	if err != nil {
		newErr := errors.New(errs.JSONerror)
		writeError(h.Logger, w, http.StatusBadRequest, newErr)
		return
	}
	actor.ID = uint32(id)

	err = h.ActorsRepo.UpdateColumnActor(actor, columnName)
	if err != nil {
		writeError(h.Logger, w, http.StatusInternalServerError, err)
		return
	}

	writeResponse(h.Logger, w, http.StatusOK, actor)
}
