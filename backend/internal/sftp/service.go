package sftp

import (
	"errors"
	"path"

	sftplib "github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"

	"goterm/backend/internal/common"
)

type ClientProvider interface {
	GetClient(sessionID string) (*ssh.Client, error)
}

type FileEntry struct {
	Name  string `json:"name"`
	Path  string `json:"path"`
	IsDir bool   `json:"isDir"`
	Size  int64  `json:"size"`
	Mode  string `json:"mode"`
	Mtime int64  `json:"mtime"`
}

type Service struct {
	provider ClientProvider
}

func NewService(provider ClientProvider) *Service {
	return &Service{provider: provider}
}

func (s *Service) List(sessionID, targetPath string) ([]FileEntry, error) {
	entries := []FileEntry{}
	err := s.withClient(sessionID, func(client *sftplib.Client) error {
		items, err := client.ReadDir(targetPath)
		if err != nil {
			return err
		}
		for _, item := range items {
			entries = append(entries, FileEntry{
				Name:  item.Name(),
				Path:  path.Join(targetPath, item.Name()),
				IsDir: item.IsDir(),
				Size:  item.Size(),
				Mode:  item.Mode().String(),
				Mtime: item.ModTime().Unix(),
			})
		}
		return nil
	})
	return entries, err
}

func (s *Service) Stat(sessionID, targetPath string) (FileEntry, error) {
	var entry FileEntry
	err := s.withClient(sessionID, func(client *sftplib.Client) error {
		info, err := client.Stat(targetPath)
		if err != nil {
			return err
		}
		entry = FileEntry{
			Name:  info.Name(),
			Path:  targetPath,
			IsDir: info.IsDir(),
			Size:  info.Size(),
			Mode:  info.Mode().String(),
			Mtime: info.ModTime().Unix(),
		}
		return nil
	})
	return entry, err
}

func (s *Service) Mkdir(sessionID, targetPath string) error {
	return s.withClient(sessionID, func(client *sftplib.Client) error {
		return client.Mkdir(targetPath)
	})
}

func (s *Service) Remove(sessionID, targetPath string, recursive bool) error {
	return s.withClient(sessionID, func(client *sftplib.Client) error {
		info, err := client.Stat(targetPath)
		if err != nil {
			return err
		}
		if info.IsDir() {
			if !recursive {
				return errors.New("directory requires recursive remove")
			}
			if err := removeAll(client, targetPath); err != nil {
				return err
			}
			return client.RemoveDirectory(targetPath)
		}
		return client.Remove(targetPath)
	})
}

func (s *Service) Rename(sessionID, fromPath, toPath string) error {
	return s.withClient(sessionID, func(client *sftplib.Client) error {
		return client.Rename(fromPath, toPath)
	})
}

func (s *Service) withClient(sessionID string, fn func(*sftplib.Client) error) error {
	client, err := s.provider.GetClient(sessionID)
	if err != nil {
		return err
	}

	sftpClient, err := sftplib.NewClient(client)
	if err != nil {
		return err
	}
	defer sftpClient.Close()

	return fn(sftpClient)
}

func removeAll(client *sftplib.Client, targetPath string) error {
	entries, err := client.ReadDir(targetPath)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		child := path.Join(targetPath, entry.Name())
		if entry.IsDir() {
			if err := removeAll(client, child); err != nil {
				return err
			}
			if err := client.RemoveDirectory(child); err != nil {
				return err
			}
			continue
		}
		if err := client.Remove(child); err != nil {
			return err
		}
	}
	return nil
}

var ErrNotFound = common.ErrNotFound
