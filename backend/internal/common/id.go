package common

import (
    "crypto/rand"
    "encoding/hex"
    "errors"
)

var ErrNotFound = errors.New("not found")

func NewID() (string, error) {
    buf := make([]byte, 16)
    if _, err := rand.Read(buf); err != nil {
        return "", err
    }
    return hex.EncodeToString(buf), nil
}
