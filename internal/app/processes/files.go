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
	pathCache = "/cache"
)

// GetFileNameWithPathCache - возвращает путь к файлу в директории pathCache.
func GetFileNameWithPathCache(filename string) string {
	return filepath.Join(".", pathCache, filename)
}

// ReadFile - читает содержимое файла.
func ReadFile(filename string) ([]byte, error) {
	filePath := GetFileNameWithPathCache(filename)

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// WriteFile - записывает содержимое файла, результат - путь к файлу.
func WriteFile(fileContent []byte, fileName string) (string, error) {
	path := filepath.Join(".", pathCache)
	// Создание директории, если её нет.
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return "", errors.New("failed to create cache directory")
	}

	// Создание пути для сохранения файла в директории pathCache
	savePath := filepath.Join(path, fileName)

	err := os.WriteFile(savePath, fileContent, os.ModePerm)
	if err != nil {
		return "", errors.New("failed to save file content: " + err.Error())
	}

	return savePath, nil
}

// CalculateChecksum - вычисляет контрольную сумму файла.
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

// DeleteFile - удаляет файл.
func DeleteFile(filename string) error {
	filePath := GetFileNameWithPathCache(filename)

	os.Remove(filePath)

	return nil
}
