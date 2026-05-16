package storage

import (
	"encoding/base64"
	"errors"
	"os"
	"strings"

	"github.com/google/uuid"
)

func Upload(file string) (string, error) {
	parts := strings.SplitN(file, ",", 2)
	if len(parts) != 2 {
		return "", errors.New("invalid base64 format")
	}

	meta := parts[0]
	extParts := strings.Split(strings.Split(meta, "/")[1], ";")
	ext := extParts[0]

	decodedFile, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return "", err
	}

	fileName := uuid.New().String() + "." + ext
	filePath := "uploads/avatars/" + fileName

	err = os.MkdirAll("uploads/avatars", os.ModePerm)
	err = os.WriteFile(filePath, decodedFile, 0666)
	if err != nil {
		return "", err
	}

	return "/uploads/avatars/" + fileName, nil
}
