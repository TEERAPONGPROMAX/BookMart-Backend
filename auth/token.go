package auth

import (
	"errors"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	Role  string `json:"role"`
	Email string `json:"email"`
	jwt.RegisteredClaims
}

type Issuer struct{ Secret []byte }

func NewIssuer(secret string) *Issuer { return &Issuer{Secret: []byte(secret)} }

func (i *Issuer) Issue(id int64, email, role string) (string, error) {
	now := time.Now()
	claims := Claims{
		Role:  role,
		Email: email,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "bookmart-api",
			Subject:   strconv.FormatInt(id, 10),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(i.Secret)
}

func (i *Issuer) Parse(tokenString string) (int64, string, string, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		return i.Secret, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}), jwt.WithExpirationRequired(), jwt.WithIssuer("bookmart-api"))
	if err != nil || !token.Valid {
		return 0, "", "", errors.New("token invalid")
	}
	id, err := strconv.ParseInt(claims.Subject, 10, 64)
	if err != nil || id <= 0 || (claims.Role != "member" && claims.Role != "admin") {
		return 0, "", "", errors.New("claims invalid")
	}
	return id, claims.Role, claims.Email, nil
}
