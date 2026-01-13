package app

import (
	"os"
	"path/filepath"

	"goterm/backend/internal/common"
	"goterm/backend/internal/mysql"
	"goterm/backend/internal/profiles"
)

type Config struct {
	DataDir      string
	Emitter      common.Emitter
	ProfileStore profiles.Store
	MySQLStore   mysql.Store
	HostKeyPath  string
}

func DefaultDataDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".goterm"), nil
}
