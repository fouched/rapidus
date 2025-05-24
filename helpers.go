package rapidus

import (
	"crypto/rand"
	"os"
)

const (
	randomString = "abcdefghijklmnopqrstuvwzyzABCDEFGHIKLMNOPQRSTUVWXYZ_+1234567890"
)

// RandomString generates a random string of length n
func (r *Rapidus) RandomString(n int) string {

	s, t := make([]rune, n), []rune(randomString)

	for i := range s {
		p, _ := rand.Prime(rand.Reader, len(t))
		x, y := p.Uint64(), uint64(len(t))
		s[i] = t[x%y]
	}
	return string(s)
}

func (r *Rapidus) CreateDirIfNotExist(path string) error {
	const mode = 0755
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := os.Mkdir(path, mode)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *Rapidus) CreateFileIfNotExist(path string) error {
	// could also inline below, just showing the normal way
	var _, err = os.Stat(path)
	if os.IsNotExist(err) {
		var file, err = os.Create(path)
		if err != nil {
			return err
		}

		defer func(file *os.File) {
			_ = file.Close()
		}(file)
	}

	return nil
}
