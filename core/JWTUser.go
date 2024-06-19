package core

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTUser struct {
	Username string
	Password string
	Token    string
}

func (ju *JWTUser) GenerateToken() error {

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"exp": time.Now().Add(time.Hour * 48).Unix(), // 2 day
	})

	tokenStr, err := token.SignedString([]byte(os.Getenv("SECRET")))
	ju.Token = tokenStr

	return err
}

func (ju *JWTUser) Resolve() error {

	if ju.Token == "" {
		return errors.New("empty_token_string")
	}

	_, err := jwt.Parse(ju.Token, func(token *jwt.Token) (interface{}, error) {

		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(os.Getenv("SECRET")), nil
	})

	if err != nil {
		return errors.New("invalid_token")
	}

	return nil
}
