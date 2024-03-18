package handlers

import (
	"encoding/json"
	"errors"
	"filmlibrary/pkg/errs"
	"net/http"
	"strconv"

	"go.uber.org/zap"
)

const (
	orderDesc = -1
	orderAsc  = 1

	fieldName   = "name"
	fieldRating = "rating"
	fieldDate   = "date"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

type Response struct {
	Data interface{} `json:"data"`
}

func writeResponse(logger *zap.SugaredLogger, w http.ResponseWriter, httpStatus int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(httpStatus)

	err := json.NewEncoder(w).Encode(Response{Data: data})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	logger.Info(data)
}

func writeError(logger *zap.SugaredLogger, w http.ResponseWriter, httpStatus int, myErr error) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(httpStatus)

	err := json.NewEncoder(w).Encode(ErrorResponse{Error: myErr.Error()})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logger.Error(err)
		return
	}

	logger.Error(myErr)
}

func parseOrderBy(r *http.Request) (string, int, error) {
	strField := r.URL.Query().Get("field")
	strOrder := r.URL.Query().Get("order")

	order := orderDesc
	if strOrder != "" {
		newOrder, err := strconv.Atoi(strOrder)
		if err != nil {
			return "", 0, errors.New(errs.ReadingOrderError)
		}

		if newOrder != orderAsc && newOrder != orderDesc {
			return "", 0, errors.New(errs.ReadingOrderByError)
		}

		order = newOrder
	}

	field := fieldRating
	if strField != "" {
		if strField != fieldName && strField != fieldRating && strField != fieldDate {
			return "", 0, errors.New(errs.ReadingOrderByError)
		}
	}

	return field, order, nil
}
