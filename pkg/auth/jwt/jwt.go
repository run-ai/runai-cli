package jwt

import (
	"github.com/dgrijalva/jwt-go"
	log "github.com/sirupsen/logrus"
	"time"
)

// Currently we only validate the token's expiry to conserve an api call to the identity provider, since we need to call it to retrieve the signing key
// see jwt.Parse for more details.
func IsTokenValid(rawToken string) bool {
	if rawToken == "" {
		return false
	}
	token, _, err := new(jwt.Parser).ParseUnverified(rawToken, jwt.StandardClaims{})
	if err != nil {
		log.Debug(err)
		return false
	}
	return token.Claims.(jwt.StandardClaims).ExpiresAt > time.Now().UnixNano()
}
