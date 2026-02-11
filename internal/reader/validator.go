package reader

import (
	"errors"
	"os"
)

func ValidateFile(filename string) error {
	info, err := os.Stat(filename)

	if os.IsNotExist(err) {
		return errors.New("arquivo não encontrado " + filename)
	}
	if info.IsDir() {
		return errors.New("caminho é um diretório, não um arquivo: " + filename)
	}

	return nil
}
