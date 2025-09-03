package util

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

/*
func GenerateTokens(studentID string) (string, string, error) {
	// Access Token (10분)
	accessClaims := jwt.MapClaims{
		"student_id": studentID,
		"exp":        time.Now().Add(10 * time.Minute).Unix(),
		"iss":        os.Getenv("JWT_ISS"),
		"aud":        os.Getenv("JWT_AUD"),
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(os.Getenv("JWT_ACCESS_SECRET")))
	if err != nil {
		return "", "", err
	}

	// Refresh Token (2주)
	refreshClaims := jwt.MapClaims{
		"student_id": studentID,
		"exp":        time.Now().Add(336 * time.Hour).Unix(),
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(os.Getenv("JWT_ACCESS_SECRET")))
	if err != nil {
		return "", "", err
	}

	return accessTokenString, refreshTokenString, nil
}
*/

// IssueAccessToken: 학번(studentID)을 sub로 하는 HS256 JWS 발급
// - iss/aud/iat/exp 등 표준 클레임을 채워 넣는다.
// - 운영에서 비대칭(EdDSA)로 바꾸면 공개키 배포/JWKS 도입이 용이.
func IssueAccessToken(studentID string) (string, error) {
	secret := []byte(os.Getenv("JWT_ACCESS_SECRET")) // 절대 유출 금지
	iss := os.Getenv("JWT_ISS")                      // 발급자
	aud := os.Getenv("JWT_AUD")                      // 대상
	ttlMin := EnvInt("JWT_ACCESS_TTL_MIN", 10)       // 만료(분)

	now := time.Now()

	// JWT payload(클레임)
	claims := jwt.MapClaims{
		"sub": studentID,                                           // 누가(학번)
		"iss": iss,                                                 // 누가 발급
		"aud": aud,                                                 // 누구에게 유효
		"iat": now.Unix(),                                          // 발급 시각
		"exp": now.Add(time.Duration(ttlMin) * time.Minute).Unix(), // 만료 시각
	}

	// 헤더의 alg는 HS256, typ는 JWT
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tok.Header["typ"] = "JWT"
	tok.Header["kid"] = "hs256-main" // 키 식별자(로테이션 대비, 지금은 고정 값)

	// 서명 후 compact 토큰 문자열 반환
	return tok.SignedString(secret)
}

// EnvInt: 환경변수를 정수로 읽는 작은 헬퍼 (비었거나 파싱 실패 시 def 반환)
func EnvInt(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	var x int
	if _, err := fmt.Sscan(v, &x); err != nil || x <= 0 {
		return def
	}
	return x
}

// RandomToken: 안전한 랜덤 토큰 생성 (auth.go에서 사용)
func RandomToken(length int) string {
	bytes := make([]byte, length)
	rand.Read(bytes)
	return base64.RawURLEncoding.EncodeToString(bytes)
}

/*
// VerifyAccessToken: Access 토큰 검증 (미들웨어에서 사용)
func VerifyAccessToken(tokenStr string) (jwt.MapClaims, error) {
	secret := []byte(os.Getenv("JWT_ACCESS_SECRET"))
	iss := os.Getenv("JWT_ISS")
	aud := os.Getenv("JWT_AUD")

	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
		if t.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, jwt.ErrSignatureInvalid
		}
		return secret, nil
	}, jwt.WithIssuer(iss), jwt.WithAudience(aud))

	if err != nil || !token.Valid {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, jwt.ErrInvalidKey
	}

	return claims, nil
}
*/
