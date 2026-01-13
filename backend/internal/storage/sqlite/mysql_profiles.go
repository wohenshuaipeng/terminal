package sqlite

import (
	"context"
	"database/sql"

	"goterm/backend/internal/common"
	"goterm/backend/internal/mysql"
)

type MySQLProfileStore struct {
	db *sql.DB
}

func (s *MySQLProfileStore) List(ctx context.Context) ([]mysql.Profile, error) {
	rows, err := s.db.QueryContext(ctx, `
        SELECT id, name, host, port, username, database_name, connection_type, ssh_profile_id,
               use_keyring, use_tls, tls_ca_path, tls_cert_path, tls_key_path, tls_skip_verify
        FROM mysql_profiles
        ORDER BY name
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []mysql.Profile
	for rows.Next() {
		var p mysql.Profile
		var useKeyringInt int
		var useTLSInt int
		var skipVerifyInt int
		if err := rows.Scan(
			&p.ID,
			&p.Name,
			&p.Host,
			&p.Port,
			&p.Username,
			&p.Database,
			&p.ConnectionType,
			&p.SSHProfileID,
			&useKeyringInt,
			&useTLSInt,
			&p.TLSCAPath,
			&p.TLSCertPath,
			&p.TLSKeyPath,
			&skipVerifyInt,
		); err != nil {
			return nil, err
		}
		p.UseKeyring = useKeyringInt != 0
		p.UseTLS = useTLSInt != 0
		p.TLSSkipVerify = skipVerifyInt != 0
		items = append(items, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (s *MySQLProfileStore) Get(ctx context.Context, id string) (mysql.Profile, error) {
	row := s.db.QueryRowContext(ctx, `
        SELECT id, name, host, port, username, database_name, connection_type, ssh_profile_id,
               use_keyring, use_tls, tls_ca_path, tls_cert_path, tls_key_path, tls_skip_verify
        FROM mysql_profiles
        WHERE id = ?
    `, id)

	var p mysql.Profile
	var useKeyringInt int
	var useTLSInt int
	var skipVerifyInt int
	if err := row.Scan(
		&p.ID,
		&p.Name,
		&p.Host,
		&p.Port,
		&p.Username,
		&p.Database,
		&p.ConnectionType,
		&p.SSHProfileID,
		&useKeyringInt,
		&useTLSInt,
		&p.TLSCAPath,
		&p.TLSCertPath,
		&p.TLSKeyPath,
		&skipVerifyInt,
	); err != nil {
		if err == sql.ErrNoRows {
			return mysql.Profile{}, common.ErrNotFound
		}
		return mysql.Profile{}, err
	}
	p.UseKeyring = useKeyringInt != 0
	p.UseTLS = useTLSInt != 0
	p.TLSSkipVerify = skipVerifyInt != 0
	return p, nil
}

func (s *MySQLProfileStore) Save(ctx context.Context, p mysql.Profile) (string, error) {
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
	useTLSInt := 0
	if p.UseTLS {
		useTLSInt = 1
	}
	skipVerifyInt := 0
	if p.TLSSkipVerify {
		skipVerifyInt = 1
	}

	_, err := s.db.ExecContext(ctx, `
        INSERT INTO mysql_profiles (
            id, name, host, port, username, database_name, connection_type, ssh_profile_id,
            use_keyring, use_tls, tls_ca_path, tls_cert_path, tls_key_path, tls_skip_verify
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
        ON CONFLICT(id) DO UPDATE SET
            name = excluded.name,
            host = excluded.host,
            port = excluded.port,
            username = excluded.username,
            database_name = excluded.database_name,
            connection_type = excluded.connection_type,
            ssh_profile_id = excluded.ssh_profile_id,
            use_keyring = excluded.use_keyring,
            use_tls = excluded.use_tls,
            tls_ca_path = excluded.tls_ca_path,
            tls_cert_path = excluded.tls_cert_path,
            tls_key_path = excluded.tls_key_path,
            tls_skip_verify = excluded.tls_skip_verify
    `,
		p.ID,
		p.Name,
		p.Host,
		p.Port,
		p.Username,
		p.Database,
		p.ConnectionType,
		p.SSHProfileID,
		useKeyringInt,
		useTLSInt,
		p.TLSCAPath,
		p.TLSCertPath,
		p.TLSKeyPath,
		skipVerifyInt,
	)
	if err != nil {
		return "", err
	}
	return p.ID, nil
}

func (s *MySQLProfileStore) Delete(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM mysql_profiles WHERE id = ?", id)
	return err
}
