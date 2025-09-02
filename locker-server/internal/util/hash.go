package util

import (
	"crypto/rand"
	"encoding/base64"
)

// RandomToken: n바이트 길이의 랜덤 바이트를 생성하여 base64url 문자열로 반환
// - refresh 토큰 평문으로 사용 (서버는 항상 해시만 저장)
func RandomToken(n int) string {
	b := make([]byte, n)
	// rand.Read는 암호학적으로 안전한 난수를 채운다. 오류는 무시(베스트 에포트).
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}
