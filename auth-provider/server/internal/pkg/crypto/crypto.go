//Package crypto contains helper cryptographic functions
package crypto

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
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

