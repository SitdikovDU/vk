package items

import (
	"database/sql"
	"errors"
	"filmlibrary/pkg/errs"
)

type Actor struct {
	ID     uint32      `json:"id"`
	Name   string      `json:"name"`
	Gender string      `json:"gender"`
	Date   interface{} `json:"date"`
	Films  interface{} `json:"films"`
}

func (actor Actor) Empty() error {
	if actor.Date == "" && actor.Gender == "" && actor.Name == "" {
		return errors.New(errs.EmptyActorError)
	}

	return nil
}

type Film struct {
	ID          uint32      `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Date        interface{} `json:"date"`
	Rating      interface{} `json:"rating"`
	Actors      []Actor     `json:"actors"`
}

type ItemRepo interface {
	CreateFilm(film Film) (uint32, error)
	GetFilmByID(id uint32) (Film, error)
	GetFilms(field string, order int) ([]Film, error)
	UpdateFilm(film Film) error
	UpdateColumnFilm(film Film, columnName string) error
	SearchFilm(searchQuery string) ([]Film, error)
	DeleteActors(filmID uint32) error
	InsertActors(filmID uint32, actors []Actor) error
	GetActorFilms(actor Actor) ([]Film, error)

	CreateActor(actor Actor) (uint32, error)
	GetActorByID(id uint32) (Actor, error)
	GetActors() ([]Actor, error)
	UpdateActor(actor Actor) error
	UpdateColumnActor(actor Actor, columnName string) error
	ActorsByFilm(film Film) ([]Actor, error)
}

type ItemMemoryRepository struct {
	DB *sql.DB
}

func NewMemoryRepo(db *sql.DB) *ItemMemoryRepository {
	return &ItemMemoryRepository{
		DB: db,
	}
}
