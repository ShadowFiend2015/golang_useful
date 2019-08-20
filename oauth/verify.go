package oauth

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"net/http"
	"time"
	"ucenter/defines"
	"ucenter/model"
)

func isValidToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
		return userTokenSec, nil
	})

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims["userid"].(string), nil
	} else {
		return "", err
	}
}

func (h *Handler) generateJWT(r *http.Request) (token string, err error) {
	if r.Form == nil {
		r.ParseForm()
	}
	var userid string
	username := r.FormValue("username")
	password := r.FormValue("password")
	userid, err = h.verifyUser(username, password)
	if err != nil {
		return
	}
	jwttoken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userid": userid,
		"exp":    time.Now().Add(expDura).Unix(),
	})

	// Sign and get the complete encoded token as a string using the secret
	return jwttoken.SignedString(userTokenSec)
}

func (h *Handler) verifyUser(username, passwd string) (userID string, err error) {
	//TODO: verify username & password
	user, err := model.VerifyUser(username, passwd, defines.ACCOUNT)
	if err != nil {
		return "", err
	}
	return user.ID.Hex(), nil
}
