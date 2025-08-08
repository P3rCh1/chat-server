package user

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"sync"

	"github.com/P3rCh1/chat-server/user-service/internal/config"
	"github.com/P3rCh1/chat-server/user-service/internal/gRPC/status_error"
	"github.com/P3rCh1/chat-server/user-service/internal/models"
	"github.com/P3rCh1/chat-server/user-service/internal/storage/cache"
	"github.com/P3rCh1/chat-server/user-service/internal/storage/database"
	sessionpb "github.com/P3rCh1/chat-server/user-service/shared/proto/gen/go/session"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type UserService struct {
	log           *slog.Logger
	pq            *database.Postgres
	redis         *cache.Cacher
	sessionClient sessionpb.SessionClient
	sessionConn   *grpc.ClientConn
}

func MustNew(log *slog.Logger, cfg *config.Config) *UserService {
	const op = "user.MustNew"
	var errPQ, errCache, errListen error
	user := &UserService{log: log}
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		user.pq, errPQ = database.New(cfg.Postgres)
	}()
	go func() {
		defer wg.Done()
		user.redis, errCache = cache.New(cfg.Redis)
	}()
	user.sessionConn, errListen = grpc.NewClient(cfg.SessionAddr, grpc.WithTransportCredentials(insecure.NewCredentials())) //TODO to config
	if errListen == nil {
		user.sessionClient = sessionpb.NewSessionClient(user.sessionConn)
	}
	wg.Wait()
	if err := errors.Join(errPQ, errCache, errListen); err != nil {
		log.Error(op, "error", err)
		user.CloseNotNil()
		os.Exit(1)
	}
	return user
}

func (s *UserService) CloseNotNil() {
	if s.pq != nil {
		s.pq.Close()
	}
	if s.redis != nil {
		s.redis.Close()
	}
	if s.sessionConn != nil {
		s.sessionConn.Close()
	}
}

func (s *UserService) Close() {
	s.pq.Close()
	s.redis.Close()
	s.sessionConn.Close()
}

func (s *UserService) Register(
	ctx context.Context,
	username,
	email,
	password string,
) (int, error) {
	const op = "user.Register"

	profile := &models.Profile{
		Username: username,
		Email:    email,
	}
	err := s.pq.CreateUser(ctx, profile, password)
	if err != nil {
		if status_error.IsStatusError(err) {
			return 0, err
		}
		s.log.Error(op, "error", err)
		return 0, fmt.Errorf("create user error: %w", err)
	}
	go func() {
		err := s.redis.Set(profile)
		if err != nil {
			s.log.Error(op, "error", err)
		}
	}()
	return profile.ID, nil
}

func (s *UserService) Login(
	ctx context.Context,
	email, password string,
) (string, error) {
	const op = "user.Login"
	profile, err := s.pq.Login(ctx, email, password)
	if err != nil {
		if status_error.IsStatusError(err) {
			return "", err
		}
		return "", fmt.Errorf("login error: %w", err)
	}
	go func() {
		err := s.redis.Set(profile)
		if err != nil {
			s.log.Error(op, "error", err)
		}
	}()
	resp, err := s.sessionClient.Generate(ctx, &sessionpb.GenerateRequest{
		Id: int32(profile.ID),
	})
	if err != nil {
		s.log.Error(op, "error", err)
		return "", fmt.Errorf("session-service error: %w", err)
	}
	return resp.Token, nil
}

func (s *UserService) ChangeName(
	ctx context.Context,
	id int,
	newName string,
) error {
	const op = "user.ChangeName"
	profile, err := s.Profile(ctx, id)
	if err != nil {
		return err
	}
	if newName == profile.Username {
		return status_error.NamesAreSame
	}
	err = s.pq.ChangeName(ctx, id, newName)
	if err != nil {
		if status_error.IsStatusError(err) {
			return err
		}
		s.log.Error(op, "error", err)
		return fmt.Errorf("change name error: %w", err)
	}
	go func() {
		profile.Username = newName
		err := s.redis.Set(profile)
		if err != nil {
			s.log.Error(op, "error", err)
		}
	}()
	return nil
}

func (s *UserService) Profile(
	ctx context.Context,
	userID int,
) (*models.Profile, error) {
	const op = "user.Profile"
	profile, err := s.redis.Get(userID)
	if profile != nil {
		return profile, nil
	}
	if err != nil {
		s.log.Error(op, "error", err)
	}
	profile, err = s.pq.Profile(ctx, userID)
	if err != nil {
		if status_error.IsStatusError(err) {
			return nil, err
		}
		s.log.Error(op, "error", err)
		return nil, fmt.Errorf("get profile error: %w", err)
	}
	go func() {
		err := s.redis.Set(profile)
		if err != nil {
			s.log.Error(op, "error", err)
		}
	}()
	return profile, nil
}
