//Package crypto contains helper cryptographic functions
package crypto

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	
	"golang.org/x/crypto/bcrypt"
)

func GenerateRandomString()(string,error){
	bytes := make ([]byte,32)
	_,err:= rand.Read(bytes);if err!=nil{
		return "",err
	}
	return hex.EncodeToString(bytes),nil
}

func HashSHA256(input string)string{
	hashedBytes := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hashedBytes[:])
}

func ValidateUserPassword(password string, hashedPassword string)error{
	if err:=bcrypt.CompareHashAndPassword([]byte(hashedPassword),[]byte(password));err!=nil{
		return err
	}
	return nil
}

func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}
