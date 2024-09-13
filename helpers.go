package gudu

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"math/big"
	"os"
	"regexp"
	"runtime"
	"time"
)

// GenerateRandomString generates a random string of n characters
func (g *Gudu) GenerateRandomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	// hold the generated random characters.
	result := make([]byte, n)

	// The loop iterates n times to fill the result slice with random characters
	for i := 0; i < n; i++ {
		// This generates a cryptographically secure random number between 0 and
		// the length of letters (which is 62, since there are 62 possible characters).
		// The rand.Reader is a secure random number generator.
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return ""
		}
		result[i] = letters[num.Int64()]
	}
	return string(result)
}

// CreateDirIfNotExists checks if a directory exists at dirPath, and
// creates it with the specified mode if it doesn't exist.
func (g *Gudu) CreateDirIfNotExists(dirPath string) error {
	// Define directory permissions mode
	const dirMode = 0755

	// Check if the directory exists
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		// Directory does not exist, create it
		if err := os.Mkdir(dirPath, dirMode); err != nil {
			return err
		}
	} else if err != nil {
		// Return error if Stat fails for any other reason
		return err
	}
	return nil
}

// CreateFileIfNotExist checks if a file exists at filePath,
// and creates it if it doesn't exist.
func (g *Gudu) CreateFileIfNotExist(filePath string) error {
	// Check if the file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// File does not exist, create it
		file, err := os.Create(filePath)
		if err != nil {
			return err
		}
		// Ensure file is closed properly
		defer func(file *os.File) {
			_ = file.Close()
		}(file)

	} else if err != nil {
		// Return error if Stat fails for any other reason
		return err
	}
	return nil
}

func (g *Gudu) LoadTime(start time.Time) {
	elapsed := time.Since(start)
	pc, _, _, _ := runtime.Caller(1)
	funcObj := runtime.FuncForPC(pc)
	runtimeFunc := regexp.MustCompile(`^.*\.(.*)$`)

	name := runtimeFunc.ReplaceAllString(funcObj.Name(), "$1")

	g.InfoLog.Println(fmt.Sprintf("Load time: %s took %s", name, elapsed))

}

type Encryption struct {
	Key []byte
}

// Encrypt function encrypts the plaintext using AES and returns the
// ciphertext as a base64 encoded string
func (e *Encryption) Encrypt(text string) (string, error) {

	plainText := []byte(text)

	// Create a new AES cipher with the provided key
	block, err := aes.NewCipher(e.Key)
	if err != nil {
		return "", err // Return an error if cipher creation fails
	}

	// Create a byte slice for the ciphertext, which is the size of the AES block plus the length of the plaintext
	ciphertext := make([]byte, aes.BlockSize+len(plainText))

	// Create an initialization vector (IV) from the first part of the ciphertext
	iv := ciphertext[:aes.BlockSize]

	// Fill the IV with random bytes
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err // Return an error if IV generation fails
	}

	// Create a new CFB encrypter with the cipher block and IV
	stream := cipher.NewCFBEncrypter(block, iv)

	// Encrypt the plaintext by XORing it with the key stream
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plainText)

	// Return the ciphertext as a base64 encoded string
	return base64.URLEncoding.EncodeToString(ciphertext), nil
}

// Decrypt function decrypts the base64 encoded ciphertext using AES
// and returns the plaintext
func (e *Encryption) Decrypt(ciphertext string) (string, error) {
	// Decode the base64 encoded ciphertext
	ciphertextBytes, err := base64.URLEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err // Return an error if decoding fails
	}
	// Create a new AES cipher with the provided key
	block, err := aes.NewCipher(e.Key)
	if err != nil {
		return "", err // Return an error if cipher creation fails
	}

	// Check if the ciphertext is at least as long as the AES block size
	if len(ciphertextBytes) < aes.BlockSize {
		return "", errors.New("ciphertext too short") // Return an error if the ciphertext is too short
	}

	// Extract the initialization vector (IV) from the ciphertext
	iv := ciphertextBytes[:aes.BlockSize]

	// Extract the actual ciphertext
	ciphertextBytes = ciphertextBytes[aes.BlockSize:]

	// Create a new CFB decrypter with the cipher block and IV
	stream := cipher.NewCFBDecrypter(block, iv)

	// Decrypt the ciphertext by XORing it with the key stream
	stream.XORKeyStream(ciphertextBytes, ciphertextBytes)

	// Return the decrypted plaintext
	return string(ciphertextBytes), nil
}
