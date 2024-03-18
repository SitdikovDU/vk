package tests

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

func PrepareFilms(db *sql.DB) {
	qs := []string{
		`DROP TABLE IF EXISTS users;`,
		`DROP TABLE IF EXISTS film_actor cascade;`,
		`DROP TABLE IF EXISTS films cascade;`,
		`DROP TABLE IF EXISTS actors cascade;`,
		`CREATE TABLE users (
			id SERIAL PRIMARY KEY,
			username VARCHAR(50) NOT NULL,
			role VARCHAR(50) NOT NULL DEFAULT 'user',
			hashed_password VARCHAR(255) NOT NULL
		);`,
		`INSERT INTO users (username, role, hashed_password) VALUES
		('admin', 'admin','$2a$10$GJJT0w7LaC4oOfTygVVi8Ofn8D5kZHTXWdhA8PzscqkK846A2oG.C');`,
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

func TestFilms(t *testing.T) {
	zapLogger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("failed to create logger: %v", err)
		return
	}

	defer func() {
		syncErr := zapLogger.Sync()
		if syncErr != nil {
			log.Println("Error syncing logger:", syncErr)
		}
	}()

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

	PrepareFilms(db)

	handler, err := fakeExplorer(db, logger) //nolint:typecheck
	if err != nil {
		panic(err)
	}

	ts := httptest.NewServer(handler)

	cases := []Case{{
		Path:   "/api/login",
		Method: http.MethodPost,
		Body: CR{
			"username": "admin",
			"password": "MySuperSecretPassword",
		},
		Status: http.StatusOK,
		Result: CR{"data": 1},
	},
		{
			Path:   "/api/films",
			Method: http.MethodPost,
			Body: CR{
				"Name":        "Властелин колец: Две крепости",
				"Description": "фильм",
				"Date":        2002,
				"Rating":      10,
			},
			Status: http.StatusCreated,
			Result: CR{"data": 1},
		},
		{
			Path:   "/api/films",
			Method: http.MethodPost,
			Body: CR{
				"Name":        "Властелин колец: Возвращение короля",
				"Description": "фильм",
				"Date":        2003,
				"Rating":      9,
			},
			Status: http.StatusCreated,
			Result: CR{"data": 2},
		},
		{
			Path:   "/api/films/search",
			Query:  "query=паук",
			Method: http.MethodGet,
			Body:   CR{},
			Status: http.StatusOK,
			Result: CR{"data": nil},
		},
		{
			Path:   "/api/films/search",
			Method: http.MethodGet,
			Body:   CR{},
			Status: http.StatusBadRequest,
			Result: CR{"error": "empty search"},
		},
		{
			Path:   "/api/films/search",
			Query:  "query=Властелин",
			Method: http.MethodGet,
			Body:   CR{},
			Status: http.StatusOK,
			Result: CR{"data": []interface{}{CR{"actors": nil, "date": 2003, "description": "фильм", "id": 2, "name": "Властелин колец: Возвращение короля", "rating": 9}, CR{"actors": nil, "date": 2002, "description": "фильм", "id": 1, "name": "Властелин колец: Две крепости", "rating": 10}}},
		},
		{
			Path:   "/api/films",
			Query:  "field=name",
			Method: http.MethodGet,
			Body:   CR{},
			Status: http.StatusOK,
			Result: CR{"data": []interface{}{CR{"actors": nil, "date": 2002, "description": "фильм", "id": 1, "name": "Властелин колец: Две крепости", "rating": 10},
				CR{"actors": nil, "date": 2003, "description": "фильм", "id": 2, "name": "Властелин колец: Возвращение короля", "rating": 9}}},
		},
		{
			Path:   "/api/films",
			Query:  "field=name&order=-1",
			Method: http.MethodGet,
			Body:   CR{},
			Status: http.StatusOK,
			Result: CR{"data": []interface{}{CR{"actors": nil, "date": 2002, "description": "фильм", "id": 1, "name": "Властелин колец: Две крепости", "rating": 10},
				CR{"actors": nil, "date": 2003, "description": "фильм", "id": 2, "name": "Властелин колец: Возвращение короля", "rating": 9}}},
		},
		{
			Path:   "/api/films",
			Query:  "field=bad&order=-1",
			Method: http.MethodGet,
			Body:   CR{},
			Status: http.StatusBadRequest,
			Result: CR{"error": "incorrect orderBy"},
		},
		{
			Path:   "/api/films",
			Query:  "field=name&order=hello",
			Method: http.MethodGet,
			Body:   CR{},
			Status: http.StatusBadRequest,
			Result: CR{"error": "error reading order"},
		},
		{
			Path:   "/api/films",
			Query:  "field=rating&order=-1",
			Method: http.MethodGet,
			Body:   CR{},
			Status: http.StatusOK,
			Result: CR{"data": []interface{}{CR{"actors": nil, "date": 2002, "description": "фильм", "id": 1, "name": "Властелин колец: Две крепости", "rating": 10},
				CR{"actors": nil, "date": 2003, "description": "фильм", "id": 2, "name": "Властелин колец: Возвращение короля", "rating": 9}}},
		},
		{
			Path:   "/api/films",
			Query:  "field=name&order=1",
			Method: http.MethodGet,
			Body:   CR{},
			Status: http.StatusOK,
			Result: CR{"data": []interface{}{CR{"actors": nil, "date": 2003, "description": "фильм", "id": 2, "name": "Властелин колец: Возвращение короля", "rating": 9}, CR{"actors": nil, "date": 2002, "description": "фильм", "id": 1, "name": "Властелин колец: Две крепости", "rating": 10}}},
		},
		{
			Path:   "/api/films",
			Method: http.MethodPost,
			Body: CR{
				"Name":        "Человек-паук",
				"Description": "фильм",
				"Rating":      10,
			},
			Status: http.StatusCreated,
			Result: CR{"data": 3},
		},
		{
			Path:   "/api/films",
			Method: http.MethodPost,
			Body: CR{
				"Name":        CR{},
				"Description": "фильм",
				"Rating":      10,
			},
			Status: http.StatusBadRequest,
			Result: CR{"error": "decode JSON error"},
		},
		{
			Path:   "/api/films",
			Method: http.MethodPost,
			Body: CR{
				"Name":        "error",
				"Description": "фильм",
				"Rating":      CR{},
			},
			Status: http.StatusInternalServerError,
			Result: CR{"error": "sql: converting argument $4 type: unsupported type map[string]interface {}, a map"},
		},
		{
			Path:   "/api/films",
			Method: http.MethodPost,
			Body: CR{
				"Name":        "Человек-паук",
				"Description": "фильм",
				"Rating":      10,
			},
			Status: http.StatusCreated,
			Result: CR{"data": 4},
		},
		{
			Path:   "/api/films/2",
			Method: http.MethodPost,
			Body: CR{
				"Name":        "Человек-паук",
				"Description": "фильм",
				"Date":        2002,
				"Rating":      10,
			},
			Status: http.StatusOK,
			Result: CR{"data": 2},
		},
		{
			Path:   "/api/films/2",
			Method: http.MethodPost,
			Body: CR{
				"Name":   "Человек-паук",
				"Date":   2002,
				"Rating": 10,
			},
			Status: http.StatusOK,
			Result: CR{"data": 2},
		},
		{
			Path:   "/api/films/2",
			Method: http.MethodPost,
			Body: CR{
				"Name": "Человек-паук",
			},
			Status: http.StatusOK,
			Result: CR{"data": 2},
		},
		{
			Path:   "/api/films/2/Rating",
			Method: http.MethodPost,
			Body: CR{
				"Name":        "Человек-паук",
				"Description": "фильм",
				"Date":        2002,
				"Rating":      10,
			},
			Status: http.StatusOK,
			Result: CR{"data": 2},
		},
		{
			Path:   "/api/films/2/Rating",
			Method: http.MethodPost,
			Body: CR{
				"Name":        "Человек-паук",
				"Description": "фильм",
				"Date":        2002,
				"Rating":      CR{},
			},
			Status: http.StatusInternalServerError,
			Result: CR{"error": "sql: converting argument $1 type: unsupported type map[string]interface {}, a map"},
		},
		{
			Path:   "/api/films/2/Rating",
			Method: http.MethodPost,
			Body: CR{
				"Name":        CR{},
				"Description": "фильм",
				"Date":        2002,
				"Rating":      CR{},
			},
			Status: http.StatusBadRequest,
			Result: CR{"error": "json: cannot unmarshal object into Go struct field Film.name of type string"},
		},
		{
			Path:   "/api/films/oovrv/Rating",
			Method: http.MethodPost,
			Body: CR{
				"Name":        CR{},
				"Description": "фильм",
				"Date":        2002,
				"Rating":      CR{},
			},
			Status: http.StatusBadRequest,
			Result: CR{"error": "strconv.ParseUint: parsing \"oovrv\": invalid syntax"},
		},
		{
			Path:   "/api/films/oovrv",
			Method: http.MethodPost,
			Body: CR{
				"Name":        CR{},
				"Description": "фильм",
				"Date":        2002,
				"Rating":      CR{},
			},
			Status: http.StatusBadRequest,
			Result: CR{"error": "strconv.ParseUint: parsing \"oovrv\": invalid syntax"},
		},
		{
			Path:   "/api/films/1000000",
			Method: http.MethodPost,
			Body: CR{
				"Name":        "good",
				"Description": "фильм",
				"Date":        2002,
				"Rating":      10,
			},
			Status: http.StatusInternalServerError,
			Result: CR{"error": "sql: no rows in result set"},
		},
		{
			Path:   "/api/films/1",
			Method: http.MethodPost,
			Body: CR{
				"Name":        "good",
				"Description": CR{},
				"Date":        2002,
				"Rating":      10,
			},
			Status: http.StatusBadRequest,
			Result: CR{"error": "decode JSON error"},
		},
		{
			Path:   "/api/films/1/unknown",
			Method: http.MethodPost,
			Body: CR{
				"Name":        "good",
				"Description": CR{},
				"Date":        2002,
				"Rating":      10,
			},
			Status: http.StatusBadRequest,
			Result: CR{"error": "wrong column name"},
		},
		{
			Path:   "/api/films/1/Actors",
			Method: http.MethodPost,
			Body: CR{
				"Name":        "good",
				"Description": "film",
				"Date":        2002,
				"Rating":      10,
				"Actors": []CR{
					{"ID": 0, "Name": "Тоби Магуайр", "gender": "Мужской", "Date": ""},
					{"ID": 1, "Name": "Эндрю Гарфилд", "gender": "Мужской", "Date": ""},
				},
			},
			Status: http.StatusOK,
			Result: CR{"data": 1},
		},
		{
			Path:   "/api/films/1",
			Method: http.MethodPost,
			Body: CR{
				"Name":        "good",
				"Description": "film",
				"Date":        2002,
				"Rating":      10,
				"Actors": []CR{
					{"ID": 0, "Name": "Тоби Магуайр", "gender": "Мужской", "Date": ""},
					{"ID": 1, "Name": "Эндрю Гарфилд", "gender": "Мужской", "Date": ""},
				},
			},
			Status: http.StatusOK,
			Result: CR{"data": 1},
		},
		{
			Path:   "/api/films",
			Method: http.MethodPost,
			Body: CR{
				"Name":        "goodFILM",
				"Description": "film",
				"Date":        2010,
				"Actors": []CR{
					{"ID": 0, "Name": "Тоби Магуайр", "gender": "Мужской", "Date": ""},
					{"ID": 1, "Name": "Эндрю Гарфилд", "gender": "Мужской", "Date": ""},
				},
			},
			Status: http.StatusCreated,
			Result: CR{"data": 5},
		},
		{
			Path:   "/api/films",
			Method: http.MethodPost,
			Body: CR{
				"Name": "Человек-паук 2",
				"Actors": []CR{
					{"ID": 0, "Name": "Тоби Магуайр", "gender": "Мужской", "Date": ""},
					{"ID": 1, "Name": "Эндрю Гарфилд", "gender": "Мужской", "Date": ""},
				},
			},
			Status: http.StatusCreated,
			Result: CR{"data": 6},
		},
		{
			Path:   "/api/films",
			Method: http.MethodPost,
			Body: CR{
				"Name": "Человек-паук 3",
				"Actors": []CR{
					{"ID": 0, "Name": "Тоби Магуайр", "gender": "Мужской", "Date": ""},
					{"ID": 1, "Name": "Эндрю Гарфилд", "gender": "Мужской", "Date": ""},
				},
			},
			Status: http.StatusCreated,
			Result: CR{"data": 7},
		},
	}

	runCases(t, ts, db, cases)
}
