package tools

import (
	"github.com/golang-jwt/jwt/v5"
	"sportgether/models"
	"strconv"
	"time"
)

func GenerateJwtToken(user *models.User) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)

	claims["sub"] = strconv.FormatInt(user.ID, 10)
	claims["iat"] = time.Now()
	claims["nbf"] = time.Now()
	claims["iss"] = "http://charmflex-98.net"
	claims["aud"] = "http://charmflex-98.net"
	claims["exp"] = jwt.NewNumericDate(time.Now().Add(720 * time.Hour))

	claims.GetAudience()

	tokenString, err := token.SignedString([]byte("This should be hide though. Just a sample now"))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
