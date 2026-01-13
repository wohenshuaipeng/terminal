package sqlite

import (
    "context"
    "database/sql"

    "goterm/backend/internal/common"
    "goterm/backend/internal/profiles"
)

type ProfileStore struct {
    db *sql.DB
}

func (s *ProfileStore) List(ctx context.Context) ([]profiles.Profile, error) {
    rows, err := s.db.QueryContext(ctx, `
        SELECT id, name, group_name, host, port, username, auth_type, private_key_path, use_keyring, known_hosts_policy
        FROM profiles
        ORDER BY group_name, name
    `)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var items []profiles.Profile
    for rows.Next() {
        var p profiles.Profile
        var useKeyringInt int
        if err := rows.Scan(
            &p.ID,
            &p.Name,
            &p.Group,
            &p.Host,
            &p.Port,
            &p.Username,
            &p.AuthType,
            &p.PrivateKeyPath,
            &useKeyringInt,
            &p.KnownHostsPolicy,
        ); err != nil {
            return nil, err
        }
        p.UseKeyring = useKeyringInt != 0
        items = append(items, p)
    }
    if err := rows.Err(); err != nil {
        return nil, err
    }
    return items, nil
}

func (s *ProfileStore) Get(ctx context.Context, id string) (profiles.Profile, error) {
    row := s.db.QueryRowContext(ctx, `
        SELECT id, name, group_name, host, port, username, auth_type, private_key_path, use_keyring, known_hosts_policy
        FROM profiles
        WHERE id = ?
    `, id)

    var p profiles.Profile
    var useKeyringInt int
    if err := row.Scan(
        &p.ID,
        &p.Name,
        &p.Group,
        &p.Host,
        &p.Port,
        &p.Username,
        &p.AuthType,
        &p.PrivateKeyPath,
        &useKeyringInt,
        &p.KnownHostsPolicy,
    ); err != nil {
        if err == sql.ErrNoRows {
            return profiles.Profile{}, common.ErrNotFound
        }
        return profiles.Profile{}, err
    }
    p.UseKeyring = useKeyringInt != 0
    return p, nil
}

func (s *ProfileStore) Save(ctx context.Context, p profiles.Profile) (string, error) {
    if p.ID == "" {
        id, err := common.NewID()
        if err != nil {
            return "", err
        }
        p.ID = id
    }

    useKeyringInt := 0
    if p.UseKeyring {
        useKeyringInt = 1
    }

    _, err := s.db.ExecContext(ctx, `
        INSERT INTO profiles (
            id, name, group_name, host, port, username, auth_type, private_key_path, use_keyring, known_hosts_policy
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
        ON CONFLICT(id) DO UPDATE SET
            name = excluded.name,
            group_name = excluded.group_name,
            host = excluded.host,
            port = excluded.port,
            username = excluded.username,
            auth_type = excluded.auth_type,
            private_key_path = excluded.private_key_path,
            use_keyring = excluded.use_keyring,
            known_hosts_policy = excluded.known_hosts_policy
    `,
        p.ID,
        p.Name,
        p.Group,
        p.Host,
        p.Port,
        p.Username,
        p.AuthType,
        p.PrivateKeyPath,
        useKeyringInt,
        p.KnownHostsPolicy,
    )
    if err != nil {
        return "", err
    }

    return p.ID, nil
}

func (s *ProfileStore) Delete(ctx context.Context, id string) error {
    _, err := s.db.ExecContext(ctx, "DELETE FROM profiles WHERE id = ?", id)
    return err
}
