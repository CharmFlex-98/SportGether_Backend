package utils

import (
	"encoding/json"
	"io"
	"os"
)

func ReadJsonFromFile(filePath string, item interface{}) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Read the file content
	content, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	err = json.Unmarshal(content, &item)
	if err != nil {
		return err
	}

	return nil
}
