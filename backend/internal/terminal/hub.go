package terminal

import (
    "errors"
    "io"
    "sync"

    "golang.org/x/crypto/ssh"

    "goterm/backend/internal/common"
)

type ClientProvider interface {
    GetClient(sessionID string) (*ssh.Client, error)
}

type Terminal struct {
    ID        string
    SessionID string
    SSH       *ssh.Session
    Stdin     io.WriteCloser
}

type DataEvent struct {
    TermID string `json:"termId"`
    Chunk  string `json:"chunk"`
}

type ExitEvent struct {
    TermID string `json:"termId"`
    Code   int    `json:"code"`
}

type Hub struct {
    provider ClientProvider
    emitter  common.Emitter

    mu    sync.Mutex
    terms map[string]*Terminal
}

func NewHub(provider ClientProvider, emitter common.Emitter) *Hub {
    if emitter == nil {
        emitter = common.NopEmitter{}
    }
    return &Hub{
        provider: provider,
        emitter:  emitter,
        terms:    map[string]*Terminal{},
    }
}

func (h *Hub) Open(sessionID string, cols, rows int) (string, error) {
    client, err := h.provider.GetClient(sessionID)
    if err != nil {
        return "", err
    }

    sshSession, err := client.NewSession()
    if err != nil {
        return "", err
    }

    if cols <= 0 {
        cols = 80
    }
    if rows <= 0 {
        rows = 24
    }

    modes := ssh.TerminalModes{
        ssh.ECHO:          1,
        ssh.TTY_OP_ISPEED: 14400,
        ssh.TTY_OP_OSPEED: 14400,
    }

    if err := sshSession.RequestPty("xterm-256color", rows, cols, modes); err != nil {
        _ = sshSession.Close()
        return "", err
    }

    stdin, err := sshSession.StdinPipe()
    if err != nil {
        _ = sshSession.Close()
        return "", err
    }

    stdout, err := sshSession.StdoutPipe()
    if err != nil {
        _ = sshSession.Close()
        return "", err
    }

    stderr, err := sshSession.StderrPipe()
    if err != nil {
        _ = sshSession.Close()
        return "", err
    }

    if err := sshSession.Shell(); err != nil {
        _ = sshSession.Close()
        return "", err
    }

    id, err := common.NewID()
    if err != nil {
        _ = sshSession.Close()
        return "", err
    }

    term := &Terminal{
        ID:        id,
        SessionID: sessionID,
        SSH:       sshSession,
        Stdin:     stdin,
    }

    h.mu.Lock()
    h.terms[id] = term
    h.mu.Unlock()

    go h.stream(id, stdout)
    go h.stream(id, stderr)
    go h.wait(id, sshSession)

    return id, nil
}

func (h *Hub) Write(termID string, data string) error {
    term, err := h.get(termID)
    if err != nil {
        return err
    }
    _, err = io.WriteString(term.Stdin, data)
    return err
}

func (h *Hub) Resize(termID string, cols, rows int) error {
    term, err := h.get(termID)
    if err != nil {
        return err
    }
    if cols <= 0 || rows <= 0 {
        return errors.New("invalid terminal size")
    }
    return term.SSH.WindowChange(rows, cols)
}

func (h *Hub) Close(termID string) error {
    term, err := h.get(termID)
    if err != nil {
        return err
    }

    h.mu.Lock()
    delete(h.terms, termID)
    h.mu.Unlock()

    return term.SSH.Close()
}

func (h *Hub) get(termID string) (*Terminal, error) {
    h.mu.Lock()
    defer h.mu.Unlock()
    term, ok := h.terms[termID]
    if !ok {
        return nil, common.ErrNotFound
    }
    return term, nil
}

func (h *Hub) stream(termID string, reader io.Reader) {
    buf := make([]byte, 4096)
    for {
        n, err := reader.Read(buf)
        if n > 0 {
            h.emitter.Emit("terminal:data", DataEvent{
                TermID: termID,
                Chunk:  string(buf[:n]),
            })
        }
        if err != nil {
            if err != io.EOF {
                h.emitter.Emit("terminal:data", DataEvent{
                    TermID: termID,
                    Chunk:  "\r\n[terminal stream error]\r\n",
                })
            }
            return
        }
    }
}

func (h *Hub) wait(termID string, sshSession *ssh.Session) {
    err := sshSession.Wait()
    code := 0
    if err != nil {
        if exitErr, ok := err.(*ssh.ExitError); ok {
            code = exitErr.ExitStatus()
        } else {
            code = 1
        }
    }

    h.emitter.Emit("terminal:exit", ExitEvent{TermID: termID, Code: code})

    h.mu.Lock()
    delete(h.terms, termID)
    h.mu.Unlock()
}
