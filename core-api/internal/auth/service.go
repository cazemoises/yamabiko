package auth

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

var ErrInvalidCredentials = errors.New("credenciais inválidas")

type Service struct {
	repo   UserRepository
	tokens *JWTIssuer
}

func NewService(repo UserRepository, tokens *JWTIssuer) *Service {
	return &Service{repo: repo, tokens: tokens}
}

func (s *Service) Register(ctx context.Context, email, password, name string) (*TokenPair, error) {
	_, err := s.repo.FindByEmail(ctx, email)
	if err == nil {
		return nil, ErrEmailTaken
	}
	if !errors.Is(err, ErrUserNotFound) {
		return nil, err
	}

	hash, err := HashPassword(password)
	if err != nil {
		return nil, err
	}

	user := &User{
		ID:               uuid.New(),
		Email:            email,
		PasswordHash:     hash,
		Name:             name,
		CurrentSprintDay: 1,
	}
	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	return s.tokens.IssueTokenPair(user.ID)
}

func (s *Service) Login(ctx context.Context, email, password string) (*TokenPair, error) {
	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}
	if !CheckPassword(user.PasswordHash, password) {
		return nil, ErrInvalidCredentials
	}
	return s.tokens.IssueTokenPair(user.ID)
}

func (s *Service) Refresh(ctx context.Context, refreshToken string) (string, error) {
	claims, err := s.tokens.Parse(refreshToken, TokenTypeRefresh)
	if err != nil {
		return "", ErrInvalidToken
	}
	if _, err := s.repo.FindByID(ctx, claims.UserID); err != nil {
		return "", ErrInvalidToken
	}
	return s.tokens.IssueAccessToken(claims.UserID)
}
