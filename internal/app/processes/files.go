package processes

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"path/filepath"
)

const (
	pathTemp  = "/tmp"
	PathCache = "/cache"
)

func ReadFile(filename string) ([]byte, error) {
	filePath := filepath.Join(".", PathCache, filename)

	// Читаем содержимое файла
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func WriteFile(fileContent []byte, fileName string) (string, error) {
	path := filepath.Join(".", PathCache)
	// Создание директории, если её нет.
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return "", errors.New("failed to create cache directory")
	}

	// Создание пути для сохранения файла в директории PathCache
	savePath := filepath.Join(path, fileName)

	err := os.WriteFile(savePath, fileContent, 0644)
	if err != nil {
		return "", errors.New("failed to save file content")
	}

	return savePath, nil
}

func CalculateChecksum(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()

	if _, err = io.Copy(hash, file); err != nil {
		return "", err
	}

	hashInBytes := hash.Sum(nil)
	checksum := hex.EncodeToString(hashInBytes)

	return checksum, nil
}
