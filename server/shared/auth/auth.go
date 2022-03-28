package auth

import (
	"context"
	"coolcar/shared/auth/token"
	"coolcar/shared/id"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/dgrijalva/jwt-go/v4"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func Interceptor(publicKeyPath string) (grpc.UnaryServerInterceptor, error) {
	f, err := os.Open(publicKeyPath)
	if err != nil {
		return nil, fmt.Errorf("cannot open public key file: %v", err)
	}

	keyBytes, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("cannot read public key file: %v", err)
	}

	key, err := jwt.ParseRSAPublicKeyFromPEM(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("cannot parse public key: %v", err)
	}

	i := &interceptor{
		verifier: &token.JWTTokenVerifier{
			PublicKey: key,
		},
	}
	return i.HandleReq, nil
}

type tokenVerifier interface {
	Verify(token string) (string, error)
}

type interceptor struct {
	verifier tokenVerifier
}

const (
	authorizationHeader = "authorization"
	bearerPrefex        = "Bearer "
)

func (i *interceptor) HandleReq(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	tkn, err := tokenFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "")
	}

	accountId, err := i.verifier.Verify(tkn)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
	}

	return handler(ContextWithAccountId(ctx, id.AccountId(accountId)), req)
}

func tokenFromContext(ctx context.Context) (string, error) {
	m, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "")
	}

	tkn := ""
	for _, v := range m[authorizationHeader] {
		if strings.HasPrefix(v, bearerPrefex) {
			tkn = v[len(bearerPrefex):]
		}
	}

	if tkn == "" {
		return "", status.Error(codes.Unauthenticated, "")
	}

	return tkn, nil
}

type accountIdKey struct{}


func ContextWithAccountId(ctx context.Context, accountId id.AccountId) context.Context {
	return context.WithValue(ctx, accountIdKey{}, accountId)
}

func AccountIdFromContext(ctx context.Context) (id.AccountId, error) {
	v := ctx.Value(accountIdKey{})
	accountId, ok := v.(id.AccountId)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "")
	}

	return accountId, nil
}
