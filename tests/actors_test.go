package tests

import (
	"database/sql"
	"filmlibrary/pkg/handlers"
	"filmlibrary/pkg/items"
	"filmlibrary/pkg/users"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

func PrepareActors(db *sql.DB) {
	qs := []string{
		`DROP TABLE IF EXISTS users cascade;`,
		`DROP TABLE IF EXISTS films cascade;`,
		`DROP TABLE IF EXISTS film_actor cascade;`,
		`CREATE TABLE users (
			id SERIAL PRIMARY KEY,
			username VARCHAR(50) NOT NULL,
			role VARCHAR(50) NOT NULL DEFAULT 'user',
			hashed_password VARCHAR(255) NOT NULL
		);`,

		`INSERT INTO users (username, role, hashed_password) VALUES
		('admin', 'admin','$2a$10$GJJT0w7LaC4oOfTygVVi8Ofn8D5kZHTXWdhA8PzscqkK846A2oG.C');`,
		`DROP TABLE IF EXISTS actors cascade;`,
		`CREATE TABLE actors (
			id SERIAL PRIMARY KEY,
			name VARCHAR(100) DEFAULT NULL,
			gender VARCHAR(10) DEFAULT NULL,
			date DATE DEFAULT NULL
		);`,
		`CREATE TABLE films (
			id SERIAL PRIMARY KEY,
			name VARCHAR(1000) NOT NULL,
			description VARCHAR(1000),
			date INTEGER CHECK (date >= 1900 AND date <= 2200),
			rating INT CHECK (rating >= 0 AND rating <= 10) DEFAULT NULL
		);`,
		`CREATE TABLE film_actor (
			film_id INTEGER,
			actor_id INTEGER,
			PRIMARY KEY (film_id, actor_id)
		);`,
	}

	for _, q := range qs {
		_, err := db.Exec(q)
		if err != nil {
			panic(err)
		}
	}
}

func TestActors(t *testing.T) {
	zapLogger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("failed to create logger: %v", err)
		return
	}

	defer zapLogger.Sync()

	logger := zapLogger.Sugar()

	DSN := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable",
		user, password, host, dbname)

	db, err := sql.Open("postgres", DSN)
	if err != nil {
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	PrepareActors(db)

	handler, err := fakeExplorer(db, logger) //nolint:typecheck
	if err != nil {
		panic(err)
	}

	ts := httptest.NewServer(handler)

	cases := []Case{Case{
		Path:   "/api/login", // список таблиц
		Method: http.MethodPost,
		Body: CR{
			"username": "admin",
			"password": "MySuperSecretPassword",
		},
		Status: http.StatusOK,
		Result: CR{"data": 1},
	},
		Case{
			Path:   "/api/actors", // список таблиц
			Method: http.MethodPost,
			Body: CR{
				"name":   "Mila",
				"gender": "female",
				"date":   "",
			},
			Status: http.StatusCreated,
			Result: CR{"data": CR{"id": 1, "name": "Mila", "gender": "female", "date": "", "films": nil}},
		},
		Case{
			Path:   "/api/actors", // список таблиц
			Method: http.MethodPost,
			Body: CR{
				"name":   CR{},
				"gender": "female",
				"date":   "",
			},
			Status: http.StatusBadRequest,
			Result: CR{"error": "decode JSON error"},
		},
		Case{
			Path:   "/api/actors",
			Method: http.MethodPost,
			Body: CR{
				"name":   "",
				"gender": "",
				"date":   "",
			},
			Status: http.StatusInternalServerError,
			Result: CR{"error": "empty actor"},
		},
		Case{
			Path:   "/api/actors/0",
			Method: http.MethodPost,
			Body: CR{
				"name":   "",
				"gender": "",
				"date":   "",
			},
			Status: http.StatusInternalServerError,
			Result: CR{"error": "sql: no rows in result set"},
		},
		Case{
			Path:   "/api/actors",
			Method: http.MethodPost,
			Body: CR{
				"name":   "Леонардо Ди Каприо",
				"gender": "male",
				"date":   "",
			},
			Status: http.StatusCreated,
			Result: CR{"data": CR{"id": 2, "name": "Леонардо Ди Каприо", "gender": "male", "date": "", "films": nil}},
		},
		Case{
			Path:   "/api/actors/2",
			Method: http.MethodPost,
			Body: CR{
				"name":   "Леонардо Ди Каприо",
				"gender": "male",
				"date":   "11-11-1974",
			},
			Status: http.StatusOK,
			Result: CR{"data": CR{"id": 2, "name": "Леонардо Ди Каприо", "gender": "male", "date": "11-11-1974", "films": nil}},
		},
		Case{
			Path:   "/api/actors/2", // список таблиц
			Method: http.MethodPost,
			Body: CR{
				"name":   "",
				"gender": "",
				"date":   "",
			},
			Status: http.StatusInternalServerError,
			Result: CR{"error": "empty actor"},
		},
		Case{
			Path:   "/api/actors/2", // список таблиц
			Method: http.MethodPost,
			Body: CR{
				"name":   "Леонардо Ди Каприо",
				"gender": "male",
				"date":   "",
			},
			Status: http.StatusOK,
			Result: CR{"data": CR{"id": 2, "name": "Леонардо Ди Каприо", "gender": "male", "date": "", "films": nil}},
		},
		Case{
			Path:   "/api/actors/2/date", // список таблиц
			Method: http.MethodPost,
			Body: CR{
				"name":   "Леонардо Ди Каприо",
				"gender": "male",
				"date":   "11-11-1974",
			},
			Status: http.StatusOK,
			Result: CR{"data": CR{"id": 2, "name": "Леонардо Ди Каприо", "gender": "male", "date": "11-11-1974", "films": nil}},
		},
		Case{
			Path:   "/api/actors/2/date", // список таблиц
			Method: http.MethodPost,
			Body: CR{
				"name":   "Леонардо Ди Каприо",
				"gender": CR{},
				"date":   "11-11-1974",
			},
			Status: http.StatusBadRequest,
			Result: CR{"error": "decode JSON error"},
		},
		Case{
			Path:   "/api/actors/pr", // список таблиц
			Method: http.MethodPost,
			Body: CR{
				"name":   "Леонардо Ди Каприо",
				"gender": CR{},
				"date":   "11-11-1974",
			},
			Status: http.StatusBadRequest,
			Result: CR{"error": "strconv.ParseUint: parsing \"pr\": invalid syntax"},
		},
		Case{
			Path:   "/api/actors/pr/name", // список таблиц
			Method: http.MethodPost,
			Body: CR{
				"name":   "Леонардо Ди Каприо",
				"gender": CR{},
				"date":   "11-11-1974",
			},
			Status: http.StatusBadRequest,
			Result: CR{"error": "strconv.ParseUint: parsing \"pr\": invalid syntax"},
		},
		Case{
			Path:   "/api/actors/2/hello", // список таблиц
			Method: http.MethodPost,
			Body: CR{
				"name":   "Леонардо Ди Каприо",
				"gender": "male",
				"date":   "11-11-1974",
			},
			Status: http.StatusBadRequest,
			Result: CR{"error": "wrong column name"},
		},
		Case{
			Path:   "/api/actors/2/name", // список таблиц
			Method: http.MethodPost,
			Body: CR{
				"name":   "",
				"gender": "",
				"date":   "",
			},
			Status: http.StatusInternalServerError,
			Result: CR{"error": "empty actor"},
		},
		Case{
			Path:   "/api/actors/100/name", // список таблиц
			Method: http.MethodPost,
			Body: CR{
				"name":   "hello",
				"gender": "it's me",
				"date":   "",
			},
			Status: http.StatusInternalServerError,
			Result: CR{"error": "sql: no rows in result set"},
		},
		Case{
			Path:   "/api/actors/1", // список таблиц
			Method: http.MethodPost,
			Body: CR{
				"name":   CR{},
				"gender": "",
				"date":   "",
			},
			Status: http.StatusBadRequest,
			Result: CR{"error": "decode JSON error"},
		},
		Case{
			Path:   "/api/actors/1/date", // список таблиц
			Method: http.MethodPost,
			Body: CR{
				"name":   "hello",
				"gender": "",
				"date":   "",
			},
			Status: http.StatusOK,
			Result: CR{"data": CR{"id": 1, "name": "hello", "gender": "", "date": "", "films": nil}},
		},
	}

	runCases(t, ts, db, cases)
}

