package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const key string = "3a56109ff7acebe22680a9ad2fad15185be84ef589882dfdbb623c229185fcdf"

func GenJWT(id int) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": id,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	})
	return token.SignedString([]byte(key))
}

func VerifyJWT(tokenString string) (int, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		return []byte(key), nil
	})
	if err != nil {
		return 0, err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return int(claims["user_id"].(float64)), nil
	}
	return 0, errors.New("Недействительный токен")
}
