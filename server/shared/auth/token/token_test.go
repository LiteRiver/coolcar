package token

import (
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go/v4"
)

const publicKey = `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAu1SU1LfVLPHCozMxH2Mo
4lgOEePzNm0tRgeLezV6ffAt0gunVTLw7onLRnrq0/IzW7yWR7QkrmBL7jTKEn5u
+qKhbwKfBstIs+bMY2Zkp18gnTxKLxoS2tFczGkPLPgizskuemMghRniWaoLcyeh
kd3qqGElvW/VDL5AaWTg0nLVkjRo9z+40RQzuVaE8AkAFmxZzow3x+VJYKdjykkJ
0iT9wCS0DRTXu269V264Vf/3jvredZiKRkgwlL9xNAwxXFg0x/XFw005UWVRIkdg
cKWTjpBP2dPwVZ4WWC+9aGVd+Gyn1o0CLelf4rEjGoXbAAEgAqeGUxrcIlbjXfbc
mwIDAQAB
-----END PUBLIC KEY-----`

func TestVerify(t *testing.T) {
	jwt.TimeFunc = func() time.Time {
		return time.Unix(1642942205, 0)
	}
	key, err := jwt.ParseRSAPublicKeyFromPEM([]byte(publicKey))
	if err != nil {
		t.Fatalf("cannot parse public key: %v", err)
	}
	v := JWTTokenVerifier{
		PublicKey: key,
	}

	cases := []struct {
		name    string
		tkn     string
		now     time.Time
		want    string
		wantErr bool
	}{
		{
			name:    "valid_token",
			tkn:     "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2NDI5NDk0MDUuNTExMzkxLCJpYXQiOjE2NDI5NDIyMDUuNTExMzkxLCJpc3MiOiJjb29sY2FyL2F1dGgiLCJzdWIiOiI2MWU2ZjkwYTYzZjFkMDA3ZjY3MWIwZjcifQ.OKH4ShSV7pmg5rukEXT0RvX0_hEatpN4xRtaKCa0yvOMUxs649aF3zxAVqlFd8deJkXsv-y9FyquCTxVTjmKzh4B_JQNF5yuLdIJEs6scdYSDy5877TcYEegCi9t2j9MwtOn1pvNEl7XNZ-Y4XrqE3OugHS7CTY96LNuXIwHMg6fFmeaYKrpILIM_03j6n0oyvH6_l8UsUtuSPaWbT4q88vU3XthL_iWX2xcSnuEXIw2xKVj3BKJLgquW96zXDORgANnlCohPQc1McQRcELNCPBHxCFNi17lT48lkMw_w-BHlQaBvBxuzz8aIrfYjsMLLfvJoP2VChEKEyaZtBaR2Q",
			now:     time.Unix(1642942205, 0),
			want:    "61e6f90a63f1d007f671b0f7",
			wantErr: false,
		},
		{
			name:    "expired_token",
			tkn:     "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2NDI5NDk0MDUuNTExMzkxLCJpYXQiOjE2NDI5NDIyMDUuNTExMzkxLCJpc3MiOiJjb29sY2FyL2F1dGgiLCJzdWIiOiI2MWU2ZjkwYTYzZjFkMDA3ZjY3MWIwZjcifQ.OKH4ShSV7pmg5rukEXT0RvX0_hEatpN4xRtaKCa0yvOMUxs649aF3zxAVqlFd8deJkXsv-y9FyquCTxVTjmKzh4B_JQNF5yuLdIJEs6scdYSDy5877TcYEegCi9t2j9MwtOn1pvNEl7XNZ-Y4XrqE3OugHS7CTY96LNuXIwHMg6fFmeaYKrpILIM_03j6n0oyvH6_l8UsUtuSPaWbT4q88vU3XthL_iWX2xcSnuEXIw2xKVj3BKJLgquW96zXDORgANnlCohPQc1McQRcELNCPBHxCFNi17lT48lkMw_w-BHlQaBvBxuzz8aIrfYjsMLLfvJoP2VChEKEyaZtBaR2Q",
			now:     time.Unix(1642949406, 0),
			want:    "",
			wantErr: true,
		},
		{
			name:    "bad_token",
			tkn:     "bad_token",
			now:     time.Unix(1642942205, 0),
			want:    "",
			wantErr: true,
		},
		{
			name:    "wrong_signature",
			tkn:     "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2NDI5NDk0MDUuNTExMzkxLCJpYXQiOjE2NDI5NDIyMDUuNTExMzkxLCJpc3MiOiJjb29sY2FyL2F1dGgiLCJzdWIiOiI2MWU2ZjkwYTYzZjFkMDA3ZjY3MWIwZjYifQ.OKH4ShSV7pmg5rukEXT0RvX0_hEatpN4xRtaKCa0yvOMUxs649aF3zxAVqlFd8deJkXsv-y9FyquCTxVTjmKzh4B_JQNF5yuLdIJEs6scdYSDy5877TcYEegCi9t2j9MwtOn1pvNEl7XNZ-Y4XrqE3OugHS7CTY96LNuXIwHMg6fFmeaYKrpILIM_03j6n0oyvH6_l8UsUtuSPaWbT4q88vU3XthL_iWX2xcSnuEXIw2xKVj3BKJLgquW96zXDORgANnlCohPQc1McQRcELNCPBHxCFNi17lT48lkMw_w-BHlQaBvBxuzz8aIrfYjsMLLfvJoP2VChEKEyaZtBaR2Q",
			now:     time.Unix(1642942205, 0),
			want:    "",
			wantErr: true,
		},
	}

	for _, cs := range cases {
		t.Run(cs.name, func(t *testing.T) {
			jwt.TimeFunc = func() time.Time {
				return cs.now
			}
			accountId, err := v.Verify(cs.tkn)
			if !cs.wantErr && err != nil {
				t.Errorf("verification failed: %v", err)
			}
			if cs.wantErr && err == nil {
				t.Errorf("want error, got not error")
			}

			if accountId != cs.want {
				t.Errorf("wrong accountId, want: %q, got: %q", cs.want, accountId)
			}
		})
	}
}
