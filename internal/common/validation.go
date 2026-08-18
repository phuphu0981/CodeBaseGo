package common

import "net/mail"

// IsValidEmail checks whether an email address format is valid.
func IsValidEmail(email string) bool {
	addr, err := mail.ParseAddress(email)
	return err == nil && addr.Address == email
}
