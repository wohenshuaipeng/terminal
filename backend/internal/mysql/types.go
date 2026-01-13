package mysql

type Profile struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Host           string `json:"host"`
	Port           int    `json:"port"`
	Username       string `json:"username"`
	Database       string `json:"database"`
	ConnectionType string `json:"connectionType"`
	SSHProfileID   string `json:"sshProfileId"`
	UseKeyring     bool   `json:"useKeyring"`
	UseTLS         bool   `json:"useTls"`
	TLSCAPath      string `json:"tlsCaPath"`
	TLSCertPath    string `json:"tlsCertPath"`
	TLSKeyPath     string `json:"tlsKeyPath"`
	TLSSkipVerify  bool   `json:"tlsSkipVerify"`
}

type Status struct {
	State     string `json:"state"`
	LastError string `json:"lastError"`
}

type Column struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Nullable string `json:"nullable"`
	Key      string `json:"key"`
	Default  string `json:"default"`
	Extra    string `json:"extra"`
}

type PreviewResult struct {
	Columns   []string   `json:"columns"`
	Rows      [][]string `json:"rows"`
	Truncated bool       `json:"truncated"`
}

type QueryResult struct {
	Kind         string     `json:"kind"`
	Columns      []string   `json:"columns"`
	Rows         [][]string `json:"rows"`
	AffectedRows int64      `json:"affectedRows"`
	LastInsertID int64      `json:"lastInsertId"`
	DurationMs   int64      `json:"durationMs"`
	Message      string     `json:"message"`
}
