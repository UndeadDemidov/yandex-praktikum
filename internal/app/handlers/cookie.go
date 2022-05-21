package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"net/http"
	"strings"
)

const USER_ID_COOKIE = "YPUserIDTest"

var (
	ErrSignedCookieInvalidValueOrUnsigned = errors.New("invalid cookie value or it is unsigned")
	ErrSignedCookieInvalidSign            = errors.New("invalid sign")
	ErrSignedCookieSaltNotSetProperly     = errors.New("SaltStartIdx and SaltEndIdx must be set properly")
)

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
			Name:   USER_ID_COOKIE,
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

//func extractUser(s string) (userID string, err error) {
//	ss := strings.Split(s, "|")
//	if len(ss) < 2 {
//		return "", errors.New("invalid cookie value")
//	}
//	userID = ss[0]
//	// ToDo может быть пусто
//	sign := ss[1]
//	if !validateSign(userID, sign) {
//		return "", errors.New("invalid sign")
//	}
//	return userID, nil
//}

//func validateSign(id string, s string) bool {
//	return sign(id) == s
//}
//
//func setNewUserCookie(w http.ResponseWriter, cookieName string) (userID string) {
//	userID = uuid.New().String()
//	signedUserID := getSignedUserID(userID)
//	cookie := &http.Cookie{
//		Name:   cookieName,
//		Value:  signedUserID,
//		MaxAge: 60 * 5,
//		// MaxAge:     60*60*24*180, // За полгода планирую уложиться
//	}
//	http.SetCookie(w, cookie)
//	return
//}
//
//func getSignedUserID(ID string) string {
//	return fmt.Sprintf("%s|%s", ID, sign(ID))
//}
//
//func sign(s string) string {
//	key := calcKey(s)
//	h := hmac.New(sha256.New, key)
//	h.Write([]byte(s))
//	return hex.EncodeToString(h.Sum(nil))
//}
//
//func calcKey(s string) []byte {
//	var secretKey = []byte("secret key")
//	return append(secretKey, []byte(s)[4:9]...)
//}
