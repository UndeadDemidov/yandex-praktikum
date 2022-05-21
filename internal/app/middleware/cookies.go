package middleware

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"net/http"
	"strings"
)

const (
	UserIdCookie     = "YPUserID"
	ContextUserIdKey = "YPUserID"
)

var (
	ErrSignedCookieInvalidValueOrUnsigned = errors.New("invalid cookie value or it is unsigned")
	ErrSignedCookieInvalidSign            = errors.New("invalid sign")
	ErrSignedCookieSaltNotSetProperly     = errors.New("SaltStartIdx and SaltEndIdx must be set properly")
)

func UserCookie(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user, err := getUserID(w, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		ctx = context.WithValue(ctx, ContextUserIdKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(fn)
}

func getUserID(w http.ResponseWriter, r *http.Request) (userID string, err error) {
	cookie, err := NewUserIDSignedCookie()
	if err != nil {
		return "", err
	}
	// получить куку пользователя
	c, err := r.Cookie(UserIdCookie)
	if errors.Is(err, http.ErrNoCookie) {
		http.SetCookie(w, cookie.Cookie)
	} else {
		cookie.Cookie = c
		err = cookie.DetachSign()
		if errors.Is(err, ErrSignedCookieInvalidSign) {
			cookie, err = NewUserIDSignedCookie()
			if err != nil {
				return "", err
			}
			http.SetCookie(w, cookie.Cookie)
		} else if err != nil {
			return "", err
		}
	}
	return cookie.BaseValue, nil
}

// GetUserID возвращает сохраненный в контексте куку UserIdCookie
func GetUserID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if reqID, ok := ctx.Value(ContextUserIdKey).(string); ok {
		return reqID
	}
	return ""
}

type SignedCookie struct {
	*http.Cookie
	SaltStartIdx uint
	SaltEndIdx   uint
	key          []byte
	sign         []byte
	BaseValue    string
}

func NewUserIDSignedCookie() (sc SignedCookie, err error) {
	sc = SignedCookie{
		Cookie: &http.Cookie{
			Path:   "/",
			Name:   UserIdCookie,
			Value:  uuid.New().String(),
			MaxAge: 60 * 10,
			// MaxAge:     60*60*24*180, // За полгода планирую уложиться
		},
		SaltStartIdx: 4,
		SaltEndIdx:   9,
	}

	err = sc.AttachSign()
	if err != nil {
		return SignedCookie{}, err
	}
	return sc, nil
}

func (sc *SignedCookie) AttachSign() (err error) {
	sc.BaseValue = sc.Value
	if len(sc.key) == 0 {
		err = sc.RecalcKey()
		if err != nil {
			return err
		}
	}
	sc.sign = sc.calcSign()

	sc.Value = fmt.Sprintf("%s|%s", sc.Value, hex.EncodeToString(sc.sign))
	return nil
}

func (sc *SignedCookie) calcSign() []byte {
	h := hmac.New(sha256.New, sc.key)
	h.Write([]byte(sc.BaseValue))
	return h.Sum(nil)
}

func (sc *SignedCookie) RecalcKey() (err error) {
	if sc.SaltStartIdx == 0 || sc.SaltEndIdx == 0 ||
		sc.SaltEndIdx < sc.SaltStartIdx || sc.SaltEndIdx > uint(len(sc.BaseValue)) {
		return ErrSignedCookieSaltNotSetProperly
	}

	var secretKey = []byte("secret key")
	secretKey = append(secretKey, []byte(sc.BaseValue)[sc.SaltStartIdx:sc.SaltEndIdx]...)
	sc.key = secretKey

	return nil
}

func (sc *SignedCookie) DetachSign() (err error) {
	ss := strings.Split(sc.Value, "|")
	if len(ss) < 2 {
		return ErrSignedCookieInvalidValueOrUnsigned
	}
	sc.BaseValue = ss[0]
	err = sc.RecalcKey()
	if err != nil {
		return err
	}

	sgn := ss[1]
	if hex.EncodeToString(sc.calcSign()) != sgn {
		return ErrSignedCookieInvalidSign
	}

	return nil
}
