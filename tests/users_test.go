package tests

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"filmlibrary/pkg/explorer"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

// CaseResponse
type CR map[string]interface{}

type Case struct {
	Method string // GET по-умолчанию в http.NewRequest если передали пустую строку
	Path   string
	Query  string
	Status int
	Result interface{}
	Body   interface{}
}

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "mysecretpassword"
	dbname   = "postgres"
)

var (
	client = &http.Client{Timeout: time.Second}
)

func PrepareTestApis(db *sql.DB) {
	qs := []string{
		`DROP TABLE IF EXISTS users;`,
		`CREATE TABLE users (
			id SERIAL PRIMARY KEY,
			username VARCHAR(50) NOT NULL,
			role VARCHAR(50) NOT NULL DEFAULT 'user',
			hashed_password VARCHAR(255) NOT NULL
		);`,

		`INSERT INTO users (username, role, hashed_password) VALUES
		('admin', 'admin','$2a$10$GJJT0w7LaC4oOfTygVVi8Ofn8D5kZHTXWdhA8PzscqkK846A2oG.C');`,
	}

	for _, q := range qs {
		_, err := db.Exec(q)
		if err != nil {
			panic(err)
		}
	}
}

func PrepareTestUsers2(db *sql.DB) {
	qs := []string{
		`DROP TABLE IF EXISTS users;`,
		`CREATE TABLE users (
			id SERIAL PRIMARY KEY,
			username VARCHAR(50) NOT NULL,
			role VARCHAR(50) NOT NULL DEFAULT 'user',
			password VARCHAR(255) NOT NULL
		);`,

		`INSERT INTO users (username, role, password) VALUES
		('admin', 'admin','$2a$10$GJJT0w7LaC4oOfTygVVi8Ofn8D5kZHTXWdhA8PzscqkK846A2oG.C');`,
	}

	for _, q := range qs {
		_, err := db.Exec(q)
		if err != nil {
			panic(err)
		}
	}
}
func CleanupTestApis(db *sql.DB) {
	qs := []string{
		`DROP TABLE IF EXISTS items;`,
		`DROP TABLE IF EXISTS users;`,
	}
	for _, q := range qs {
		_, err := db.Exec(q)
		if err != nil {
			panic(err)
		}
	}
}

func TestUsers(t *testing.T) {
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

	PrepareTestApis(db)

	// возможно вам будет удобно закомментировать это, чтобы смотреть результат после теста
	defer CleanupTestApis(db)

	handler, err := explorer.NewExplorer(db, logger) //nolint:typecheck
	if err != nil {
		panic(err)
	}

	ts := httptest.NewServer(handler)

	cases := []Case{
		Case{
			Path:   "/api/register", // список таблиц
			Method: http.MethodPost,
			Body: CR{
				"username": "hello",
				"password": "privetMir",
			},
			Status: http.StatusCreated,
			Result: CR{"data": 2},
		},
		Case{
			Path:   "/api/register", // список таблиц
			Method: http.MethodPost,
			Body: CR{
				"username": "hello",
				"password": "privetMir",
			},
			Status: http.StatusUnprocessableEntity,
			Result: CR{"error": "user already exists"},
		},
		Case{
			Path:   "/api/register", // список таблиц
			Method: http.MethodPost,
			Body: CR{
				"username": "helloWorld",
				"password": "pri",
			},
			Status: http.StatusUnprocessableEntity,
			Result: CR{"error": "short password"},
		},
		Case{
			Path:   "/api/login", // список таблиц
			Method: http.MethodPost,
			Body: CR{
				"username": "hello",
				"password": "privetMir",
			},
			Status: http.StatusOK,
			Result: CR{"data": 2},
		},
		Case{
			Path:   "/api/login", // список таблиц
			Method: http.MethodPost,
			Body: CR{
				"usee":     "hello",
				"password": "privetMir",
			},
			Status: http.StatusUnauthorized,
			Result: CR{"error": "empty username"},
		},
		Case{
			Path:   "/api/login", // список таблиц
			Method: http.MethodPost,
			Body:   CR{},
			Status: http.StatusUnauthorized,
			Result: CR{"error": "empty username"},
		},
		Case{
			Path:   "/api/login", // список таблиц
			Method: http.MethodPost,
			Body: CR{
				"username": CR{},
				"password": "privetMir"},
			Status: http.StatusBadRequest,
			Result: CR{"error": "decode JSON error"},
		},
		Case{
			Path:   "/api/register", // список таблиц
			Method: http.MethodPost,
			Body: CR{
				"username": CR{},
				"password": "privetMir"},
			Status: http.StatusBadRequest,
			Result: CR{"error": "decode JSON error"},
		},
		Case{
			Path:   "/api/register", // список таблиц
			Method: http.MethodPost,
			Body: CR{
				"username": "admin",
				"password": "privetMir"},
			Status: http.StatusUnprocessableEntity,
			Result: CR{"error": "user already exists"},
		},
		Case{
			Path:   "/api/register", // список таблиц
			Method: http.MethodPost,
			Body: CR{
				"username": "",
				"password": "privetMir"},
			Status: http.StatusUnprocessableEntity,
			Result: CR{"error": "empty username"},
		},
		Case{
			Path:   "/api/login", // список таблиц
			Method: http.MethodPost,
			Body: CR{
				"username": "priiiivet",
				"password": "privetMir"},
			Status: http.StatusUnauthorized,
			Result: CR{"error": "user not exist"},
		},
		Case{
			Path:   "/api/register", // список таблиц
			Method: http.MethodPost,
			Body: CR{
				"username": "priiiivetpriiiivetpriiiivetpriiiivetpriiiivetpriiiivetpriiiivetpriiiivetpriiiivetpriiiivetpriiiivetpriiiivetpriiiivetpriiiivetpriiiivet",
				"password": "privetMir"},
			Status: http.StatusUnprocessableEntity,
			Result: CR{"error": "DB error"},
		},
		Case{
			Path:   "/api/login", // список таблиц
			Method: http.MethodPost,
			Body: CR{
				"username": "admin",
				"password": "privetMir"},
			Status: http.StatusUnauthorized,
			Result: CR{"error": "invalid password"},
		},
	}

	runCases(t, ts, db, cases)
}

