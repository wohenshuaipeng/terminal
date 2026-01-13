package transfer

import (
	"context"
	"errors"
	"io"
	"os"
	"sync"
	"time"

	sftplib "github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"

	"goterm/backend/internal/common"
)

type ClientProvider interface {
	GetClient(sessionID string) (*ssh.Client, error)
}

type Task struct {
	ID         string `json:"id"`
	SessionID  string `json:"sessionId"`
	LocalPath  string `json:"localPath"`
	RemotePath string `json:"remotePath"`
	TotalBytes int64  `json:"totalBytes"`
	DoneBytes  int64  `json:"doneBytes"`
	State      string `json:"state"`
	Direction  string `json:"direction"`
}

type ProgressEvent struct {
	TaskID     string `json:"taskId"`
	SessionID  string `json:"sessionId"`
	LocalPath  string `json:"localPath"`
	RemotePath string `json:"remotePath"`
	Direction  string `json:"direction"`
	DoneBytes  int64  `json:"doneBytes"`
	TotalBytes int64  `json:"totalBytes"`
	SpeedBytes int64  `json:"speedBytes"`
	State      string `json:"state"`
}

type DoneEvent struct {
	TaskID     string `json:"taskId"`
	SessionID  string `json:"sessionId"`
	LocalPath  string `json:"localPath"`
	RemotePath string `json:"remotePath"`
	Direction  string `json:"direction"`
}

type ErrorEvent struct {
	TaskID     string `json:"taskId"`
	SessionID  string `json:"sessionId"`
	LocalPath  string `json:"localPath"`
	RemotePath string `json:"remotePath"`
	Direction  string `json:"direction"`
	Message    string `json:"message"`
}

type taskState struct {
	task   Task
	cancel context.CancelFunc
}

type Queue struct {
	provider ClientProvider
	emitter  common.Emitter
	sem      chan struct{}

	mu    sync.Mutex
	tasks map[string]*taskState
}

func NewQueue(provider ClientProvider, emitter common.Emitter, maxConcurrent int) *Queue {
	if emitter == nil {
		emitter = common.NopEmitter{}
	}
	if maxConcurrent <= 0 {
		maxConcurrent = 2
	}
	return &Queue{
		provider: provider,
		emitter:  emitter,
		sem:      make(chan struct{}, maxConcurrent),
		tasks:    map[string]*taskState{},
	}
}

func (q *Queue) Download(sessionID, remotePath, localPath string) (string, error) {
	id, err := common.NewID()
	if err != nil {
		return "", err
	}

	state := &taskState{
		task: Task{
			ID:         id,
			SessionID:  sessionID,
			LocalPath:  localPath,
			RemotePath: remotePath,
			State:      "queued",
			Direction:  "download",
		},
	}

	q.mu.Lock()
	q.tasks[id] = state
	q.mu.Unlock()

	go q.runDownload(state)

	return id, nil
}

func (q *Queue) Upload(sessionID, localPath, remotePath string) (string, error) {
	id, err := common.NewID()
	if err != nil {
		return "", err
	}

	state := &taskState{
		task: Task{
			ID:         id,
			SessionID:  sessionID,
			LocalPath:  localPath,
			RemotePath: remotePath,
			State:      "queued",
			Direction:  "upload",
		},
	}

	q.mu.Lock()
	q.tasks[id] = state
	q.mu.Unlock()

	go q.runUpload(state)

	return id, nil
}

func (q *Queue) Cancel(taskID string) error {
	q.mu.Lock()
	state, ok := q.tasks[taskID]
	q.mu.Unlock()
	if !ok {
		return common.ErrNotFound
	}

	if state.cancel != nil {
		state.cancel()
		return nil
	}

	return errors.New("task not started")
}

func (q *Queue) ListTasks() []Task {
	q.mu.Lock()
	defer q.mu.Unlock()

	tasks := make([]Task, 0, len(q.tasks))
	for _, state := range q.tasks {
		tasks = append(tasks, state.task)
	}
	return tasks
}

