package tokens

import (
	"errors"
	"time"

	"github.com/P3rCh1/chat-server/internal/config"
	"github.com/golang-jwt/jwt/v5"
)

type JWTProvider struct {
	Secret []byte
	Expire time.Duration
}

func NewJWT(c *config.JWT) *JWTProvider {
	return &JWTProvider{
		Secret: []byte(c.Secret),
		Expire: c.Expire,
	}
}

func (c *JWTProvider) Gen(id int) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": id,
		"exp":     time.Now().Add(c.Expire).Unix(),
	})
	return token.SignedString(c.Secret)
}

func (c *JWTProvider) Verify(tokenString string) (int, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		return c.Secret, nil
	})
	if err != nil {
		return 0, err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return int(claims["user_id"].(float64)), nil
	}
	return 0, errors.New("invalid token")
}
