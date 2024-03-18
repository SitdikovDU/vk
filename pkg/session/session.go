package session

import (
	"errors"
	"filmlibrary/pkg/errs"
	"filmlibrary/pkg/users"
	"fmt"
	"net/http"
	"strings"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
)

var secret = []byte("secretKey")

func generateToken(user *users.User) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user": user,
		"iat":  time.Now().Unix(),
		"exp":  time.Now().Add(time.Hour * 24 * 7).Unix(),
	})

	strToken, err := token.SignedString(secret)
	if err != nil {
		return "", err
	}

	return strToken, nil
}

func GetUser(authStr string, repo *users.UserMemoryRepository) (*users.User, error) {
	auth := strings.Fields(authStr)

	if len(auth) < 2 || auth[0] != "Bearer" {
		return &users.User{}, nil
	}

	hashSecretGetter := func(token *jwt.Token) (interface{}, error) {
		method, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok || method.Alg() != "HS256" {
			return nil, fmt.Errorf("bad sign method")
		}
		return secret, nil
	}

	token, err := jwt.Parse(auth[1], hashSecretGetter)
	if err != nil {
		return &users.User{}, errors.New(errs.FailPassing)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return &users.User{}, errors.New(errs.InvalidToken)
	}

	u, ok := claims["user"].(map[string]interface{})
	if !ok {
		return &users.User{}, errors.New(errs.UserClaimsError)
	}

	user := users.User{}
	if user.Login, ok = u["username"].(string); !ok {
		return &users.User{}, errors.New(errs.UserClaimsError)
	}

	if exist, err := repo.UserExists(user.Login); !exist || err != nil {
		return &users.User{}, errors.New(errs.UserNotExist)
	}

	if user.Role, err = repo.GetUserRole(user.Login); err != nil {
		return &users.User{}, errors.New(errs.DatabaseError)
	}

	return &user, nil
}

func CreateToken(w http.ResponseWriter, user *users.User) error {
	token, err := generateToken(user)
	if err != nil {
		return err
	}

	w.Header().Set("Authorization", fmt.Sprintf("Bearer %s", token))

	return nil
}