func fakeExplorer(db *sql.DB, logger *zap.SugaredLogger) (http.Handler, error) {
	// тут вы пишете код
	// обращаю ваше внимание - в этом задании запрещены глобальные переменные
	itemRepo := items.NewMemoryRepo(db)
	userRepo := users.NewMemoryRepo(db)

	actorHandler := &handlers.ActorsHandler{
		ActorsRepo: itemRepo,
		Logger:     logger,
	}
	filmHandler := &handlers.FilmsHandler{
		FilmsRepo: itemRepo,
		Logger:    logger,
	}
	userHandler := &handlers.UsersHandler{
		UserRepo: userRepo,
		Logger:   logger,
	}

	router := mux.NewRouter()

	router.HandleFunc("/api/actors", actorHandler.CreateActor).Methods("POST")
	router.HandleFunc("/api/actors", actorHandler.GetActors).Methods("GET")
	router.HandleFunc("/api/actors/{ACTOR_ID}", actorHandler.GetActor).Methods("GET")
	router.HandleFunc("/api/actors/{ACTOR_ID}", actorHandler.UpdateActor).Methods("POST")
	router.HandleFunc("/api/actors/{ACTOR_ID}/{COLUMN_NAME}", actorHandler.UpdateColumnActor).Methods("POST")

	router.HandleFunc("/api/films/search", filmHandler.SearchFilm).Methods("GET")
	router.HandleFunc("/api/films", filmHandler.GetFilms).Methods("GET")
	router.HandleFunc("/api/films", filmHandler.CreateFilm).Methods("POST")
	router.HandleFunc("/api/films/{FILM_ID}", filmHandler.UpdateFilm).Methods("POST")
	router.HandleFunc("/api/films/{FILM_ID}/{COLUMN_NAME}", filmHandler.UpdateColumnFilm).Methods("POST")

	router.HandleFunc("/api/login", userHandler.Login).Methods("POST")
	router.HandleFunc("/api/register", userHandler.Register).Methods("POST")

	return router, nil
}
