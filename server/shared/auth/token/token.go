package token

import (
	"crypto/rsa"
	"fmt"

	"github.com/dgrijalva/jwt-go/v4"
)

type JWTTokenVerifier struct {
	PublicKey *rsa.PublicKey
}

func (v *JWTTokenVerifier) Verify(token string) (string, error) {
	tkn, err := jwt.ParseWithClaims(token, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
		return v.PublicKey, nil
	})
	if err != nil {
		return "", fmt.Errorf("cannot parse token: %v", err)
	}

	if !tkn.Valid {
		return "", fmt.Errorf("invlaid token")
	}

	clm, ok := tkn.Claims.(*jwt.StandardClaims)
	if !ok {
		return "", fmt.Errorf("token claim is not StandardClaims")
	}

	if err := clm.Valid(nil); err != nil {
		return "", fmt.Errorf("invalid claim: %v", err)
	}

	return clm.Subject, nil
}
