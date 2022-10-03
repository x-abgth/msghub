package jwt

import (
	"log"
	"msghub-server/models"
	"net/http"
	"os"

	"github.com/golang-jwt/jwt/v4"
	"github.com/joho/godotenv"
)

type UserJwtClaim struct {
	User            models.UserModel
	IsAuthenticated bool
	jwt.RegisteredClaims
}

var JwtKey []byte

func InitJwtKey() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal(".env file loading error -- ", err)
		os.Exit(0)
	}

	key := os.Getenv("JWT_KEY")

	JwtKey = []byte(key)
}

func SignJwtToken(u *UserJwtClaim) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, u)
	tokenString, err := token.SignedString(JwtKey)
	if err != nil {
		panic("Internal server error!")
	}
	return tokenString
}

func GetValueFromJwt(c *http.Cookie) *UserJwtClaim {
	tknStr := c.Value

	// Initialize a new instance of `Claims`
	claims := &UserJwtClaim{}

	// Parse the JWT string and store the result in `claims`.
	// Note that we are passing the key in this method as well. This method will return an error
	// if the token is invalid (if it has expired according to the expiry time we set on sign in),
	// or if the signature does not match
	tkn, err := jwt.ParseWithClaims(tknStr, claims, func(token *jwt.Token) (interface{}, error) {
		return JwtKey, nil
	})
	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			panic("Invalid signature")
		}
		panic("Bad request")
	}
	if !tkn.Valid {
		panic("Unauthorized")
	}
	return claims
}
