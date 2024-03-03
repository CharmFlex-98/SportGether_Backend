package tools

import (
	"errors"
	"slices"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	BadTokenSignatureError = errors.New("bad token")
)

const (
	JWT_SECRET_KEY       = "This should be hide though. Just a sample now"
	AUTHENTICATION_SCOPE = "Authentication scope"
)

func GenerateJwtToken(userId int64, username string, expiryInHour int, scope string) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)

	claims["sub"] = strconv.FormatInt(userId, 10)
	claims["usrnm"] = username
	claims["iat"] = jwt.NewNumericDate(time.Now())
	claims["nbf"] = jwt.NewNumericDate(time.Now())
	claims["iss"] = "http://charmflex-98.net"
	claims["aud"] = "http://charmflex-98.net"
	claims["exp"] = jwt.NewNumericDate(time.Now().Add(time.Duration(expiryInHour) * time.Hour))
	claims["scope"] = scope

	tokenString, err := token.SignedString([]byte(JWT_SECRET_KEY))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func ParseJwtToken(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, keyFuncCallback)
}

func keyFuncCallback(token *jwt.Token) (interface{}, error) {
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, BadTokenSignatureError
	}

	return []byte(JWT_SECRET_KEY), nil
}

func IsValidClaims(claim jwt.MapClaims) (*int64, bool) {
	iss, err := claim.GetIssuer()
	if err != nil || iss != "http://charmflex-98.net" {
		return nil, false
	}

	aud, err := claim.GetAudience()
	if err != nil || !slices.Contains(aud, "http://charmflex-98.net") {
		return nil, false
	}

	exp, err := claim.GetExpirationTime()
	if err != nil || exp.Before(time.Now()) {
		return nil, false
	}

	userId, err := claim.GetSubject()
	if err != nil {
		return nil, false
	}

	res, err := strconv.ParseInt(userId, 10, 64)
	if err != nil {
		return nil, false
	}

	return &res, true
}
