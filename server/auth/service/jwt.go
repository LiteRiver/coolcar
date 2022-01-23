package service

import (
	"crypto/rsa"
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go/v4"
)

type TokenProvider struct {
	privateKey *rsa.PrivateKey
	issuer     string
	GetNow     func() time.Time
}

func CreateTokenProvider(issuer string, privateKey *rsa.PrivateKey) *TokenProvider {
	return &TokenProvider{
		issuer:     issuer,
		privateKey: privateKey,
		GetNow:     time.Now,
	}
}

func (t *TokenProvider) GenerateToken(accountId string, exipresIn time.Duration) (string, error) {
	now := jwt.Time{Time: t.GetNow()}
	expiresAt := jwt.Time{Time: now.Add(exipresIn)}

	token := jwt.NewWithClaims(
		jwt.SigningMethodRS512,
		jwt.StandardClaims{
			Issuer:    t.issuer,
			IssuedAt:  &now,
			ExpiresAt: &expiresAt,
			Subject:   accountId,
		})

	return token.SignedString(t.privateKey)
}
