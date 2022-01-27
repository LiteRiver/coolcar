package service

import (
	"crypto/rsa"
	"time"

	"github.com/dgrijalva/jwt-go/v4"
)

type TokenProvider struct {
	privateKeyProvider PrivateKeyProvider
	issuer             string
	GetNow             func() time.Time
}
type PrivateKeyProvider interface {
	GetPrivateKey() (*rsa.PrivateKey, error)
}

func CreateTokenProvider(issuer string, privateKeyProvider PrivateKeyProvider) *TokenProvider {
	return &TokenProvider{
		issuer:             issuer,
		privateKeyProvider: privateKeyProvider,
		GetNow:             time.Now,
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

	privateKey, err := t.privateKeyProvider.GetPrivateKey()
	if err != nil {
		return "", nil
	}

	return token.SignedString(privateKey)
}
