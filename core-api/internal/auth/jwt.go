package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var ErrInvalidToken = errors.New("token inválido")

type TokenType string

const (
	TokenTypeAccess  TokenType = "access"
	TokenTypeRefresh TokenType = "refresh"
)

type Claims struct {
	UserID uuid.UUID `json:"user_id"`
	Type   TokenType `json:"type"`
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type JWTIssuer struct {
	secret          []byte
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

func NewJWTIssuer(secret string, accessTTL, refreshTTL time.Duration) *JWTIssuer {
	return &JWTIssuer{secret: []byte(secret), accessTokenTTL: accessTTL, refreshTokenTTL: refreshTTL}
}

func (j *JWTIssuer) IssueTokenPair(userID uuid.UUID) (*TokenPair, error) {
	access, err := j.IssueAccessToken(userID)
	if err != nil {
		return nil, err
	}
	refresh, err := j.issue(userID, TokenTypeRefresh, j.refreshTokenTTL)
	if err != nil {
		return nil, err
	}
	return &TokenPair{AccessToken: access, RefreshToken: refresh}, nil
}

func (j *JWTIssuer) IssueAccessToken(userID uuid.UUID) (string, error) {
	return j.issue(userID, TokenTypeAccess, j.accessTokenTTL)
}

func (j *JWTIssuer) issue(userID uuid.UUID, tokenType TokenType, ttl time.Duration) (string, error) {
	claims := Claims{
		UserID: userID,
		Type:   tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secret)
}

func (j *JWTIssuer) Parse(tokenString string, expectedType TokenType) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return j.secret, nil
	})
	if err != nil || !token.Valid {
		return nil, ErrInvalidToken
	}
	if claims.Type != expectedType {
		return nil, ErrInvalidToken
	}
	return claims, nil
}
