package dto

import (
	"github.com/golang-jwt/jwt/v4"
	"time"
)

type CustomClaims struct {
	Exp time.Duration `json:"exp"`
	jwt.RegisteredClaims
}
