package app

import (
    "errors"
    "sync"
    "time"

    "golang.org/x/crypto/ssh"

    "goterm/backend/internal/common"
)

type HostKeyPrompt struct {
    ID          string `json:"id"`
    Host        string `json:"host"`
    Fingerprint string `json:"fingerprint"`
}

type HostKeyPromptManager struct {
    emitter common.Emitter

    mu      sync.Mutex
    pending map[string]chan bool
}

func NewHostKeyPromptManager(emitter common.Emitter) *HostKeyPromptManager {
    if emitter == nil {
        emitter = common.NopEmitter{}
    }
    return &HostKeyPromptManager{
        emitter: emitter,
        pending: map[string]chan bool{},
    }
}

func (m *HostKeyPromptManager) Ask(host string, _ ssh.PublicKey, fingerprint string) (bool, error) {
    id, err := common.NewID()
    if err != nil {
        return false, err
    }

    ch := make(chan bool, 1)

    m.mu.Lock()
    m.pending[id] = ch
    m.mu.Unlock()

    m.emitter.Emit("hostkey:prompt", HostKeyPrompt{
        ID:          id,
        Host:        host,
        Fingerprint: fingerprint,
    })

    select {
    case allowed := <-ch:
        return allowed, nil
    case <-time.After(2 * time.Minute):
        m.mu.Lock()
        delete(m.pending, id)
        m.mu.Unlock()
        return false, errors.New("host key prompt timeout")
    }
}

func (m *HostKeyPromptManager) Resolve(id string, allow bool) error {
    m.mu.Lock()
    ch, ok := m.pending[id]
    if ok {
        delete(m.pending, id)
    }
    m.mu.Unlock()

    if !ok {
        return common.ErrNotFound
    }

    ch <- allow
    close(ch)
    return nil
}
