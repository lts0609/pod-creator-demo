package utils

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"golang.org/x/crypto/argon2"
	"math/big"
)

const (
	DefaultPasswordLength = 12
	Digits                = "0123456789"
	LowercaseChars        = "abcdefghijklmnopqrstuvwxyz"
	UppercaseChars        = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

// 哈希默认参数
const (
	memory      = 10240
	iterations  = 1
	parallelism = 8
	keyLength   = 32
)

func GenerateJupyterPassword() ([]byte, []byte, error) {
	password, err := GenerateRandomString(DefaultPasswordLength)
	if err != nil {
		return nil, nil, err
	}
	hashedPassword, err := HashPasswordWithSalt(password)
	if err != nil {
		return nil, nil, err
	}
	return password, hashedPassword, nil
}

func GenerateRandomDigits(length int) (string, error) {
	bytes := make([]byte, length)

	for i := 0; i < length; i++ {
		digit, _ := rand.Int(rand.Reader, big.NewInt(int64(len(Digits))))
		bytes[i] = Digits[digit.Int64()]
	}
	return string(bytes), nil
}

func GenerateRandomString(length int) ([]byte, error) {
	charset := Digits + LowercaseChars + UppercaseChars
	bytes := make([]byte, length)

	for i := 0; i < length; i++ {
		char, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return nil, err
		}
		bytes[i] = charset[char.Int64()]
	}

	return bytes, nil
}

func HashPasswordWithSalt(password []byte) ([]byte, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return nil, err
	}

	hash := argon2.IDKey(password, salt, iterations, memory, parallelism, keyLength)
	return []byte(fmt.Sprintf(
		"argon2:$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		memory,
		iterations,
		parallelism,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	)), nil
}
