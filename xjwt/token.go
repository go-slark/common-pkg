package xjwt

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
	"strings"
)

func GenerateToken(signKey interface{}, claims jwt.Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	if token == nil {
		return "", errors.New("generate token fail")
	}

	return token.SignedString(signKey)
}

func VerifyToken(tokenStr string, claims jwt.Claims, signKey string) (jwt.Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(signKey), nil
	})

	if token == nil {
		return nil, errors.WithStack(err)
	}

	tokenRaw := strings.FieldsFunc(token.Raw, func(r rune) bool {
		if r == '.' {
			return true
		}
		return false
	})

	if len(tokenRaw) != 3 {
		return nil, errors.New("token string invalid")
	}

	signature, err := jwt.SigningMethodHS256.Sign(tokenRaw[0] + "." + tokenRaw[1], []byte(signKey))
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if signature != token.Signature {
		return nil, errors.New("signature had been changed")
	}

	return token.Claims, nil
}