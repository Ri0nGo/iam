package password

import "golang.org/x/crypto/bcrypt"

func Hash(raw string, cost int) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(raw), cost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func Verify(hash string, raw string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(raw)) == nil
}
