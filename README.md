# GoTerm

GoTerm is a cross-platform desktop SSH/SFTP client built with Wails v2 (Go backend + WebView frontend). It focuses on fast terminal interaction, file browsing, and secure credential handling.

## Features

- Host profiles (add/edit/delete, grouping, tags)
- SSH connect/disconnect with host key verification
- Interactive terminal (xterm.js) with resize support
- SFTP file browsing (list/stat/mkdir/remove/rename)
- Download/upload tasks with progress events
- Credential storage via OS keyring (no plain-text passwords)

## Tech Stack

- Desktop shell: Wails v2
- Frontend: Vue 3 + Vite
- UI: Naive UI or Element Plus
- Terminal: xterm.js
- Backend: Go
- SSH: golang.org/x/crypto/ssh
- SFTP: github.com/pkg/sftp
- Storage: SQLite (modernc.org/sqlite)
- Keyring: github.com/zalando/go-keyring

## Architecture

UI (Vue)
- Terminal view (xterm.js)
- File manager (tree/list/drag/drop)
- Connection management (hosts/groups/tags)
- Transfer task list (queue/progress/cancel)

Go App Layer
- Profiles
- Session manager (SSH reuse/reconnect)
- Terminal gateway (PTY, IO streaming)
- FileService (SFTP operations)
- TransferService (upload/download with progress)
- Keyring
- HostKey (known_hosts verification)

SSH/SFTP
- Remote servers

Core idea: terminal uses streaming events, file ops use RPC, transfer uses tasks + event push.

## Project Structure

- `frontend/` Vue 3 + Vite
- `backend/` Go services and Wails bindings
- `build/` Wails build output
- `wails.json` Wails project config

## Development

Prerequisites
- Go 1.22+
- Node.js 18+
- Wails v2

Commands

```bash
# install frontend deps
pnpm i
# or npm i

# start dev mode
wails dev

# build for current platform
wails build
```

## Security Notes

- known_hosts policy: ask (default), strict, accept-new
- credentials are stored in OS keyring; database stores only key references

## Roadmap

MVP
- Host profiles CRUD
- SSH connect/disconnect
- Terminal open/write/resize
- SFTP list/stat
- Download with progress
- Host key prompt and persistence

Enhancements
- Upload + drag and drop
- Multi-tab sessions
- Search/favorites/recent
- Jump host (bastion) and ProxyCommand
- File preview and permission tools

## More Design Notes

See `文档.md` for detailed architecture, API drafts, and implementation notes.
