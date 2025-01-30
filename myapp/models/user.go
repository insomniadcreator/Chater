package models

type User struct {
	ID               int
	Username         string
	Email            string
	PasswordHash     string
	EmailVerified    bool
	VerificationCode string
}
