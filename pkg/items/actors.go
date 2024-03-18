package items

import (
	"errors"
	"filmlibrary/pkg/errs"
	"fmt"
	"reflect"
)

func (repo *ItemMemoryRepository) GetActors() ([]Actor, error) {
	rows, err := repo.DB.Query("SELECT id, name, gender, date FROM actors")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var actors []Actor
	for rows.Next() {
		var actor Actor
		err := rows.Scan(&actor.ID, &actor.Name, &actor.Gender, &actor.Date)
		if err != nil {
			return nil, err
		}

		actor.Films, err = repo.GetActorFilms(actor)
		if err != nil {
			return nil, err
		}

		actors = append(actors, actor)
	}
	return actors, nil
}

func (repo *ItemMemoryRepository) GetActorByID(id uint32) (Actor, error) {
	var actor Actor

	stmt, err := repo.DB.Prepare("SELECT id, name, gender, date FROM actors WHERE id = $1")
	if err != nil {
		return Actor{}, err
	}
	defer stmt.Close()

	err = stmt.QueryRow(id).Scan(&actor.ID, &actor.Name, &actor.Gender, &actor.Date)
	if err != nil {
		return Actor{}, err
	}

	actor.Films, err = repo.GetActorFilms(actor)
	return actor, err
}

func (repo *ItemMemoryRepository) CreateActor(actor Actor) (uint32, error) {
	err := actor.Empty()
	if err != nil {
		return 0, err
	}

	switch actor.Date.(type) {
	case string:
		if actor.Date.(string) == "" {
			actor.Date = nil
		}
	default:
		actor.Date = nil
	}

	stmt, err := repo.DB.Prepare("INSERT INTO actors(name, gender, date) VALUES($1, $2, $3) RETURNING id")
	if err != nil {
		return 0, errors.New(errs.DatabaseError)
	}
	defer stmt.Close()

	err = stmt.QueryRow(actor.Name, actor.Gender, actor.Date).Scan(&actor.ID)
	return actor.ID, err
}

func (repo *ItemMemoryRepository) UpdateActor(actor Actor) error {
	id := actor.ID
	_, err := repo.GetActorByID(id)
	if err != nil {
		return err
	}

	err = actor.Empty()
	if err != nil {
		return err
	}

	if actor.Date.(string) == "" {
		actor.Date = nil
	}
	stmt, err := repo.DB.Prepare("UPDATE actors SET name = $1, gender = $2, date = $3 WHERE id = $4")
	if err != nil {
		return err
	}

	defer stmt.Close()

	_, err = stmt.Exec(actor.Name, actor.Gender, actor.Date, id)
	return err
}

func (repo *ItemMemoryRepository) UpdateColumnActor(actor Actor, columnName string) error {
	id := actor.ID
	_, err := repo.GetActorByID(id)
	if err != nil {
		return err
	}

	err = actor.Empty()
	if err != nil {
		return err
	}

	r := reflect.ValueOf(actor)
	fmt.Println(r, columnName, reflect.Indirect(r).FieldByName(columnName))
	value := reflect.Indirect(r).FieldByName(columnName).Interface()
	if value.(string) == "" {
		value = nil
	}
	query := fmt.Sprintf("UPDATE actors SET %s = $1 WHERE id = $2", columnName)
	stmt, err := repo.DB.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(value, id)

	return err
}

func (repo *ItemMemoryRepository) ActorsByFilm(film Film) ([]Actor, error) {
	stmt, err := repo.DB.Prepare(`
        SELECT id, name, gender, date 
        FROM actors 
        JOIN (SELECT actor_id FROM film_actor WHERE film_id = $1) AS film_actors
        ON actors.id = film_actors.actor_id`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(film.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var filmActors []Actor
	for rows.Next() {
		var actor Actor
		err := rows.Scan(&actor.ID, &actor.Name, &actor.Gender, &actor.Date)
		if err != nil {
			return nil, err
		}
		filmActors = append(filmActors, actor)
	}
	return filmActors, nil
}
