package app

import (
	"context"
	"path/filepath"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"goterm/backend/internal/common"
	"goterm/backend/internal/metrics"
	"goterm/backend/internal/mysql"
	"goterm/backend/internal/profiles"
	"goterm/backend/internal/security/hostkey"
	"goterm/backend/internal/security/keyring"
	"goterm/backend/internal/session"
	"goterm/backend/internal/sftp"
	"goterm/backend/internal/storage/sqlite"
	"goterm/backend/internal/terminal"
	"goterm/backend/internal/transfer"
)

type App struct {
	ctx context.Context

	store       profiles.Store
	mysqlStore  mysql.Store
	sessions    *session.Manager
	terminals   *terminal.Hub
	files       *sftp.Service
	transfers   *transfer.Queue
	mysql       *mysql.Manager
	prompts     *HostKeyPromptManager
	dataDir     string
	hostKeyPath string
}

func NewApp(cfg Config) (*App, error) {
	dataDir := cfg.DataDir
	if dataDir == "" {
		var err error
		dataDir, err = DefaultDataDir()
		if err != nil {
			return nil, err
		}
	}

	emitter := cfg.Emitter
	if emitter == nil {
		emitter = common.NopEmitter{}
	}

	hostKeyPath := cfg.HostKeyPath
	if hostKeyPath == "" {
		hostKeyPath = filepath.Join(dataDir, "known_hosts")
	}

	store := cfg.ProfileStore
	if store == nil {
		sqliteStore, err := sqlite.OpenProfileStore(filepath.Join(dataDir, "profiles.db"))
		if err != nil {
			return nil, err
		}
		store = sqliteStore
	}

	mysqlStore := cfg.MySQLStore
	if mysqlStore == nil {
		sqliteStore, err := sqlite.OpenMySQLProfileStore(filepath.Join(dataDir, "mysql.db"))
		if err != nil {
			return nil, err
		}
		mysqlStore = sqliteStore
	}

	promptManager := NewHostKeyPromptManager(emitter)

	verifier := &hostkey.Verifier{
		Path:   hostKeyPath,
		Policy: hostkey.PolicyAsk,
		Prompt: promptManager.Ask,
	}

	sessions := session.NewManager(store, verifier, emitter)

	app := &App{
		store:       store,
		mysqlStore:  mysqlStore,
		sessions:    sessions,
		terminals:   terminal.NewHub(sessions, emitter),
		files:       sftp.NewService(sessions),
		transfers:   transfer.NewQueue(sessions, emitter, 2),
		mysql:       mysql.NewManager(mysqlStore, sessions),
		prompts:     promptManager,
		dataDir:     dataDir,
		hostKeyPath: hostKeyPath,
	}

	return app, nil
}

func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) ProfilesList() ([]profiles.Profile, error) {
	return a.store.List(a.ctxOrBackground())
}

func (a *App) ProfilesSave(profile profiles.Profile) (string, error) {
	return a.store.Save(a.ctxOrBackground(), profile)
}

func (a *App) ProfilesDelete(profileID string) error {
	return a.store.Delete(a.ctxOrBackground(), profileID)
}

func (a *App) MySQLProfilesList() ([]mysql.Profile, error) {
	return a.mysqlStore.List(a.ctxOrBackground())
}

func (a *App) MySQLProfilesSave(profile mysql.Profile) (string, error) {
	return a.mysqlStore.Save(a.ctxOrBackground(), profile)
}

func (a *App) MySQLProfilesDelete(profileID string) error {
	return a.mysqlStore.Delete(a.ctxOrBackground(), profileID)
}

func (a *App) MySQLConnect(profileID string) (mysql.Status, error) {
	return a.mysql.Connect(a.ctxOrBackground(), profileID)
}

func (a *App) MySQLDisconnect(profileID string) error {
	return a.mysql.Disconnect(profileID)
}

func (a *App) MySQLStatus(profileID string) (mysql.Status, error) {
	return a.mysql.Status(profileID)
}

func (a *App) MySQLListDatabases(profileID string) ([]string, error) {
	return a.mysql.ListDatabases(a.ctxOrBackground(), profileID)
}

func (a *App) MySQLListTables(profileID, database string) ([]string, error) {
	return a.mysql.ListTables(a.ctxOrBackground(), profileID, database)
}

func (a *App) MySQLTableSchema(profileID, database, table string) ([]mysql.Column, error) {
	return a.mysql.TableSchema(a.ctxOrBackground(), profileID, database, table)
}

