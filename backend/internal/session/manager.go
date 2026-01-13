package session

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"

	"goterm/backend/internal/common"
	"goterm/backend/internal/profiles"
	"goterm/backend/internal/security/hostkey"
	"goterm/backend/internal/security/keyring"
)

type Session struct {
	ID        string
	ProfileID string
	Client    *ssh.Client
	State     string
	LastError string
	stopCh    chan struct{}
}

type Status struct {
	State     string `json:"state"`
	LastError string `json:"lastError"`
}

type StateEvent struct {
	SessionID string `json:"sessionId"`
	ProfileID string `json:"profileId"`
	State     string `json:"state"`
	Error     string `json:"error"`
}

type Manager struct {
	store    profiles.Store
	hostKeys *hostkey.Verifier
	emitter  common.Emitter

	mu        sync.Mutex
	sessions  map[string]*Session
	byProfile map[string]*Session
}

func NewManager(store profiles.Store, hostKeys *hostkey.Verifier, emitter common.Emitter) *Manager {
	if emitter == nil {
		emitter = common.NopEmitter{}
	}

	return &Manager{
		store:     store,
		hostKeys:  hostKeys,
		emitter:   emitter,
		sessions:  map[string]*Session{},
		byProfile: map[string]*Session{},
	}
}

func (m *Manager) Connect(ctx context.Context, profileID string) (string, error) {
	m.mu.Lock()
	if existing, ok := m.byProfile[profileID]; ok && existing.State == "connected" {
		id := existing.ID
		m.mu.Unlock()
		return id, nil
	}
	m.mu.Unlock()

	profile, err := m.store.Get(ctx, profileID)
	if err != nil {
		return "", err
	}

	config, err := m.clientConfig(profile)
	if err != nil {
		return "", err
	}

	port := profile.Port
	if port <= 0 {
		port = 22
	}
	addr := net.JoinHostPort(profile.Host, strconv.Itoa(port))

	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return "", err
	}

	id, err := common.NewID()
	if err != nil {
		_ = client.Close()
		return "", err
	}

	sess := &Session{
		ID:        id,
		ProfileID: profile.ID,
		Client:    client,
		State:     "connected",
		stopCh:    make(chan struct{}),
	}

	m.mu.Lock()
	m.sessions[id] = sess
	m.byProfile[profile.ID] = sess
	m.mu.Unlock()

	m.emitState(sess, "")

	go m.keepAlive(sess)

	return id, nil
}

func (m *Manager) Disconnect(sessionID string) error {
	sess, err := m.getSession(sessionID)
	if err != nil {
		return err
	}

	m.mu.Lock()
	delete(m.sessions, sessionID)
	delete(m.byProfile, sess.ProfileID)
	m.mu.Unlock()

	close(sess.stopCh)
	_ = sess.Client.Close()
	sess.State = "disconnected"
	m.emitState(sess, "")
	return nil
}

func (m *Manager) Status(sessionID string) (Status, error) {
	sess, err := m.getSession(sessionID)
	if err != nil {
		return Status{}, err
	}
	return Status{State: sess.State, LastError: sess.LastError}, nil
}

func (m *Manager) GetClient(sessionID string) (*ssh.Client, error) {
	sess, err := m.getSession(sessionID)
	if err != nil {
		return nil, err
	}
	if sess.State != "connected" {
		return nil, errors.New("session not connected")
	}
	return sess.Client, nil
}

func (m *Manager) getSession(sessionID string) (*Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	sess, ok := m.sessions[sessionID]
	if !ok {
		return nil, common.ErrNotFound
	}
	return sess, nil
}

func (m *Manager) keepAlive(sess *Session) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if _, _, err := sess.Client.SendRequest("keepalive@goterm", true, nil); err != nil {
				sess.LastError = err.Error()
				sess.State = "disconnected"
				m.emitState(sess, err.Error())
				return
			}
		case <-sess.stopCh:
			return
		}
	}
}

func (m *Manager) emitState(sess *Session, errMsg string) {
	m.emitter.Emit("session:state", StateEvent{
		SessionID: sess.ID,
		ProfileID: sess.ProfileID,
		State:     sess.State,
		Error:     errMsg,
	})
}

func (m *Manager) clientConfig(profile profiles.Profile) (*ssh.ClientConfig, error) {
	var auths []ssh.AuthMethod

	switch profile.AuthType {
	case "password":
		if !profile.UseKeyring {
			return nil, errors.New("password auth requires keyring storage")
		}
		password, err := keyring.GetPassword(profile.ID)
		if err != nil {
			if errors.Is(err, keyring.ErrNotFound) {
				return nil, errors.New("password not found in keyring")
			}
			return nil, fmt.Errorf("load password: %w", err)
		}
		auths = append(auths, ssh.Password(password))
	case "privateKey":
		if profile.PrivateKeyPath == "" {
			return nil, errors.New("private key path is required")
		}
		signer, err := loadSigner(profile.ID, profile.PrivateKeyPath)
		if err != nil {
			return nil, err
		}
		auths = append(auths, ssh.PublicKeys(signer))
	default:
		return nil, fmt.Errorf("unsupported auth type: %s", profile.AuthType)
	}

	verifier := *m.hostKeys
	verifier.Policy = policyFromProfile(profile.KnownHostsPolicy)

	return &ssh.ClientConfig{
		User:            profile.Username,
		Auth:            auths,
		HostKeyCallback: verifier.Callback(),
		Timeout:         10 * time.Second,
	}, nil
}

func policyFromProfile(policy string) hostkey.Policy {
	switch policy {
	case string(hostkey.PolicyStrict):
		return hostkey.PolicyStrict
	case string(hostkey.PolicyAcceptNew):
		return hostkey.PolicyAcceptNew
	default:
		return hostkey.PolicyAsk
	}
}

func loadSigner(profileID, keyPath string) (ssh.Signer, error) {
	keyBytes, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, err
	}

	signer, err := ssh.ParsePrivateKey(keyBytes)
	if err == nil {
		return signer, nil
	}

	passphrase, passErr := keyring.GetPrivateKeyPassphrase(profileID)
	if passErr != nil {
		return nil, err
	}

	signer, err = ssh.ParsePrivateKeyWithPassphrase(keyBytes, []byte(passphrase))
	if err != nil {
		return nil, err
	}
	return signer, nil
}
