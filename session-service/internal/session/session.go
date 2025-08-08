package session

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/P3rCh1/chat-server/session/internal/config"
	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrTokenExpired = errors.New("token expired")
)

type SessionService struct {
	Secret []byte
	Expire time.Duration
	log    *slog.Logger
}

func New(cfg *config.Config, log *slog.Logger) *SessionService {
	return &SessionService{
		Secret: []byte(cfg.Secret),
		Expire: cfg.Expire,
		log:    log,
	}
}

func (s *SessionService) Generate(ctx context.Context, id int) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"uid": id,
		"exp": time.Now().Add(s.Expire).Unix(),
	})
	return token.SignedString(s.Secret)
}

func (s *SessionService) Verify(ctx context.Context, tokenString string) (int, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		return s.Secret, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return 0, ErrTokenExpired
		}
		return 0, ErrInvalidToken
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return int(claims["uid"].(float64)), nil
	}
	return 0, ErrInvalidToken
}