func TestUsers2(t *testing.T) {
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

	PrepareTestUsers2(db)

	// возможно вам будет удобно закомментировать это, чтобы смотреть результат после теста
	defer CleanupTestApis(db)

	handler, err := explorer.NewExplorer(db, logger) //nolint:typecheck
	if err != nil {
		panic(err)
	}

	ts := httptest.NewServer(handler)

	cases := []Case{
		Case{
			Path:   "/api/register", // список таблиц
			Method: http.MethodPost,
			Body: CR{
				"username": "hello",
				"password": "privetMir",
			},
			Status: http.StatusUnprocessableEntity,
			Result: CR{"error": "DB error"},
		},
		Case{
			Path:   "/api/login", // список таблиц
			Method: http.MethodPost,
			Body: CR{
				"username": "admin",
				"password": "privetMir",
			},
			Status: http.StatusUnauthorized,
			Result: CR{"error": "pq: column \"hashed_password\" does not exist"},
		},
	}

	runCases(t, ts, db, cases)
}

func runCases(t *testing.T, ts *httptest.Server, db *sql.DB, cases []Case) {
	for idx, item := range cases {
		var (
			result   interface{}
			expected interface{}
			req      *http.Request
		)

		caseName := fmt.Sprintf("case %d: [%s] %s %s", idx, item.Method, item.Path, item.Query)

		if db.Stats().OpenConnections != 1 {
			t.Fatalf("[%s] you have %d open connections, must be 1", caseName, db.Stats().OpenConnections)
		}

		if item.Method == "" || item.Method == http.MethodGet {
			var errNewReq error
			req, errNewReq = http.NewRequest(item.Method, ts.URL+item.Path+"?"+item.Query, nil)
			if errNewReq != nil {
				panic(errNewReq)
			}
		} else {
			data, errMarshal := json.Marshal(item.Body)
			if errMarshal != nil {
				panic(errMarshal)
			}
			reqBody := bytes.NewReader(data)
			var errNewReq error
			req, errNewReq = http.NewRequest(item.Method, ts.URL+item.Path, reqBody)
			if errNewReq != nil {
				panic(errNewReq)
			}
			req.Header.Add("Content-Type", "application/json")
		}

		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("[%s] request error: %v", caseName, err)
			continue
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("[%s] error readall: %v", caseName, err)
			continue
		}

		if item.Status == 0 {
			item.Status = http.StatusOK
		}

		if resp.StatusCode != item.Status {
			t.Fatalf("[%s] expected http status %v, got %v", caseName, item.Status, resp.StatusCode)
			continue
		}

		err = json.Unmarshal(body, &result)
		if err != nil {
			t.Fatalf("[%s] cant unpack json: %v", caseName, err)
			continue
		}

		data, err := json.Marshal(item.Result)
		if err != nil {
			t.Fatalf("[%s] cant marshal result json: %v", caseName, err)
			continue
		}
		err = json.Unmarshal(data, &expected)
		if err != nil {
			t.Fatalf("[%s] cant unmarshal expected json: %v", caseName, err)
			continue
		}

		if !reflect.DeepEqual(result, expected) {
			t.Fatalf("[%s] results not match\nGot : %#v\nWant: %#v", caseName, result, expected)
			continue
		}
	}

}
