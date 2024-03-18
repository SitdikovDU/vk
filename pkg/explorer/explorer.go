package explorer

import (
	"database/sql"
	_ "filmlibrary/docs"
	"filmlibrary/pkg/handlers"
	"filmlibrary/pkg/items"
	"filmlibrary/pkg/middleware"
	"filmlibrary/pkg/users"
	"net/http"

	httpSwagger "github.com/swaggo/http-swagger"

	"github.com/gorilla/mux"

	"go.uber.org/zap"
)

func NewExplorer(db *sql.DB, logger *zap.SugaredLogger) (http.Handler, error) {
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

	router.PathPrefix("/swagger/").Handler(httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"), // Путь к вашему файлу swagger.json
	))

	router.HandleFunc("/api/login", userHandler.Login).Methods("POST")
	router.HandleFunc("/api/register", userHandler.Register).Methods("POST")

	myMux := middleware.Auth(logger, router, userRepo)
	myMux = middleware.AccessLog(logger, myMux)
	myMux = middleware.Panic(logger, myMux)

	return myMux, nil
}
