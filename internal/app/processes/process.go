package processes

import (
	"errors"
	"os"
	"sort"

	"karma8/internal/models"
)

// SplitFile разбивает файл на несколько частей и возвращает массив элементов BucketItem.
func SplitFile(path string, partsIDs []int64) ([]models.BucketItem, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, err
	}

	partsCount := len(partsIDs)
	if partsCount == 0 {
		return nil, errors.New("bucket items are empty")
	}

	fileSize := fileInfo.Size()
	if fileSize == 0 {
		return nil, errors.New("file is empty")
	}

	partSize := fileSize / int64(partsCount)

	var bucketItems []models.BucketItem

	for i := 0; i < partsCount; i++ {
		partID := partsIDs[i]
		offset := int64(i) * partSize

		_, err := file.Seek(offset, 0)
		if err != nil {
			return nil, err
		}

		partData := make([]byte, partSize)
		n, err := file.Read(partData)
		if err != nil {
			return nil, err
		}

		// Создание элемента корзины.
		bucketItem := models.BucketItem{
			ID:     partID,
			Source: partData[:n],
		}

		bucketItems = append(bucketItems, bucketItem)
	}

	return bucketItems, nil
}

// MergeFile объединяет элементы BucketItem в один файл и возвращает его содержимое.
func MergeFile(bucketItems []models.BucketItem) []byte {
	// Сортируем элементы по порядку разбиения BucketItem.ID.
	sort.Slice(bucketItems, func(i, j int) bool {
		return bucketItems[i].ID < bucketItems[j].ID
	})

	var mergedData []byte

	for _, item := range bucketItems {
		mergedData = append(mergedData, item.Source...)
	}

	return mergedData
}
