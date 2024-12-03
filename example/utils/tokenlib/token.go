package tokenlib

import (
	"github.com/golang-jwt/jwt/v5"
	"time"
)

const TokenSecret = "idlerpggame"

type IdleRpgClaims struct {
	jwt.RegisteredClaims
	UserID     int32 `json:"user_id"`
	ExpireTime int64 `json:"expire_time"`
}

func createTokenExpireTime() time.Duration {
	return time.Hour * 24 * 15 // 15天过期
}

// CreateJwtToken 生成一个jwt token
func CreateJwtToken(userID int32) (string, error) {
	jtc := jwt.NewWithClaims(jwt.SigningMethodHS256, IdleRpgClaims{
		UserID:     userID,
		ExpireTime: int64(createTokenExpireTime()),
	})

	return jtc.SignedString([]byte(TokenSecret))
}

func CheckJwtToken(token string) bool {
	if _, err := jwt.ParseWithClaims(token, &IdleRpgClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(TokenSecret), nil
	}); err != nil {
		return false
	}
	return true
}

func ParseJwtToken(token string) (*IdleRpgClaims, error) {
	claims := &IdleRpgClaims{}
	_, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(TokenSecret), nil
	})
	return claims, err
}
