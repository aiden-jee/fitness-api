package authstore

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"sync"
	"time"
	"unicode"
)

const PW_MIN_LENGTH = 6
const TOKEN_DURATION = time.Hour * 24
const TOKEN_BYTES = 32

type Session struct {
	UserId    int
	TokenHash [TOKEN_BYTES]byte
	ExpiresAt time.Time
}

type User struct {
	Id        int
	User      string `json:"user"`
	Pw        string `json:"pw"`
	LastLogin *time.Time
}

type AuthStore struct {
	sync.Mutex
	Users    map[string]User
	UserId   int
	Sessions map[[TOKEN_BYTES]byte]*Session
}

func New() *AuthStore {
	as := &AuthStore{}
	as.Users = make(map[string]User)
	as.Sessions = make(map[[TOKEN_BYTES]byte]*Session)
	as.UserId = 1
	return as
}

/* Helper Functions */
func GenerateToken() (token string, hashToken [TOKEN_BYTES]byte) {
	stream := make([]byte, TOKEN_BYTES)
	rand.Read(stream)

	// generate token from bytes
	token = base64.RawURLEncoding.EncodeToString(stream)

	// generate hash using sha256
	hashToken = sha256.Sum256(stream)

	return token, hashToken
}

/* POST Endpoints */
func (as *AuthStore) RegisterUser(user, pw string) error {
	// mutex lock
	as.Lock()
	defer as.Unlock()

	// check if user exists in store
	if _, ok := as.Users[user]; ok {
		return fmt.Errorf("%s already exists! Use a different username.", user)
	}

	// check if pw meets strength conditions
	if len(pw) < PW_MIN_LENGTH {
		return fmt.Errorf("Password is too short! Must be at least %v characters", PW_MIN_LENGTH)
	}
	hasUpper := false
	for _, c := range pw {
		if unicode.IsUpper(c) {
			hasUpper = true
			break
		}
	}
	if !hasUpper {
		return fmt.Errorf("Password must contain at least one uppercase letter!")

	}

	// create user
	u := User{
		Id:        as.UserId,
		User:      user,
		Pw:        pw,
		LastLogin: nil,
	}
	as.UserId += 1
	as.Users[user] = u
	return nil
}

func (as *AuthStore) LoginUser(user, pw string) (*Session, string, error) {
	// mutex lock
	as.Lock()
	defer as.Unlock()
	u, ok := as.Users[user]
	// check if user exists
	if !ok {
		return nil, "", fmt.Errorf("%s does not exist!", user)
	} else {
		if u.Pw != pw {
			return nil, "", fmt.Errorf("Incorrect password, try again.")
		}
	}

	// authenticate by creating new session
	token, hashToken := GenerateToken()
	expire := time.Now().Add(TOKEN_DURATION)
	s := &Session{
		UserId:    u.Id,
		TokenHash: hashToken,
		ExpiresAt: expire,
	}
	as.Sessions[hashToken] = s

	return s, token, nil
}

/* DELETE Endpoints */
func (as *AuthStore) LogOutUser(token [TOKEN_BYTES]byte) error {
	as.Lock()
	defer as.Unlock()

	if _, ok := as.Sessions[token]; !ok {
		return fmt.Errorf("Session not found!")
	} else {
		delete(as.Sessions, token)
	}

	return nil
}
