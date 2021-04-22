// Package user contains the basic admin user creation and authentication code,
// specific to eshop
package user

import (
	"bytes"
	crand "crypto/rand"
	"encoding/base64"
	mrand "math/rand"
	"net/http"

	"strings"
	"time"

	"github.com/agreyfox/eshop/system/logs"
	"github.com/nilslice/jwt"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// User defines a admin user in the system
type User struct {
	ID      int    `json:"id"`
	Email   string `json:"email"`
	Hash    string `json:"hash"`
	Salt    string
	Locale  string `json:"locale"`
	Perm    Permissions
	IsAdmin bool   `json:"isAdmin"`
	Phone   string `json:phone,omitempty`
	Social  string `json:"social,omitempty"`
	Meta    string `json:metadata,omitempty`
}

const (
	// use for cookie name
	Lqcmstoken string = "lqcms_token"
)

var (
	r      = mrand.New(mrand.NewSource(time.Now().Unix()))
	err    error
	logger *zap.SugaredLogger = logs.Log.Sugar()
)

// New creates a user template with normal user  function
func New(email, password string) (*User, error) {
	salt, err := randSalt()
	if err != nil {
		return nil, err
	}

	hash, err := hashPassword([]byte(password), salt)
	if err != nil {
		return nil, err
	}

	user := &User{
		Email:   email,
		Hash:    string(hash),
		Salt:    base64.StdEncoding.EncodeToString(salt),
		Perm:    CustomerPermission,
		IsAdmin: false, //add 2021/4/13
	}
	return user, nil
}

// New creates a customer user from web and ready to push to db
func NewCustomer(email, password string) (*User, error) {
	salt, err := randSalt()
	if err != nil {
		return nil, err
	}

	hash, err := hashPassword([]byte(password), salt)
	if err != nil {
		return nil, err
	}

	user := &User{
		Email:  email,
		Hash:   string(hash),
		Salt:   base64.StdEncoding.EncodeToString(salt),
		Perm:   CustomerPermission,
		Social: "",
		Meta:   "",
	}

	return user, nil
}

// New creates a customer user from web and ready to push to db
func NewCustomerWithSocial(email, password, social, value string) (*User, error) {
	salt, err := randSalt()
	if err != nil {
		return nil, err
	}

	hash, err := hashPassword([]byte(password), salt)
	if err != nil {
		return nil, err
	}

	user := &User{
		Email:   email,
		Hash:    string(hash),
		Salt:    base64.StdEncoding.EncodeToString(salt),
		Perm:    CustomerPermission,
		IsAdmin: false,
		Meta:    social,
		Social:  value, // 保存用户的social内容
	}

	return user, nil
}

// Auth is HTTP middleware to ensure the request has proper token credentials
func Auth(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		redir := req.URL.Scheme + req.URL.Host + "/admin/v1/login"
		//logger.Debugf("Auth with request is %v", req)
		if IsValid(req) {
			next.ServeHTTP(res, req)
		} else {
			logger.Debugf("no Auth with request is %v", req.URL)
			http.Redirect(res, req, redir, http.StatusFound)
		}
	})
}

// Auth is HTTP middleware to ensure the request has proper token credentials
func CustomerAuth(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		//redir := req.URL.Scheme + req.URL.Host + "/api/login"

		if IsValid(req) {
			next.ServeHTTP(res, req)
		} else {
			http.Redirect(res, req, "/", http.StatusForbidden)
		}
	})
}

// IsValid checks if the user request is authenticated
func IsValid(req *http.Request) bool {
	var token string
	tt := req.Header.Get(Lqcmstoken)
	if len(tt) > 0 {
		logger.Debug("get token from header:", tt)
		token = tt
		return jwt.Passes(token)
	}

	cookie, err := req.Cookie(Lqcmstoken)
	if err != nil {
		logger.Error("Cookie Token is not found!")
		logger.Debug(req)
	} else {
		logger.Debug("get token from cookei:", tt)
		token = cookie.Value
	}
	if len(token) == 0 {
		reqToken := req.Header.Get("Authorization")
		if len(reqToken) == 0 {
			logger.Error("Authorization Token is not found!")
			logger.Debug(req)
			return false
		} else {
			splitToken := strings.Split(reqToken, "Bearer ")
			if len(splitToken) >= 1 {
				token = splitToken[1]
			} else {
				logger.Error("Authorization Token is not present!")
				logger.Debug(req)
				return false
			}
		}
	}
	// check if token exists in cookie

	// validate it and allow or redirect request
	if len(token) > 0 {
		return jwt.Passes(token)
	}
	return false
}

// IsValid checks if the user request is authenticated
func IsValidAdmin(req *http.Request) bool {
	cookie, err := req.Cookie(Lqcmstoken)
	if err != nil {
		return false
	}
	logger.Debug("user cookie:", cookie)
	// validate it and allow or redirect request
	token := cookie.Value
	if jwt.Passes(token) {
		clienInfo := jwt.GetClaims(token)

		username := clienInfo["user"].(string)
		logger.Debug("User validation for admin:", username)

		return true
	} else {
		return false
	}
}

// IsUser checks for consistency in email/pass combination
func IsUser(usr *User, password string) bool {
	salt, err := base64.StdEncoding.DecodeString(usr.Salt)
	if err != nil {
		return false
	}

	err = checkPassword([]byte(usr.Hash), []byte(password), salt)
	if err != nil {
		logger.Error("Error checking password:", err)
		return false
	}

	return true
}

// IsUser checks for consistency in email/pass combination
func IsAdminUser(usr *User, password string) bool {
	salt, err := base64.StdEncoding.DecodeString(usr.Salt)
	if err != nil {
		return false
	}

	err = checkPassword([]byte(usr.Hash), []byte(password), salt)
	if err != nil {
		logger.Error("Error checking password:", err)
		return false
	}

	return true
}

// randSalt generates 16 * 8 bits of data for a random salt
func randSalt() ([]byte, error) {
	buf := make([]byte, 16)
	count := len(buf)
	n, err := crand.Read(buf)
	if err != nil {
		return nil, err
	}

	if n != count || err != nil {
		for count > 0 {
			count--
			buf[count] = byte(r.Int31n(256))
		}
	}

	return buf, nil
}

// saltPassword combines the salt and password provided
func saltPassword(password, salt []byte) ([]byte, error) {
	salted := &bytes.Buffer{}
	_, err := salted.Write(append(salt, password...))
	if err != nil {
		return nil, err
	}

	return salted.Bytes(), nil
}

// hashPassword encrypts the salted password using bcrypt
func hashPassword(password, salt []byte) ([]byte, error) {
	salted, err := saltPassword(password, salt)
	if err != nil {
		return nil, err
	}

	hash, err := bcrypt.GenerateFromPassword(salted, 10)
	if err != nil {
		return nil, err
	}

	return hash, nil
}

// checkPassword compares the hash with the salted password. A nil return means
// the password is correct, but an error could mean either the password is not
// correct, or the salt process failed - indicated in logs
func checkPassword(hash, password, salt []byte) error {
	salted, err := saltPassword(password, salt)
	if err != nil {
		return err
	}

	return bcrypt.CompareHashAndPassword(hash, salted)
}
