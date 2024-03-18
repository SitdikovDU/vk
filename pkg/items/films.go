package items

import (
	"fmt"
	"reflect"
)

func (repo *ItemMemoryRepository) GetFilms(field string, order int) ([]Film, error) {
	orderBy := "ORDER BY " + field + " DESC"
	if order == 1 {
		orderBy = "ORDER BY " + field + " ASC"
	}

	rows, err := repo.DB.Query("SELECT id, name, description, date, rating FROM films " + orderBy)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var films []Film
	for rows.Next() {
		var film Film
		err := rows.Scan(&film.ID, &film.Name, &film.Description, &film.Date, &film.Rating)
		if err != nil {
			return nil, err
		}
		films = append(films, film)
	}

	return films, nil
}

func (repo *ItemMemoryRepository) GetFilmByID(id uint32) (Film, error) {
	var film Film

	err := repo.DB.QueryRow("SELECT id, name, description, date, rating FROM films WHERE id = $1", id).Scan(&film.ID, &film.Name, &film.Description, &film.Date, &film.Rating)
	if err != nil {
		return Film{}, err
	}

	film.Actors, err = repo.ActorsByFilm(film)
	if err != nil {
		return film, err
	}

	return film, nil
}

func (repo *ItemMemoryRepository) DeleteActors(filmID uint32) error {
	_, err := repo.DB.Exec("DELETE FROM film_actor WHERE film_id = $1;", filmID)
	if err != nil {
		return err
	}

	return nil
}

func (repo *ItemMemoryRepository) InsertActors(filmID uint32, actors []Actor) error {
	actorIDs := make([]uint32, len(actors))
	for i, actor := range actors {
		actorIDs[i] = actor.ID
	}

	stmt, err := repo.DB.Prepare("INSERT INTO film_actor (film_id, actor_id) VALUES ($1, $2)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, actorID := range actorIDs {
		_, err := stmt.Exec(filmID, actorID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (repo *ItemMemoryRepository) CreateFilm(film Film) (uint32, error) {
	stmt, err := repo.DB.Prepare("INSERT INTO films(name, description, date, rating) VALUES($1, $2, $3, $4) RETURNING id")
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	err = stmt.QueryRow(film.Name, film.Description, film.Date, film.Rating).Scan(&film.ID)
	if err != nil {
		return 0, err
	}

	if len(film.Actors) == 0 {
		return uint32(film.ID), err
	}

	err = repo.InsertActors(film.ID, film.Actors)

	return uint32(film.ID), err
}

func (repo *ItemMemoryRepository) UpdateFilm(film Film) error {
	_, err := repo.GetFilmByID(film.ID)
	if err != nil {
		return err
	}

	_, err = repo.DB.Exec("UPDATE films SET name = $1, description = $2, date = $3, rating = $4 WHERE id = $5", film.Name, film.Description, film.Date, film.Rating, film.ID)
	if err != nil {
		return err
	}

	err = repo.DeleteActors(film.ID)
	if err != nil {
		return err
	}

	if len(film.Actors) == 0 {
		return err
	}

	err = repo.InsertActors(film.ID, film.Actors)
	return err
}

func (repo *ItemMemoryRepository) UpdateColumnFilm(film Film, columnName string) error {
	id := film.ID
	_, err := repo.GetFilmByID(id)
	if err != nil {
		return err
	}

	r := reflect.ValueOf(film)
	columnValue := reflect.Indirect(r).FieldByName(columnName).Interface()

	query := fmt.Sprintf("UPDATE films SET %s = $1 WHERE id = $2", columnName)
	_, err = repo.DB.Exec(query, columnValue, id)
	return err
}

func (repo *ItemMemoryRepository) SearchFilm(searchQuery string) ([]Film, error) {
	rows, err := repo.DB.Query("SELECT DISTINCT films.id, films.name, films.description, films.Date, films.rating FROM films WHERE films.name ILIKE $1 "+
		"UNION SELECT DISTINCT films.id, films.name, films.description, films.Date, films.rating FROM films JOIN film_actor ON films.id = film_actor.film_id "+
		"JOIN actors ON film_actor.actor_id = actors.id "+
		"WHERE actors.name ILIKE $2", "%"+searchQuery+"%", "%"+searchQuery+"%")

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var films []Film
	for rows.Next() {
		var film Film
		err := rows.Scan(&film.ID, &film.Name, &film.Description, &film.Date, &film.Rating)
		if err != nil {
			return nil, err
		}
		films = append(films, film)
	}

	return films, nil
}

func (repo *ItemMemoryRepository) GetActorFilms(actor Actor) ([]Film, error) {
	stmt, err := repo.DB.Prepare(`
        SELECT id, name, description, date, rating 
        FROM films
        JOIN (SELECT film_id FROM film_actor WHERE actor_id = $1) AS film_actors
        ON films.id = film_actors.film_id`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(actor.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var films []Film
	for rows.Next() {
		var film Film
		err := rows.Scan(&film.ID, &film.Name, &film.Description, &film.Date, &film.Rating)
		if err != nil {
			return nil, err
		}
		films = append(films, film)
	}
	return films, nil
}
