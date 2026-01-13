package mysql

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	mysqlDriver "github.com/go-sql-driver/mysql"

	"goterm/backend/internal/security/keyring"
	"goterm/backend/internal/session"
)

var ErrNotConnected = errors.New("mysql not connected")

const (
	defaultPreviewLimit = 200
	defaultQueryLimit   = 500
)

type connection struct {
	profile   Profile
	db        *sql.DB
	state     string
	lastError string
	sessionID string
	dialName  string
	tlsName   string
}

type Manager struct {
	store    Store
	sessions *session.Manager

	mu         sync.Mutex
	conns      map[string]*connection
	dialers    map[string]struct{}
	tlsConfigs map[string]struct{}
}

func NewManager(store Store, sessions *session.Manager) *Manager {
	return &Manager{
		store:      store,
		sessions:   sessions,
		conns:      map[string]*connection{},
		dialers:    map[string]struct{}{},
		tlsConfigs: map[string]struct{}{},
	}
}

func (m *Manager) Connect(ctx context.Context, profileID string) (Status, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	profile, err := m.store.Get(ctx, profileID)
	if err != nil {
		return Status{}, err
	}

	if existing := m.getConnection(profile.ID); existing != nil && existing.state == "connected" {
		return Status{State: existing.state, LastError: existing.lastError}, nil
	}

	if !profile.UseKeyring {
		return Status{}, errors.New("mysql password requires keyring storage")
	}
	password, err := keyring.GetMySQLPassword(profile.ID)
	if err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			return Status{}, errors.New("mysql password not found in keyring")
		}
		return Status{}, fmt.Errorf("load mysql password: %w", err)
	}

	port := profile.Port
	if port <= 0 {
		port = 3306
	}
	addr := net.JoinHostPort(profile.Host, strconv.Itoa(port))

	cfg := mysqlDriver.NewConfig()
	cfg.User = profile.Username
	cfg.Passwd = password
	cfg.DBName = profile.Database
	cfg.ParseTime = true
	cfg.Net = "tcp"
	cfg.Addr = addr

	var sessionID string
	if profile.ConnectionType == "ssh" {
		if profile.SSHProfileID == "" {
			return Status{}, errors.New("ssh profile is required for tunnel connection")
		}
		if m.sessions == nil {
			return Status{}, errors.New("ssh session manager unavailable")
		}
		sessionID, err = m.sessions.Connect(ctx, profile.SSHProfileID)
		if err != nil {
			return Status{}, err
		}
		dialName := "ssh+" + profile.ID
		if err := m.registerDialer(dialName, profile.ID); err != nil {
			return Status{}, err
		}
		cfg.Net = dialName
	}

	tlsName, err := m.ensureTLSConfig(profile)
	if err != nil {
		return Status{}, err
	}
	if tlsName != "" {
		cfg.TLSConfig = tlsName
	}

	conn := &connection{
		profile:   profile,
		state:     "connecting",
		sessionID: sessionID,
	}
	m.mu.Lock()
	m.conns[profile.ID] = conn
	m.mu.Unlock()

	db, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		m.removeConnection(profile.ID)
		return Status{}, err
	}
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetMaxOpenConns(4)
	db.SetMaxIdleConns(2)

	pingCtx, cancel := context.WithTimeout(ctx, 6*time.Second)
	err = db.PingContext(pingCtx)
	cancel()
	if err != nil {
		_ = db.Close()
		m.removeConnection(profile.ID)
		return Status{}, err
	}

	m.mu.Lock()
	conn.db = db
	conn.state = "connected"
	conn.lastError = ""
	m.mu.Unlock()

	return Status{State: "connected", LastError: ""}, nil
}

func (m *Manager) Disconnect(profileID string) error {
	conn := m.getConnection(profileID)
	if conn == nil {
		return nil
	}
	m.removeConnection(profileID)
	if conn.db != nil {
		return conn.db.Close()
	}
	return nil
}

func (m *Manager) Status(profileID string) (Status, error) {
	conn := m.getConnection(profileID)
	if conn == nil {
		return Status{State: "disconnected", LastError: ""}, nil
	}
	return Status{State: conn.state, LastError: conn.lastError}, nil
}

