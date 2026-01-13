package sqlite

import (
	"database/sql"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

const mysqlSchema = `
CREATE TABLE IF NOT EXISTS mysql_profiles (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    host TEXT NOT NULL,
    port INTEGER NOT NULL,
    username TEXT NOT NULL,
    database_name TEXT NOT NULL,
    connection_type TEXT NOT NULL,
    ssh_profile_id TEXT NOT NULL,
    use_keyring INTEGER NOT NULL,
    use_tls INTEGER NOT NULL,
    tls_ca_path TEXT NOT NULL,
    tls_cert_path TEXT NOT NULL,
    tls_key_path TEXT NOT NULL,
    tls_skip_verify INTEGER NOT NULL
);
`

func OpenMySQLProfileStore(path string) (*MySQLProfileStore, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(1)

	if _, err := db.Exec(mysqlSchema); err != nil {
		_ = db.Close()
		return nil, err
	}

	return &MySQLProfileStore{db: db}, nil
}