func (a *App) MySQLPreviewTable(profileID, database, table, filter, orderBy, orderDir string, limit, offset int) (mysql.PreviewResult, error) {
	return a.mysql.PreviewTable(a.ctxOrBackground(), profileID, database, table, filter, orderBy, orderDir, limit, offset)
}

func (a *App) MySQLQuery(profileID, database, query string) (mysql.QueryResult, error) {
	return a.mysql.Query(a.ctxOrBackground(), profileID, database, query)
}

func (a *App) MySQLCreateDatabase(profileID, name string) error {
	return a.mysql.CreateDatabase(a.ctxOrBackground(), profileID, name)
}

func (a *App) MySQLDropDatabase(profileID, name string) error {
	return a.mysql.DropDatabase(a.ctxOrBackground(), profileID, name)
}

func (a *App) MySQLDropTable(profileID, database, table string) error {
	return a.mysql.DropTable(a.ctxOrBackground(), profileID, database, table)
}

func (a *App) SystemStats() (metrics.Stats, error) {
	return metrics.Snapshot()
}

func (a *App) MySQLCredentialsSetPassword(profileID, password string) error {
	return keyring.SetMySQLPassword(profileID, password)
}

func (a *App) MySQLCredentialsDelete(profileID string) error {
	return keyring.DeleteMySQLPassword(profileID)
}

func (a *App) SessionConnect(profileID string) (string, error) {
	return a.sessions.Connect(a.ctxOrBackground(), profileID)
}

func (a *App) SessionDisconnect(sessionID string) error {
	return a.sessions.Disconnect(sessionID)
}

func (a *App) SessionStatus(sessionID string) (session.Status, error) {
	return a.sessions.Status(sessionID)
}

func (a *App) TerminalOpen(sessionID string, cols, rows int) (string, error) {
	return a.terminals.Open(sessionID, cols, rows)
}

func (a *App) TerminalWrite(termID, data string) error {
	return a.terminals.Write(termID, data)
}

func (a *App) TerminalResize(termID string, cols, rows int) error {
	return a.terminals.Resize(termID, cols, rows)
}

func (a *App) TerminalClose(termID string) error {
	return a.terminals.Close(termID)
}

func (a *App) FilesList(sessionID, path string) ([]sftp.FileEntry, error) {
	return a.files.List(sessionID, path)
}

func (a *App) FilesStat(sessionID, path string) (sftp.FileEntry, error) {
	return a.files.Stat(sessionID, path)
}

func (a *App) FilesMkdir(sessionID, path string) error {
	return a.files.Mkdir(sessionID, path)
}

func (a *App) FilesRemove(sessionID, path string, recursive bool) error {
	return a.files.Remove(sessionID, path, recursive)
}

func (a *App) FilesRename(sessionID, fromPath, toPath string) error {
	return a.files.Rename(sessionID, fromPath, toPath)
}

func (a *App) TransferDownload(sessionID, remotePath, localPath string) (string, error) {
	return a.transfers.Download(sessionID, remotePath, localPath)
}

func (a *App) TransferUpload(sessionID, localPath, remotePath string) (string, error) {
	return a.transfers.Upload(sessionID, localPath, remotePath)
}

func (a *App) TransferCancel(taskID string) error {
	return a.transfers.Cancel(taskID)
}

func (a *App) TransferListTasks() []transfer.Task {
	return a.transfers.ListTasks()
}

func (a *App) HostKeyRespond(requestID string, allow bool) error {
	return a.prompts.Resolve(requestID, allow)
}

func (a *App) CredentialsSetPassword(profileID, password string) error {
	return keyring.SetPassword(profileID, password)
}

func (a *App) CredentialsSetPrivateKeyPassphrase(profileID, passphrase string) error {
	return keyring.SetPrivateKeyPassphrase(profileID, passphrase)
}

func (a *App) CredentialsDelete(profileID string) error {
	if err := keyring.DeletePassword(profileID); err != nil {
		return err
	}
	return keyring.DeletePrivateKeyPassphrase(profileID)
}

func (a *App) DialogOpenFile(title string) (string, error) {
	return runtime.OpenFileDialog(a.ctxOrBackground(), runtime.OpenDialogOptions{
		Title: title,
	})
}

func (a *App) DialogSaveFile(title, filename string) (string, error) {
	return runtime.SaveFileDialog(a.ctxOrBackground(), runtime.SaveDialogOptions{
		Title:           title,
		DefaultFilename: filename,
	})
}

func (a *App) ctxOrBackground() context.Context {
	if a.ctx != nil {
		return a.ctx
	}
	return context.Background()
}
