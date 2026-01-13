package sqlite

import (
    "database/sql"
    "os"
    "path/filepath"

    _ "modernc.org/sqlite"
)

const schema = `
CREATE TABLE IF NOT EXISTS profiles (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    group_name TEXT NOT NULL,
    host TEXT NOT NULL,
    port INTEGER NOT NULL,
    username TEXT NOT NULL,
    auth_type TEXT NOT NULL,
    private_key_path TEXT NOT NULL,
    use_keyring INTEGER NOT NULL,
    known_hosts_policy TEXT NOT NULL
);
`

func OpenProfileStore(path string) (*ProfileStore, error) {
    if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
        return nil, err
    }

    db, err := sql.Open("sqlite", path)
    if err != nil {
        return nil, err
    }

    db.SetMaxOpenConns(1)

    if _, err := db.Exec(schema); err != nil {
        _ = db.Close()
        return nil, err
    }

    return &ProfileStore{db: db}, nil
}
