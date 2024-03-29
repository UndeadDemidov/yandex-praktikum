package middleware

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

const UserIDCookie = "YPUserID"

var (
	ErrSignedCookieInvalidValueOrUnsigned = errors.New("invalid cookie value or it is unsigned")
	ErrSignedCookieInvalidSign            = errors.New("invalid sign")
	ErrSignedCookieSaltNotSetProperly     = errors.New("SaltStartIdx and SaltEndIdx must be set properly")
	ContextUserIDKey                      = LocalContext("YPUserID")
)

type LocalContext string

func UserCookie(next http.Handler) http.Handler {
	middleware := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user, err := getUserID(w, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		ctx = context.WithValue(ctx, ContextUserIDKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(middleware)
}

func getUserID(w http.ResponseWriter, r *http.Request) (userID string, err error) {
	cookie := NewUserIDSignedCookie()
	// получить куку пользователя
	c, err := r.Cookie(UserIDCookie)
	// куки нет
	if errors.Is(err, http.ErrNoCookie) {
		http.SetCookie(w, cookie.Cookie)
		return cookie.BaseValue, nil
	}
	if err != nil {
		return "", err
	}
	// кука есть
	cookie.Cookie = c
	err = cookie.DetachSign()
	switch {
	case err == nil: // кука подписана верно
		return cookie.BaseValue, nil
	case errors.Is(err, ErrSignedCookieInvalidSign): // кука подписана неверно
		cookie = NewUserIDSignedCookie()
		http.SetCookie(w, cookie.Cookie)
		return cookie.BaseValue, nil
	}
	return "", err
}

// GetUserID возвращает сохраненный в контексте куку UserIDCookie
func GetUserID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if reqID, ok := ctx.Value(ContextUserIDKey).(string); ok {
		return reqID
	}
	return ""
}

type SignedCookie struct {
	*http.Cookie
	BaseValue    string
	key          []byte
	sign         []byte
	SaltStartIdx uint
	SaltEndIdx   uint
}

func NewUserIDSignedCookie() (sc SignedCookie) {
	sc = SignedCookie{
		Cookie: &http.Cookie{
			Path:   "/",
			Name:   UserIDCookie,
			Value:  uuid.New().String(),
			MaxAge: 60 * 10,
			// MaxAge:     60*60*24*180, // За полгода планирую уложиться
		},
		SaltStartIdx: 4,
		SaltEndIdx:   9,
	}

	sc.AttachSign()
	return sc
}

func (sc *SignedCookie) AttachSign() {
	sc.BaseValue = sc.Value
	if len(sc.key) == 0 {
		sc.RecalcKey()
	}
	sc.sign = sc.calcSign()
	sc.Value = fmt.Sprintf("%s|%s", sc.Value, hex.EncodeToString(sc.sign))
}

func (sc *SignedCookie) calcSign() []byte {
	h := hmac.New(sha256.New, sc.key)
	h.Write([]byte(sc.BaseValue))
	return h.Sum(nil)
}

func (sc *SignedCookie) RecalcKey() {
	if sc.SaltStartIdx == 0 || sc.SaltEndIdx == 0 ||
		sc.SaltEndIdx < sc.SaltStartIdx || sc.SaltEndIdx > uint(len(sc.BaseValue)) {
		log.Panic().Msg(ErrSignedCookieSaltNotSetProperly.Error())
	}

	var secretKey = []byte("secret key")
	secretKey = append(secretKey, []byte(sc.BaseValue)[sc.SaltStartIdx:sc.SaltEndIdx]...)
	sc.key = secretKey
}

func (sc *SignedCookie) DetachSign() (err error) {
	ss := strings.Split(sc.Value, "|")
	if len(ss) < 2 {
		return ErrSignedCookieInvalidValueOrUnsigned
	}
	sc.BaseValue = ss[0]
	sc.RecalcKey()

	sgn := ss[1]
	if hex.EncodeToString(sc.calcSign()) != sgn {
		return ErrSignedCookieInvalidSign
	}

	return nil
}
