package handlers

import (
	"encoding/json"
	"errors"
	"filmlibrary/pkg/errs"
	"filmlibrary/pkg/session"
	"filmlibrary/pkg/users"
	"net/http"
	"text/template"

	"go.uber.org/zap"
)

type UsersHandler struct {
	Tmpl     *template.Template
	UserRepo users.UserRepo
	Logger   *zap.SugaredLogger
}

type UserData struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Token struct {
	Token string `json:"token"`
}

// @Summary Register
// @Description register new user
// @Tags users
// @Accept json
// @Produce json
// @Param  actor body UserData true "user data"
// @Success 201 {object} Response
// @Failed 400 {object} ErrorResponse
// @Failed 422 {object} ErrorResponse
// @Failed 500 {object} ErrorResponse
// @Router /api/register [post]
func (h *UsersHandler) Register(w http.ResponseWriter, r *http.Request) {
	var data UserData

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		newErr := errors.New(errs.JSONerror)
		writeError(h.Logger, w, http.StatusBadRequest, newErr)
		return
	}

	if len(data.Password) < 8 {
		err := errors.New(errs.ShortPass)
		writeError(h.Logger, w, http.StatusUnprocessableEntity, err)
		return
	}

	u, err := h.UserRepo.Signup(data.Username, data.Password)
	if err != nil {
		writeError(h.Logger, w, http.StatusUnprocessableEntity, err)
		return
	}
	h.Logger.Infof("created user %v", u.Login)

	err = session.CreateToken(w, u)
	if err != nil {
		writeError(h.Logger, w, http.StatusInternalServerError, err)
		return
	}

	writeResponse(h.Logger, w, http.StatusCreated, u.ID)
	h.Logger.Infof("created session for %v", u.ID)
}

// @Summary Login
// @Description  user login
// @Tags users
// @Accept json
// @Produce json
// @Param  actor body UserData true "user data"
// @Success 200 {object} Response
// @Failed 400 {object} ErrorResponse
// @Failed 401 {object} ErrorResponse
// @Failed 500 {object} ErrorResponse
// @Router /api/login [post]
func (h *UsersHandler) Login(w http.ResponseWriter, r *http.Request) {
	var data UserData

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		newErr := errors.New(errs.JSONerror)
		writeError(h.Logger, w, http.StatusBadRequest, newErr)
		return
	}

	u, err := h.UserRepo.Authorize(data.Username, data.Password)
	if err != nil {
		writeError(h.Logger, w, http.StatusUnauthorized, err)
		return
	}

	err = session.CreateToken(w, u)
	if err != nil {
		writeError(h.Logger, w, http.StatusInternalServerError, err)
		return
	}

	writeResponse(h.Logger, w, http.StatusOK, u.ID)
	h.Logger.Infof("created session for %v", u.ID)
}