func (q *Queue) runDownload(state *taskState) {
	q.sem <- struct{}{}
	defer func() { <-q.sem }()

	ctx, cancel := context.WithCancel(context.Background())
	state.cancel = cancel

	q.setState(state, "running")

	client, err := q.provider.GetClient(state.task.SessionID)
	if err != nil {
		q.fail(state, err)
		return
	}

	sftpClient, err := sftplib.NewClient(client)
	if err != nil {
		q.fail(state, err)
		return
	}
	defer sftpClient.Close()

	remoteFile, err := sftpClient.Open(state.task.RemotePath)
	if err != nil {
		q.fail(state, err)
		return
	}
	defer remoteFile.Close()

	info, err := remoteFile.Stat()
	if err != nil {
		q.fail(state, err)
		return
	}

	state.task.TotalBytes = info.Size()

	localFile, err := os.Create(state.task.LocalPath)
	if err != nil {
		q.fail(state, err)
		return
	}
	defer localFile.Close()

	q.copyWithProgress(ctx, state, remoteFile, localFile)
}

func (q *Queue) runUpload(state *taskState) {
	q.sem <- struct{}{}
	defer func() { <-q.sem }()

	ctx, cancel := context.WithCancel(context.Background())
	state.cancel = cancel

	q.setState(state, "running")

	client, err := q.provider.GetClient(state.task.SessionID)
	if err != nil {
		q.fail(state, err)
		return
	}

	sftpClient, err := sftplib.NewClient(client)
	if err != nil {
		q.fail(state, err)
		return
	}
	defer sftpClient.Close()

	localFile, err := os.Open(state.task.LocalPath)
	if err != nil {
		q.fail(state, err)
		return
	}
	defer localFile.Close()

	info, err := localFile.Stat()
	if err != nil {
		q.fail(state, err)
		return
	}
	state.task.TotalBytes = info.Size()

	remoteFile, err := sftpClient.Create(state.task.RemotePath)
	if err != nil {
		q.fail(state, err)
		return
	}
	defer remoteFile.Close()

	q.copyWithProgress(ctx, state, localFile, remoteFile)
}

func (q *Queue) copyWithProgress(ctx context.Context, state *taskState, reader io.Reader, writer io.Writer) {
	buf := make([]byte, 32*1024)
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	var lastEmitted time.Time
	var lastDone int64

	for {
		select {
		case <-ctx.Done():
			q.fail(state, errors.New("canceled"))
			return
		default:
		}

		n, err := reader.Read(buf)
		if n > 0 {
			if _, werr := writer.Write(buf[:n]); werr != nil {
				q.fail(state, werr)
				return
			}
			state.task.DoneBytes += int64(n)
		}

		select {
		case <-ticker.C:
			speed := int64(0)
			now := time.Now()
			if !lastEmitted.IsZero() {
				elapsed := now.Sub(lastEmitted).Seconds()
				if elapsed > 0 {
					speed = int64(float64(state.task.DoneBytes-lastDone) / elapsed)
				}
			}
			lastEmitted = now
			lastDone = state.task.DoneBytes
			q.emitProgress(state, speed)
		default:
		}

		if err != nil {
			if err == io.EOF {
				q.emitProgress(state, 0)
				q.complete(state)
				return
			}
			q.fail(state, err)
			return
		}
	}
}

func (q *Queue) setState(state *taskState, newState string) {
	q.mu.Lock()
	state.task.State = newState
	q.mu.Unlock()
}

func (q *Queue) emitProgress(state *taskState, speed int64) {
	q.emitter.Emit("transfer:progress", ProgressEvent{
		TaskID:     state.task.ID,
		SessionID:  state.task.SessionID,
		LocalPath:  state.task.LocalPath,
		RemotePath: state.task.RemotePath,
		Direction:  state.task.Direction,
		DoneBytes:  state.task.DoneBytes,
		TotalBytes: state.task.TotalBytes,
		SpeedBytes: speed,
		State:      state.task.State,
	})
}

func (q *Queue) complete(state *taskState) {
	q.setState(state, "done")
	q.emitter.Emit("transfer:done", DoneEvent{
		TaskID:     state.task.ID,
		SessionID:  state.task.SessionID,
		LocalPath:  state.task.LocalPath,
		RemotePath: state.task.RemotePath,
		Direction:  state.task.Direction,
	})
}

func (q *Queue) fail(state *taskState, err error) {
	q.setState(state, "error")
	q.emitter.Emit("transfer:error", ErrorEvent{
		TaskID:     state.task.ID,
		SessionID:  state.task.SessionID,
		LocalPath:  state.task.LocalPath,
		RemotePath: state.task.RemotePath,
		Direction:  state.task.Direction,
		Message:    err.Error(),
	})
}
