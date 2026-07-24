package lib

import "crypto/rand"

type Session struct {
	Id []byte
}

// GenerateSessionId generates a random session ID. Returns the session ID or an error that occurred.
func GenerateSessionId() (*Session, error) {
	sessionId := make([]byte, 32)
	_, err := rand.Read(sessionId)
	if err != nil {
		return nil, err
	}
	return &Session{Id: sessionId}, nil
}
