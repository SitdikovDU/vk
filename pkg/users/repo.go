package users

import (
	"database/sql"
	"errors"
	"filmlibrary/pkg/errs"

	"golang.org/x/crypto/bcrypt"
)

type UserMemoryRepository struct {
	DB *sql.DB
}

func NewMemoryRepo(db *sql.DB) *UserMemoryRepository {
	return &UserMemoryRepository{
		DB: db,
	}
}

func (repo *UserMemoryRepository) GetUserRole(username string) (string, error) {
	var role string
	err := repo.DB.QueryRow("SELECT role FROM users WHERE username = $1", username).Scan(&role)
	if err != nil {
		return "", err
	}
	return role, nil
}

func (repo *UserMemoryRepository) UserExists(username string) (bool, error) {
	var exists bool
	if username == "" {
		return false, errors.New(errs.EmptyUsernameError)
	}

	err := repo.DB.QueryRow("SELECT EXISTS (SELECT 1 FROM users WHERE username = $1)", username).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (repo *UserMemoryRepository) getUserByUsername(username string) (User, error) {
	var user User

	exist, err := repo.UserExists(username)
	if err != nil {
		return User{}, err
	}

	if !exist {
		return User{}, errors.New(errs.UserNotExist)
	}

	query := "SELECT id, username, hashed_password FROM users WHERE username=$1"
	stmt, err := repo.DB.Prepare(query)
	if err != nil {
		return User{}, err
	}
	defer stmt.Close()

	err = stmt.QueryRow(username).Scan(&user.ID, &user.Login, &user.password)
	if err != nil {
		return User{}, err
	}

	return user, nil
}

func (repo *UserMemoryRepository) Authorize(login, password string) (*User, error) {
	user, err := repo.getUserByUsername(login)
	if err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.password), []byte(password)); err != nil {
		return nil, errors.New(errs.BadPass)
	}

	return &user, nil
}

func (repo *UserMemoryRepository) Signup(username, pass string) (*User, error) {
	exist, err := repo.UserExists(username)
	if err != nil {
		return nil, err
	}
	if exist {
		return nil, errors.New(errs.UserExistError)
	}

	hashedPass, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New(errs.HashPasswordError)
	}

	user := &User{Login: username, password: string(hashedPass)}

	err = repo.DB.QueryRow("INSERT INTO users (username, hashed_password) VALUES ($1, $2) RETURNING id", username, string(hashedPass)).Scan(&user.ID)
	if err != nil {
		return nil, errors.New(errs.DatabaseError)
	}

	return user, nil
}
