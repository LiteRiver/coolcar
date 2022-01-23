package service

import (
	"crypto/rsa"
	"io/ioutil"
	"os"

	"github.com/dgrijalva/jwt-go/v4"
	"go.uber.org/zap"
)

type FilePrivateKeyProvider struct {
	Logger *zap.Logger
}

func (p *FilePrivateKeyProvider) GetPrivateKey(logger *zap.Logger) *rsa.PrivateKey {
	f, err := os.Open("auth/private.key")
	if err != nil {
		p.Logger.Fatal("can not read private key", zap.Error(err))
	}

	pkBytes, err := ioutil.ReadAll(f)
	if err != nil {
		p.Logger.Fatal("can not read private key", zap.Error(err))
	}

	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(pkBytes)
	if err != nil {
		p.Logger.Fatal("can not parse private key", zap.Error(err))
	}

	return privateKey
}
