package hostkey

import (
    "errors"
    "net"
    "os"
    "path/filepath"

    "golang.org/x/crypto/ssh"
    "golang.org/x/crypto/ssh/knownhosts"
)

type Policy string

const (
    PolicyAsk       Policy = "ask"
    PolicyStrict    Policy = "strict"
    PolicyAcceptNew Policy = "accept-new"
)

var ErrHostKeyRejected = errors.New("host key rejected")

type PromptFunc func(host string, key ssh.PublicKey, fingerprint string) (bool, error)

type Verifier struct {
    Path   string
    Policy Policy
    Prompt PromptFunc
}

func (v *Verifier) Callback() ssh.HostKeyCallback {
    return func(hostname string, remote net.Addr, key ssh.PublicKey) error {
        if err := ensureFile(v.Path); err != nil {
            return err
        }

        callback, err := knownhosts.New(v.Path)
        if err != nil {
            return err
        }

        if err := callback(hostname, remote, key); err == nil {
            return nil
        } else if keyErr, ok := err.(*knownhosts.KeyError); ok {
            if len(keyErr.Want) > 0 {
                return err
            }

            switch v.Policy {
            case PolicyStrict:
                return err
            case PolicyAcceptNew:
                return appendHostKey(v.Path, hostname, key)
            case PolicyAsk:
                if v.Prompt == nil {
                    return ErrHostKeyRejected
                }
                allowed, promptErr := v.Prompt(hostname, key, ssh.FingerprintSHA256(key))
                if promptErr != nil {
                    return promptErr
                }
                if !allowed {
                    return ErrHostKeyRejected
                }
                return appendHostKey(v.Path, hostname, key)
            default:
                return err
            }
        } else {
            return err
        }
    }
}

func ensureFile(path string) error {
    if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
        return err
    }
    file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o600)
    if err != nil {
        return err
    }
    return file.Close()
}

func appendHostKey(path, hostname string, key ssh.PublicKey) error {
    normalized := knownhosts.Normalize(hostname)
    line := knownhosts.Line([]string{normalized}, key)

    file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o600)
    if err != nil {
        return err
    }
    defer file.Close()

    if _, err := file.WriteString(line + "\n"); err != nil {
        return err
    }
    return nil
}
