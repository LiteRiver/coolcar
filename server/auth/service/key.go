package service

import (
	"crypto/rsa"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/dgrijalva/jwt-go/v4"
)

type FilePrivateKeyProvider struct{}

func (p *FilePrivateKeyProvider) GetPrivateKey() (*rsa.PrivateKey, error) {
	f, err := os.Open("auth/private.key")
	if err != nil {
		return nil, fmt.Errorf("cannot open private key file: %v", err)
	}

	pkBytes, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("cannot read private key: %v", err)
	}

	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(pkBytes)
	if err != nil {
		return nil, fmt.Errorf("cannot parse private key: %v", err)
	}

	return privateKey, nil
}
