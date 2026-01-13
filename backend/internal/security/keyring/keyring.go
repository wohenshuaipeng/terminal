package keyring

import (
	"errors"

	"github.com/zalando/go-keyring"
)

const serviceName = "goterm"

var ErrNotFound = keyring.ErrNotFound

func passwordKey(profileID string) string {
	return "cred:" + profileID
}

func passphraseKey(profileID string) string {
	return "keypass:" + profileID
}

func mysqlPasswordKey(profileID string) string {
	return "mysql:" + profileID
}

func SetPassword(profileID, password string) error {
	return keyring.Set(serviceName, passwordKey(profileID), password)
}

func GetPassword(profileID string) (string, error) {
	return keyring.Get(serviceName, passwordKey(profileID))
}

func DeletePassword(profileID string) error {
	if err := keyring.Delete(serviceName, passwordKey(profileID)); err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			return nil
		}
		return err
	}
	return nil
}

func SetPrivateKeyPassphrase(profileID, passphrase string) error {
	return keyring.Set(serviceName, passphraseKey(profileID), passphrase)
}

func GetPrivateKeyPassphrase(profileID string) (string, error) {
	return keyring.Get(serviceName, passphraseKey(profileID))
}

func DeletePrivateKeyPassphrase(profileID string) error {
	if err := keyring.Delete(serviceName, passphraseKey(profileID)); err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			return nil
		}
		return err
	}
	return nil
}

func SetMySQLPassword(profileID, password string) error {
	return keyring.Set(serviceName, mysqlPasswordKey(profileID), password)
}

func GetMySQLPassword(profileID string) (string, error) {
	return keyring.Get(serviceName, mysqlPasswordKey(profileID))
}

func DeleteMySQLPassword(profileID string) error {
	if err := keyring.Delete(serviceName, mysqlPasswordKey(profileID)); err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			return nil
		}
		return err
	}
	return nil
}
