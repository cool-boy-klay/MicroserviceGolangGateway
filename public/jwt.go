package public

import (
	"errors"
	"github.com/dgrijalva/jwt-go"
	"time"
)




func JwtDecode(tokenString string) (*jwt.StandardClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(JwtSignKey), nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*jwt.StandardClaims); ok {
		if claims.ExpiresAt < time.Now().Unix() {
			return nil, errors.New("request expired")
		}

		return claims ,nil
	} else {
		return nil, errors.New("token is not jtw.StandardClaims")
	}
}

func JwtEncode(claims jwt.StandardClaims) (string, error) {

	mySigningKey := []byte(JwtSignKey)
	//使用HS256、参数claims的方法组装token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	//使用token结构体进行编码
	return token.SignedString(mySigningKey)
}