func (m *Manager) ListDatabases(ctx context.Context, profileID string) ([]string, error) {
	db, err := m.getDB(profileID)
	if err != nil {
		return nil, err
	}
	rows, err := db.QueryContext(ctxOrBackground(ctx), "SHOW DATABASES")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		items = append(items, name)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (m *Manager) ListTables(ctx context.Context, profileID, database string) ([]string, error) {
	if database == "" {
		return nil, errors.New("database is required")
	}
	db, err := m.getDB(profileID)
	if err != nil {
		return nil, err
	}
	query := fmt.Sprintf("SHOW TABLES FROM %s", quoteIdent(database))
	rows, err := db.QueryContext(ctxOrBackground(ctx), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		items = append(items, name)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (m *Manager) TableSchema(ctx context.Context, profileID, database, table string) ([]Column, error) {
	if database == "" || table == "" {
		return nil, errors.New("database and table are required")
	}
	db, err := m.getDB(profileID)
	if err != nil {
		return nil, err
	}
	query := fmt.Sprintf("SHOW COLUMNS FROM %s.%s", quoteIdent(database), quoteIdent(table))
	rows, err := db.QueryContext(ctxOrBackground(ctx), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []Column
	for rows.Next() {
		var col Column
		var def sql.NullString
		if err := rows.Scan(&col.Name, &col.Type, &col.Nullable, &col.Key, &def, &col.Extra); err != nil {
			return nil, err
		}
		if def.Valid {
			col.Default = def.String
		}
		items = append(items, col)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (m *Manager) PreviewTable(ctx context.Context, profileID, database, table, filter, orderBy, orderDir string, limit, offset int) (PreviewResult, error) {
	if database == "" || table == "" {
		return PreviewResult{}, errors.New("database and table are required")
	}
	if limit <= 0 {
		limit = defaultPreviewLimit
	}
	if offset < 0 {
		offset = 0
	}
	queryLimit := limit + 1
	query := fmt.Sprintf("SELECT * FROM %s.%s", quoteIdent(database), quoteIdent(table))
	if strings.TrimSpace(filter) != "" {
		query += " WHERE " + filter
	}
	if orderBy != "" {
		dir := strings.ToUpper(strings.TrimSpace(orderDir))
		if dir != "DESC" {
			dir = "ASC"
		}
		query += fmt.Sprintf(" ORDER BY %s %s", quoteIdent(orderBy), dir)
	}
	query += fmt.Sprintf(" LIMIT %d OFFSET %d", queryLimit, offset)

	columns, rows, truncated, err := m.queryRows(ctx, profileID, database, query, limit)
	if err != nil {
		return PreviewResult{}, err
	}
	return PreviewResult{
		Columns:   columns,
		Rows:      rows,
		Truncated: truncated,
	}, nil
}

func (m *Manager) Query(ctx context.Context, profileID, database, query string) (QueryResult, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return QueryResult{}, errors.New("query is empty")
	}
	start := time.Now()

	var result QueryResult
	err := m.withConn(ctx, profileID, database, func(conn *sql.Conn) error {
		if isReadQuery(query) {
			columns, rows, truncated, err := queryRowsConn(ctx, conn, query, defaultQueryLimit)
			if err != nil {
				return err
			}
			result.Kind = "rows"
			result.Columns = columns
			result.Rows = rows
			if truncated {
				result.Message = "rows truncated"
			}
			return nil
		}
		res, err := conn.ExecContext(ctxOrBackground(ctx), query)
		if err != nil {
			return err
		}
		affected, _ := res.RowsAffected()
		lastID, _ := res.LastInsertId()
		result.Kind = "exec"
		result.AffectedRows = affected
		result.LastInsertID = lastID
		return nil
	})
	if err != nil {
		return QueryResult{}, err
	}
	result.DurationMs = time.Since(start).Milliseconds()
	return result, nil
}

func (m *Manager) CreateDatabase(ctx context.Context, profileID, name string) error {
	if name == "" {
		return errors.New("database name is required")
	}
	query := fmt.Sprintf("CREATE DATABASE %s", quoteIdent(name))
	_, err := m.exec(ctx, profileID, "", query)
	return err
}

func (m *Manager) DropDatabase(ctx context.Context, profileID, name string) error {
	if name == "" {
		return errors.New("database name is required")
	}
	query := fmt.Sprintf("DROP DATABASE %s", quoteIdent(name))
	_, err := m.exec(ctx, profileID, "", query)
	return err
}

func (m *Manager) DropTable(ctx context.Context, profileID, database, table string) error {
	if database == "" || table == "" {
		return errors.New("database and table are required")
	}
	query := fmt.Sprintf("DROP TABLE %s.%s", quoteIdent(database), quoteIdent(table))
	_, err := m.exec(ctx, profileID, database, query)
	return err
}

func (m *Manager) exec(ctx context.Context, profileID, database, query string) (sql.Result, error) {
	var res sql.Result
	err := m.withConn(ctx, profileID, database, func(conn *sql.Conn) error {
		var err error
		res, err = conn.ExecContext(ctxOrBackground(ctx), query)
		return err
	})
	return res, err
}

func (m *Manager) withConn(ctx context.Context, profileID, database string, fn func(*sql.Conn) error) error {
	db, err := m.getDB(profileID)
	if err != nil {
		return err
	}
	conn, err := db.Conn(ctxOrBackground(ctx))
	if err != nil {
		return err
	}
	defer conn.Close()
	if database != "" {
		if _, err := conn.ExecContext(ctxOrBackground(ctx), "USE "+quoteIdent(database)); err != nil {
			return err
		}
	}
	return fn(conn)
}

func (m *Manager) getDB(profileID string) (*sql.DB, error) {
	conn := m.getConnection(profileID)
	if conn == nil || conn.state != "connected" || conn.db == nil {
		return nil, ErrNotConnected
	}
	return conn.db, nil
}

func (m *Manager) getConnection(profileID string) *connection {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.conns[profileID]
}

func (m *Manager) removeConnection(profileID string) {
	m.mu.Lock()
	delete(m.conns, profileID)
	m.mu.Unlock()
}

func (m *Manager) registerDialer(name, profileID string) error {
	m.mu.Lock()
	_, exists := m.dialers[name]
	m.mu.Unlock()
	if exists {
		return nil
	}
	mysqlDriver.RegisterDialContext(name, func(ctx context.Context, addr string) (net.Conn, error) {
		return m.dialViaSSH(profileID, addr)
	})
	m.mu.Lock()
	m.dialers[name] = struct{}{}
	m.mu.Unlock()
	return nil
}

func (m *Manager) dialViaSSH(profileID, addr string) (net.Conn, error) {
	conn := m.getConnection(profileID)
	if conn == nil || conn.sessionID == "" {
		return nil, errors.New("ssh session not ready")
	}
	if m.sessions == nil {
		return nil, errors.New("ssh session manager unavailable")
	}
	client, err := m.sessions.GetClient(conn.sessionID)
	if err != nil {
		return nil, err
	}
	return client.Dial("tcp", addr)
}

func (m *Manager) ensureTLSConfig(profile Profile) (string, error) {
	if !profile.UseTLS {
		return "", nil
	}
	name := "goterm-" + profile.ID
	m.mu.Lock()
	_, exists := m.tlsConfigs[name]
	m.mu.Unlock()
	if exists {
		return name, nil
	}

	tlsConfig := &tls.Config{
		MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: profile.TLSSkipVerify,
	}

	if profile.TLSCAPath != "" {
		caData, err := os.ReadFile(profile.TLSCAPath)
		if err != nil {
			return "", err
		}
		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM(caData) {
			return "", errors.New("failed to parse tls CA file")
		}
		tlsConfig.RootCAs = pool
	}

	if profile.TLSCertPath != "" || profile.TLSKeyPath != "" {
		if profile.TLSCertPath == "" || profile.TLSKeyPath == "" {
			return "", errors.New("tls cert and key are required together")
		}
		cert, err := tls.LoadX509KeyPair(profile.TLSCertPath, profile.TLSKeyPath)
		if err != nil {
			return "", err
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	if err := mysqlDriver.RegisterTLSConfig(name, tlsConfig); err != nil {
		return "", err
	}
	m.mu.Lock()
	m.tlsConfigs[name] = struct{}{}
	m.mu.Unlock()
	return name, nil
}

func (m *Manager) queryRows(ctx context.Context, profileID, database, query string, limit int) ([]string, [][]string, bool, error) {
	var columns []string
	var rows [][]string
	var truncated bool
	err := m.withConn(ctx, profileID, database, func(conn *sql.Conn) error {
		var err error
		columns, rows, truncated, err = queryRowsConn(ctx, conn, query, limit)
		return err
	})
	return columns, rows, truncated, err
}

func queryRowsConn(ctx context.Context, conn *sql.Conn, query string, limit int) ([]string, [][]string, bool, error) {
	if limit <= 0 {
		limit = defaultQueryLimit
	}
	rows, err := conn.QueryContext(ctxOrBackground(ctx), query)
	if err != nil {
		return nil, nil, false, err
	}
	defer rows.Close()
	return rowsToStrings(rows, limit)
}

func rowsToStrings(rows *sql.Rows, limit int) ([]string, [][]string, bool, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, nil, false, err
	}
	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	var output [][]string
	truncated := false
	count := 0
	for rows.Next() {
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, nil, false, err
		}
		row := make([]string, len(columns))
		for i, val := range values {
			row[i] = formatValue(val)
		}
		output = append(output, row)
		count++
		if limit > 0 && count >= limit {
			if rows.Next() {
				truncated = true
			}
			break
		}
	}
	if err := rows.Err(); err != nil {
		return nil, nil, false, err
	}
	return columns, output, truncated, nil
}

func formatValue(val interface{}) string {
	switch v := val.(type) {
	case nil:
		return "NULL"
	case []byte:
		return string(v)
	case time.Time:
		return v.Format(time.RFC3339)
	default:
		return fmt.Sprint(v)
	}
}

func isReadQuery(query string) bool {
	first := strings.ToLower(firstToken(query))
	switch first {
	case "select", "show", "describe", "desc", "explain", "with":
		return true
	default:
		return false
	}
}

func firstToken(input string) string {
	fields := strings.Fields(input)
	if len(fields) == 0 {
		return ""
	}
	return fields[0]
}

func quoteIdent(name string) string {
	return "`" + strings.ReplaceAll(name, "`", "``") + "`"
}

func ctxOrBackground(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return ctx
}
