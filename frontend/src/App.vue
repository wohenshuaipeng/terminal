<template>
  <div class="shell theme-commercial min-h-screen">
    <aside class="sidebar">
      <div class="brand">
        <p class="eyebrow">GoTerm</p>
        <p class="sidebar-title">Ops Hub</p>
      </div>
      <nav class="sidebar-nav">
        <button
          v-for="item in sidebarItems"
          :key="item.id"
          class="nav-icon"
          :class="{ active: activeTab === item.id }"
          :title="item.label"
          type="button"
          @click="openTab(item.id)"
        >
          <el-icon>
            <component :is="item.icon" />
          </el-icon>
          <span class="sr-only">{{ item.label }}</span>
        </button>
      </nav>
      <div class="sidebar-footer">
        <span :class="['dot', backendReady ? 'ok' : 'warn']"></span>
      </div>
    </aside>

    <div class="grid">
      <div class="appbar">
        <div>
          <p class="appbar-title">Workspace</p>
          <p class="muted">Operational overview and quick access.</p>
        </div>
        <div class="appbar-actions">
          <el-button plain size="small" @click="reloadProfiles" :disabled="!backendReady">
            Sync
          </el-button>
          <el-button type="primary" size="small" @click="reloadMySQLProfiles" :disabled="!backendReady">
            Refresh DB
          </el-button>
          <span class="appbar-pill">{{ backendReady ? "Online" : "Offline" }}</span>
        </div>
      </div>
      <el-tabs class="workspace-tabs" v-model="activeTab" type="card" @tab-remove="closeTab">
        <el-tab-pane
          v-for="tab in openTabs"
          :key="tab.id"
          :name="tab.id"
          closable
        >
          <template #label>
            <span class="tab-label">
              <el-icon><component :is="tab.icon" /></el-icon>
              <span>{{ tab.label }}</span>
            </span>
          </template>
        </el-tab-pane>
      </el-tabs>
      <section id="profiles" class="panel profiles" v-show="activeTab === 'profiles'">
        <div class="panel-head">
          <h2>Profiles</h2>
          <el-button type="primary" size="small" @click="newProfile">New</el-button>
        </div>
        <div v-if="loading" class="muted">Loading...</div>
        <div v-else-if="profiles.length === 0" class="muted">No profiles yet.</div>
        <ul class="profile-list" v-else>
          <li
            v-for="profile in profiles"
            :key="profile.id"
            :class="{ active: profile.id === form.id }"
          >
            <div class="profile-main" @click="editProfile(profile)">
              <div class="profile-title">
                <span class="name">{{ profile.name || profile.host }}</span>
                <span v-if="profile.group" class="tag">{{ profile.group }}</span>
              </div>
              <div class="profile-meta">
                {{ profile.username }}@{{ profile.host }}:{{ profile.port || 22 }}
              </div>
              <div class="profile-status">
                <span :class="['dot', statusClass(profile.id)]"></span>
                <span>{{ statusLabel(profile.id) }}</span>
              </div>
            </div>
            <div class="profile-actions">
              <el-button
                type="primary"
                size="small"
                @click.stop="connectProfile(profile)"
                :disabled="!backendReady || statusLabel(profile.id) === 'connected'"
              >
                Connect
              </el-button>
              <el-button
                plain
                size="small"
                @click.stop="disconnectProfile(profile)"
                :disabled="!backendReady || !sessionByProfile[profile.id]"
              >
                Disconnect
              </el-button>
              <el-button
                plain
                size="small"
                @click.stop="openTerminalForProfile(profile)"
                :disabled="!backendReady"
              >
                Terminal
              </el-button>
              <el-button plain size="small" @click.stop="editProfile(profile)">Edit</el-button>
              <el-button type="danger" size="small" @click.stop="deleteProfile(profile)">Delete</el-button>
            </div>
          </li>
        </ul>
      </section>

      <section id="editor" class="panel editor" v-show="activeTab === 'profiles'">
        <div class="panel-head">
          <h2>{{ form.id ? 'Edit profile' : 'New profile' }}</h2>
          <el-button v-if="form.id" plain size="small" @click="clearCredentials">
            Clear credentials
          </el-button>
        </div>
        <el-form class="form" label-position="top">
          <el-form-item label="Name">
            <el-input v-model="form.name" placeholder="prod-1" />
          </el-form-item>
          <el-form-item label="Group">
            <el-input v-model="form.group" placeholder="Production" />
          </el-form-item>
          <el-form-item label="Host">
            <el-input v-model="form.host" placeholder="1.2.3.4" />
          </el-form-item>
          <el-form-item label="Port">
            <el-input-number v-model="form.port" :min="1" :max="65535" :controls="false" />
          </el-form-item>
          <el-form-item label="Username">
            <el-input v-model="form.username" placeholder="root" />
          </el-form-item>
          <el-form-item label="Auth type">
            <el-select v-model="form.authType" placeholder="Select auth">
              <el-option label="Password" value="password" />
              <el-option label="Private key" value="privateKey" />
            </el-select>
          </el-form-item>
          <el-form-item v-if="form.authType === 'privateKey'" label="Private key path">
            <el-input v-model="form.privateKeyPath" placeholder="C:\\Users\\me\\.ssh\\id_ed25519" />
          </el-form-item>
          <el-form-item label="Use system keyring">
            <el-switch v-model="form.useKeyring" />
          </el-form-item>
          <el-form-item label="Known hosts policy">
            <el-select v-model="form.knownHostsPolicy" placeholder="Select policy">
              <el-option label="Ask" value="ask" />
              <el-option label="Strict" value="strict" />
              <el-option label="Accept new" value="accept-new" />
            </el-select>
          </el-form-item>
          <el-form-item v-if="form.useKeyring && form.authType === 'password'" label="Password">
            <el-input v-model="form.password" type="password" show-password placeholder="Stored in keyring" />
          </el-form-item>
          <el-form-item v-if="form.useKeyring && form.authType === 'privateKey'" label="Key passphrase">
            <el-input v-model="form.passphrase" type="password" show-password placeholder="Stored in keyring" />
          </el-form-item>
          <div class="form-actions">
            <el-button type="primary" @click="saveProfile" :disabled="!backendReady">
              Save profile
            </el-button>
            <el-button plain @click="newProfile">Reset</el-button>
          </div>
        </el-form>
        <p v-if="error" class="error">{{ error }}</p>
      </section>

      <section id="terminal" class="panel terminal" v-show="activeTab === 'terminal'">
        <div class="panel-head">
          <h2>Terminal</h2>
          <div class="panel-actions">
            <el-select
              v-model="terminalSessionId"
              placeholder="Select session"
              :disabled="!backendReady || connectedProfiles.length === 0"
            >
              <el-option
                v-for="item in connectedProfiles"
                :key="item.sessionId"
                :label="item.label"
                :value="item.sessionId"
              />
            </el-select>
            <el-button type="primary" size="small" @click="openTerminal" :disabled="!terminalSessionId">
              Open
            </el-button>
            <el-button plain size="small" @click="quickPanelVisible = true">
              Quick Ops
            </el-button>
          </div>
        </div>
        <div class="terminal-layout">
          <div class="terminal-main">
            <div class="terminal-search" v-if="terminalSearchVisible">
              <el-input
                ref="terminalSearchInput"
                v-model="terminalSearchQuery"
                size="small"
                clearable
                placeholder="Search in terminal"
                @keydown="handleTerminalSearchKey"
              />
              <el-button plain size="small" @click="findTerminalPrev" :disabled="!terminalSearchQuery">
                Prev
              </el-button>
              <el-button plain size="small" @click="findTerminalNext" :disabled="!terminalSearchQuery">
                Next
              </el-button>
              <el-button plain size="small" @click="closeTerminalSearch">Close</el-button>
            </div>
            <div class="terminal-toolbar">
              <el-button plain size="small" @click="copySelection" :disabled="!activeTermId">Copy</el-button>
              <el-button plain size="small" @click="pasteClipboard" :disabled="!activeTermId">Paste</el-button>
              <el-button plain size="small" @click="clearTerminal" :disabled="!activeTermId">Clear</el-button>
              <div class="terminal-font">
                <span class="muted">Font</span>
                <el-button plain size="small" @click="setTerminalFontSize(terminalFontSize - 1)">-</el-button>
                <span class="font-value">{{ terminalFontSize }}</span>
                <el-button plain size="small" @click="setTerminalFontSize(terminalFontSize + 1)">+</el-button>
              </div>
              <el-button plain size="small" @click="fitTerminal(activeTermId)" :disabled="!activeTermId">
                Fit
              </el-button>
            </div>
            <div v-if="terminals.length" class="tab-list">
              <el-button
                v-for="term in terminals"
                :key="term.id"
                class="tab"
                text
                :class="{ active: term.id === activeTermId }"
                @click="activeTermId = term.id"
              >
                <span>{{ term.title }}</span>
                <span class="tab-close" @click.stop="closeTerminal(term.id)">x</span>
              </el-button>
            </div>
            <div class="terminal-stage">
              <div v-if="terminals.length === 0" class="muted">No terminal open.</div>
              <div
                v-for="term in terminals"
                :key="term.id"
                class="terminal-shell"
                v-show="term.id === activeTermId"
              >
                <div
                  class="terminal-canvas"
                  :ref="(el) => setTermRef(term.id, el)"
                ></div>
              </div>
            </div>
            <div class="metrics-inline" v-if="activeTermId">
              <span v-if="metricsError" class="error">{{ metricsError }}</span>
              <template v-else-if="systemStats">
                <div class="metric-inline">
                  <span class="metric-inline-label">CPU</span>
                  <span class="metric-inline-value">{{ cpuTotal.toFixed(1) }}%</span>
                  <svg class="inline-sparkline" viewBox="0 0 160 40" aria-hidden="true">
                    <polyline :points="cpuSparkline" />
                  </svg>
                </div>
                <div class="metric-inline">
                  <span class="metric-inline-label">Memory</span>
                  <span class="metric-inline-value">{{ memUsedPercent.toFixed(1) }}%</span>
                  <span class="metric-inline-sub">
                    {{ formatBytes(memUsed) }} / {{ formatBytes(memTotal) }}
                  </span>
                  <svg class="inline-sparkline" viewBox="0 0 160 40" aria-hidden="true">
                    <polyline :points="memSparkline" />
                  </svg>
                </div>
              </template>
              <span v-else class="muted">Loading metrics...</span>
            </div>
            <div class="files terminal-files" ref="dropZoneRef">
              <div class="panel-head">
                <h3>Files</h3>
                <div class="panel-actions">
                  <el-select
                    v-model="filesSessionId"
                    placeholder="Select session"
                    :disabled="!backendReady || connectedProfiles.length === 0"
                  >
                    <el-option
                      v-for="item in connectedProfiles"
                      :key="item.sessionId"
                      :label="item.label"
                      :value="item.sessionId"
                    />
                  </el-select>
                  <el-button plain size="small" @click="reloadFiles" :disabled="!filesSessionId">
                    Reload
                  </el-button>
                </div>
              </div>
              <div class="files-toolbar">
                <el-input
                  v-model="filesPath"
                  class="path-input"
                  placeholder="/var/www"
                  @keyup.enter="loadFiles(filesPath)"
                />
                <el-button plain size="small" @click="loadFiles(filesPath)" :disabled="!filesSessionId">
                  Load
                </el-button>
                <el-button plain size="small" @click="goUp" :disabled="!filesSessionId">
                  Up
                </el-button>
              </div>
              <div class="file-list">
                <div class="file-row header">
                  <span class="sortable" @click="toggleFileSort('name')">
                    Name
                    <span class="sort-indicator">{{ fileSortIndicator('name') }}</span>
                  </span>
                  <span>Perm</span>
                  <span>Size</span>
                  <span class="sortable" @click="toggleFileSort('mtime')">
                    Modified
                    <span class="sort-indicator">{{ fileSortIndicator('mtime') }}</span>
                  </span>
                  <span></span>
                </div>
                <div v-if="filesLoading" class="muted">Loading...</div>
                <div v-else-if="filesEntries.length === 0" class="muted">No entries.</div>
                <template v-for="entry in fileSortedEntries" :key="entry.path">
                  <div
                    class="file-row"
                    :class="{ active: selectedEntry && selectedEntry.path === entry.path }"
                    @click="selectEntry(entry)"
                    @dblclick="openEntry(entry)"
                    @contextmenu="openFileContextMenu(entry, $event)"
                  >
                    <span class="file-name">
                      <span class="pill">{{ entry.isDir ? '[DIR]' : '[FILE]' }}</span>
                      {{ entry.name }}
                    </span>
                    <span class="file-mode">{{ entry.mode }}</span>
                    <span>{{ formatFileSize(entry.size) }}</span>
                    <span>{{ formatTime(entry.mtime) }}</span>
                    <span class="file-actions-inline">
                      <el-button
                        v-if="entry.isDir"
                        text
                        size="small"
                        @click.stop="openEntry(entry)"
                      >
                        Open
                      </el-button>
                      <el-button
                        v-if="!entry.isDir"
                        text
                        size="small"
                        @click.stop="downloadEntry(entry)"
                      >
                        Download
                      </el-button>
                      <el-button
                        text
                        size="small"
                        type="danger"
                        @click.stop="deleteEntry(entry)"
                      >
                        Delete
                      </el-button>
                    </span>
                  </div>
                  <div
                    v-if="isActiveTransfer(fileDownloadTransfers[entry.path])"
                    class="file-progress"
                  >
                    <div class="file-progress-bar">
                      <span
                        class="file-progress-fill"
                        :style="{ width: `${transferPercent(fileDownloadTransfers[entry.path]).toFixed(1)}%` }"
                      ></span>
                    </div>
                    <div class="file-progress-meta">
                      <span>
                        {{ transferPercent(fileDownloadTransfers[entry.path]).toFixed(1) }}%
                        ({{ formatBytes(fileDownloadTransfers[entry.path].doneBytes) }} /
                        {{ formatBytes(fileDownloadTransfers[entry.path].totalBytes) }})
                      </span>
                      <span class="muted">
                        {{ fileDownloadTransfers[entry.path].state }}
                      </span>
                    </div>
                    <div v-if="fileDownloadTransfers[entry.path].message" class="error">
                      {{ fileDownloadTransfers[entry.path].message }}
                    </div>
                  </div>
                </template>
              </div>
              <div
                v-if="fileContextMenu.visible"
                class="file-menu"
                :style="{ top: `${fileContextMenu.y}px`, left: `${fileContextMenu.x}px` }"
              >
                <button
                  class="file-menu-item"
                  type="button"
                  @click="downloadEntry(fileContextMenu.entry); closeFileContextMenu()"
                  :disabled="!fileContextMenu.entry || fileContextMenu.entry.isDir"
                >
                  Download
                </button>
                <button
                  class="file-menu-item"
                  type="button"
                  @click="renameEntry(fileContextMenu.entry); closeFileContextMenu()"
                  :disabled="!fileContextMenu.entry"
                >
                  Rename
                </button>
                <button
                  class="file-menu-item danger"
                  type="button"
                  @click="deleteEntry(fileContextMenu.entry); closeFileContextMenu()"
                  :disabled="!fileContextMenu.entry"
                >
                  Delete
                </button>
              </div>
              <div class="file-actions">
                <el-input
                  v-model="newFolderName"
                  placeholder="New folder name"
                  :disabled="!filesSessionId"
                />
                <el-button plain size="small" @click="createFolder" :disabled="!filesSessionId || !newFolderName">
                  Mkdir
                </el-button>
                <el-input
                  v-model="renameName"
                  placeholder="Rename selected"
                  :disabled="!filesSessionId || !selectedEntry"
                />
                <el-button
                  plain
                  size="small"
                  @click="renameSelected"
                  :disabled="!filesSessionId || !selectedEntry || !renameName"
                >
                  Rename
                </el-button>
                <el-button
                  type="danger"
                  size="small"
                  @click="removeSelected"
                  :disabled="!filesSessionId || !selectedEntry"
                >
                  Delete
                </el-button>
              </div>
              <div class="upload-bar">
                <el-input v-model="uploadLocalPath" placeholder="Local file path" :disabled="!filesSessionId" />
                <el-button plain size="small" @click="browseUploadLocal" :disabled="!filesSessionId">
                  Browse
                </el-button>
                <el-input v-model="uploadRemotePath" placeholder="Remote path" :disabled="!filesSessionId" />
                <el-button
                  type="primary"
                  size="small"
                  @click="uploadFile"
                  :disabled="!filesSessionId || !uploadLocalPath || !uploadRemotePath"
                >
                  Upload
                </el-button>
              </div>
              <div class="download-bar">
                <el-input
                  v-model="downloadPath"
                  placeholder="Local path for download"
                  :disabled="!filesSessionId"
                />
                <el-button
                  plain
                  size="small"
                  @click="browseDownloadPath"
                  :disabled="!filesSessionId || !selectedEntry || selectedEntry.isDir"
                >
                  Browse
                </el-button>
                <el-button
                  type="primary"
                  size="small"
                  @click="downloadSelected"
                  :disabled="!filesSessionId || !selectedEntry || selectedEntry.isDir"
                >
                  Download
                </el-button>
              </div>
              <div class="drop-hint muted">Drop files here to upload.</div>
            </div>
          </div>
        </div>
      </section>

      <section id="mysql" class="panel mysql" v-show="activeTab === 'mysql'">
        <div class="panel-head">
          <h2>MySQL</h2>
          <div class="panel-actions">
            <el-button plain size="small" @click="reloadMySQLProfiles" :disabled="!backendReady">
              Reload
            </el-button>
            <el-button type="primary" size="small" @click="showMySQLDialog()">Add</el-button>
          </div>
        </div>
        <div v-if="mysqlError" class="error">{{ mysqlError }}</div>
        <div class="mysql-grid">
          <div class="mysql-profiles">
            <h3>Connections</h3>
            <div v-if="mysqlLoading" class="muted">Loading...</div>
            <div v-else-if="mysqlProfiles.length === 0" class="muted">No MySQL profiles yet.</div>
            <ul class="profile-list" v-else>
              <li
                v-for="profile in mysqlProfiles"
                :key="profile.id"
                :class="{ active: profile.id === mysqlForm.id }"
              >
                <div class="profile-main" @click="showMySQLDialog(profile)">
                  <div class="profile-title">
                    <span class="name">{{ profile.name || profile.host }}</span>
                    <span class="tag">{{ mysqlConnectionLabel(profile) }}</span>
                  </div>
                  <div class="profile-meta">
                    {{ profile.username }}@{{ profile.host }}:{{ profile.port || 3306 }}
                  </div>
                  <div class="profile-status">
                    <span :class="['dot', mysqlStatusClass(profile.id)]"></span>
                    <span>{{ mysqlStatusLabel(profile.id) }}</span>
                  </div>
                </div>
                <div class="profile-actions">
                  <el-button
                    type="primary"
                    size="small"
                    @click.stop="connectMySQLProfile(profile)"
                    :disabled="!backendReady"
                  >
                    {{ mysqlStatusLabel(profile.id) === "connected" ? "Open" : "Connect" }}
                  </el-button>
                  <el-button
                    plain
                    size="small"
                    @click.stop="disconnectMySQLProfile(profile)"
                    :disabled="!backendReady || mysqlStatusLabel(profile.id) !== 'connected'"
                  >
                    Disconnect
                  </el-button>
                  <el-button plain size="small" @click.stop="showMySQLDialog(profile)">Edit</el-button>
                  <el-button type="danger" size="small" @click.stop="deleteMySQLProfile(profile)">
                    Delete
                  </el-button>
                </div>
              </li>
            </ul>
          </div>
          <div class="mysql-explorer">
            <el-tabs
              class="mysql-tabs"
              v-model="activeMySQLTab"
              type="card"
              @tab-remove="closeMySQLTab"
            >
              <el-tab-pane
                v-for="tab in mysqlTabs"
                :key="tab.id"
                :name="tab.id"
                closable
              >
                <template #label>
                  <span class="tab-label">
                    <el-icon><DataLine /></el-icon>
                    <span>{{ tab.label }}</span>
                  </span>
                </template>
              </el-tab-pane>
            </el-tabs>
            <div class="mysql-explorer-head">
              <h3>Explorer</h3>
              <div class="panel-actions">
                <span class="muted" v-if="!mysqlState.profileId">No MySQL tab open.</span>
                <el-button plain size="small" @click="loadMySQLDatabases()" :disabled="!mysqlState.profileId">
                  Refresh
                </el-button>
              </div>
            </div>
              <div class="mysql-actions">
                <el-input v-model="mysqlState.newDatabase" placeholder="New database name" />
                <el-button
                  plain
                  size="small"
                  @click="createMySQLDatabase"
                  :disabled="!mysqlState.profileId || !mysqlState.newDatabase"
                >
                  Create DB
                </el-button>
                <el-button
                  type="danger"
                  size="small"
                  @click="dropMySQLDatabase"
                  :disabled="!mysqlState.profileId || !mysqlState.activeDatabase"
                >
                  Drop DB
                </el-button>
              </div>
            <div class="mysql-list-grid">
              <div>
                <div class="mysql-list-head">
                  <h4>Databases</h4>
                  <div class="panel-actions">
                    <span v-if="mysqlDefaultDatabase" class="muted">
                      Default: {{ mysqlDefaultDatabase }}
                    </span>
                    <el-button
                      plain
                      size="small"
                      @click="saveMySQLDefaultDatabase"
                      :disabled="!mysqlState.profileId || !mysqlState.activeDatabase"
                    >
                      Set Default
                    </el-button>
                  </div>
                </div>
                <ul
                  ref="mysqlTableListRef"
                  class="mysql-list"
                  tabindex="0"
                  @keydown="handleMySQLTableKeydown"
                  @click="focusMySQLTableList"
                >
                  <li
                    v-for="name in mysqlState.databases"
                    :key="name"
                    :class="{ active: name === mysqlState.activeDatabase }"
                    @click="selectMySQLDatabase(name)"
                  >
                    {{ name }}
                  </li>
                </ul>
              </div>
              <div>
                <div class="mysql-list-head">
                  <h4>Tables</h4>
                  <el-input
                    v-model="mysqlState.tableSearch"
                    size="small"
                    placeholder="Search tables"
                    clearable
                  />
                </div>
                <ul class="mysql-list">
                  <li
                    v-for="(name, idx) in mysqlFilteredTables"
                    :key="name"
                    :data-index="idx"
                    :class="{ active: name === mysqlState.activeTable }"
                    @click="selectMySQLTable(name)"
                  >
                    {{ name }}
                  </li>
                </ul>
                <el-button
                  type="danger"
                  size="small"
                  @click="dropMySQLTable"
                  :disabled="!mysqlState.profileId || !mysqlState.activeDatabase || !mysqlState.activeTable"
                >
                  Drop table
                </el-button>
              </div>
            </div>
          </div>
        </div>
        <el-dialog v-model="mysqlDialogVisible" :title="mysqlDialogTitle" width="520px">
          <el-form class="form" label-position="top">
            <el-form-item label="Name">
              <el-input v-model="mysqlForm.name" placeholder="prod-db" />
            </el-form-item>
            <el-form-item label="Host">
              <el-input v-model="mysqlForm.host" placeholder="127.0.0.1" />
            </el-form-item>
            <el-form-item label="Port">
              <el-input-number v-model="mysqlForm.port" :min="1" :max="65535" :controls="false" />
            </el-form-item>
            <el-form-item label="Username">
              <el-input v-model="mysqlForm.username" placeholder="root" />
            </el-form-item>
            <el-form-item label="Default database">
              <el-input v-model="mysqlForm.database" placeholder="mydb" />
            </el-form-item>
            <el-form-item label="Connection">
              <el-select v-model="mysqlForm.connectionType" placeholder="Connection type">
                <el-option label="Direct" value="direct" />
                <el-option label="SSH tunnel" value="ssh" />
              </el-select>
            </el-form-item>
            <el-form-item v-if="mysqlForm.connectionType === 'ssh'" label="SSH profile">
              <el-select v-model="mysqlForm.sshProfileId" placeholder="Select SSH profile">
                <el-option
                  v-for="item in profiles"
                  :key="item.id"
                  :label="item.name || item.host"
                  :value="item.id"
                />
              </el-select>
            </el-form-item>
            <el-form-item label="Use system keyring">
              <el-switch v-model="mysqlForm.useKeyring" />
            </el-form-item>
            <el-form-item v-if="mysqlForm.useKeyring" label="Password">
              <el-input v-model="mysqlForm.password" type="password" show-password placeholder="Stored in keyring" />
            </el-form-item>
            <el-form-item label="Enable TLS">
              <el-switch v-model="mysqlForm.useTls" />
            </el-form-item>
            <el-form-item v-if="mysqlForm.useTls" label="TLS CA path">
              <el-input v-model="mysqlForm.tlsCaPath" placeholder="C:\\certs\\ca.pem" />
            </el-form-item>
            <el-form-item v-if="mysqlForm.useTls" label="TLS client cert">
              <el-input v-model="mysqlForm.tlsCertPath" placeholder="C:\\certs\\client.pem" />
            </el-form-item>
            <el-form-item v-if="mysqlForm.useTls" label="TLS client key">
              <el-input v-model="mysqlForm.tlsKeyPath" placeholder="C:\\certs\\client-key.pem" />
            </el-form-item>
            <el-form-item v-if="mysqlForm.useTls" label="Skip TLS verify">
              <el-switch v-model="mysqlForm.tlsSkipVerify" />
            </el-form-item>
            <div class="form-actions">
              <el-button type="primary" @click="saveMySQLProfile" :disabled="!backendReady">
                Save profile
              </el-button>
              <el-button plain @click="newMySQLProfile">Reset</el-button>
            </div>
          </el-form>
        </el-dialog>
        <el-dialog
          v-model="mysqlState.schemaDialogVisible"
          :title="mysqlState.schemaMode === 'edit' ? 'Edit column' : 'Add column'"
          width="520px"
        >
          <el-form class="form" label-position="top">
            <el-form-item label="Name">
              <el-input v-model="mysqlState.schemaForm.name" placeholder="column_name" />
            </el-form-item>
            <el-form-item label="Type">
              <el-input v-model="mysqlState.schemaForm.type" placeholder="VARCHAR(255)" />
            </el-form-item>
            <el-form-item label="Nullable">
              <el-switch v-model="mysqlState.schemaForm.nullable" />
            </el-form-item>
            <el-form-item label="Default (raw)">
              <el-input v-model="mysqlState.schemaForm.defaultValue" placeholder="NULL / 0 / 'text'" />
            </el-form-item>
            <el-form-item label="Extra / Options">
              <el-input v-model="mysqlState.schemaForm.extra" placeholder="AUTO_INCREMENT, COMMENT '...'" />
            </el-form-item>
            <p v-if="mysqlState.schemaError" class="error">{{ mysqlState.schemaError }}</p>
            <div class="form-actions">
              <el-button type="primary" @click="saveMySQLSchemaChange">Apply</el-button>
              <el-button plain @click="mysqlState.schemaDialogVisible = false">Cancel</el-button>
            </div>
          </el-form>
        </el-dialog>
          <div class="mysql-preview">
            <div class="preview-head">
              <span>Table preview</span>
              <div class="panel-actions">
                <span v-if="mysqlState.preview && mysqlState.preview.truncated" class="muted">truncated</span>
                <el-button
                  type="primary"
                  size="small"
                  @click="submitMySQLEdits"
                  :disabled="
                    !mysqlState.profileId ||
                    !mysqlState.activeTable ||
                    !mysqlHasEdits ||
                    mysqlState.editSubmitting
                  "
                >
                  {{ mysqlState.editSubmitting ? "Submitting..." : "Submit changes" }}
                </el-button>
              </div>
            </div>
            <div class="mysql-preview-controls">
              <el-input v-model="mysqlState.filter" placeholder="Filter (e.g. status = 1)" />
              <el-input-number v-model="mysqlState.limit" :min="1" :controls="false" placeholder="Limit" />
              <el-input-number v-model="mysqlState.offset" :min="0" :controls="false" placeholder="Offset" />
                <el-button
                  plain
                  size="small"
                  @click="loadMySQLPreview()"
                  :disabled="!mysqlState.profileId || !mysqlState.activeDatabase || !mysqlState.activeTable"
                >
                  Refresh
                </el-button>
            </div>
          <div v-if="mysqlState.previewLoading" class="muted">Loading preview...</div>
          <div v-else-if="mysqlState.previewError" class="error">{{ mysqlState.previewError }}</div>
          <div v-else-if="mysqlState.preview && mysqlState.preview.columns && mysqlState.preview.columns.length">
            <div class="mysql-table-wrap">
              <table class="mysql-table">
                <thead>
                  <tr>
                    <th
                      v-for="col in mysqlState.preview.columns"
                      :key="col"
                      class="sortable"
                      @click="toggleMySQLSort(col)"
                    >
                      <span>{{ col }}</span>
                      <span v-if="mysqlState.sortColumn === col" class="sort-indicator">
                        {{ mysqlState.sortDirection === "asc" ? "ASC" : "DESC" }}
                      </span>
                    </th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="(row, idx) in mysqlState.preview.rows" :key="idx">
                    <td
                      v-for="(cell, cidx) in row"
                      :key="cidx"
                      :class="{
                        edited:
                          mysqlState.rowEdits[idx] &&
                          mysqlState.rowEdits[idx][mysqlState.preview.columns[cidx]] !== undefined
                      }"
                      @click="startEditCell(idx, cidx)"
                    >
                      <el-input
                        v-if="isEditingCell(idx, cidx)"
                        v-model="mysqlState.preview.rows[idx][cidx]"
                        size="small"
                        @blur="commitEditCell(idx, cidx)"
                        @keyup.enter="commitEditCell(idx, cidx)"
                      />
                      <span v-else>{{ cell }}</span>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
          <div v-else class="muted">Select a table to preview.</div>
        </div>
        <div class="mysql-schema">
          <div class="preview-head">
            <span>Schema</span>
            <div class="panel-actions">
                <el-button
                  plain
                  size="small"
                  @click="loadMySQLSchema()"
                  :disabled="!mysqlState.profileId || !mysqlState.activeTable"
                >
                  Refresh
                </el-button>
              <el-button
                type="primary"
                size="small"
                @click="openMySQLSchemaDialog('add')"
                :disabled="!mysqlState.profileId || !mysqlState.activeTable"
              >
                Add column
              </el-button>
            </div>
          </div>
          <div v-if="!mysqlState.activeTable" class="muted">Select a table to view schema.</div>
          <div v-else-if="mysqlState.schemaLoading" class="muted">Loading schema...</div>
          <div v-else-if="mysqlState.schemaError" class="error">{{ mysqlState.schemaError }}</div>
          <div v-else-if="mysqlState.schemaColumns.length">
            <div class="mysql-table-wrap">
              <table class="mysql-table">
                <thead>
                  <tr>
                    <th>Name</th>
                    <th>Type</th>
                    <th>Null</th>
                    <th>Key</th>
                    <th>Default</th>
                    <th>Extra</th>
                    <th></th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="col in mysqlState.schemaColumns" :key="col.name">
                    <td>{{ col.name }}</td>
                    <td>{{ col.type }}</td>
                    <td>{{ col.nullable }}</td>
                    <td>{{ col.key }}</td>
                    <td>{{ col.default }}</td>
                    <td>{{ col.extra }}</td>
                    <td class="schema-actions">
                      <el-button text size="small" @click="openMySQLSchemaDialog('edit', col)">
                        Edit
                      </el-button>
                      <el-button text size="small" type="danger" @click="dropMySQLColumn(col)">
                        Drop
                      </el-button>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
          <div v-else class="muted">No columns found.</div>
        </div>
          <div class="mysql-console">
            <div class="preview-head">
              <span>SQL Console</span>
              <div class="panel-actions">
              <el-select
                v-model="mysqlState.queryDatabase"
                placeholder="Use default"
                :disabled="!mysqlState.profileId || mysqlState.databases.length === 0"
              >
                <el-option label="Use default" value="" />
                <el-option v-for="name in mysqlState.databases" :key="name" :label="name" :value="name" />
              </el-select>
              <el-button
                type="primary"
                size="small"
                @click="runMySQLQuery"
                :disabled="!mysqlState.profileId || mysqlState.queryRunning"
              >
                Run
              </el-button>
              </div>
            </div>
          <div class="sql-editor">
            <el-input
              ref="mysqlQueryInputRef"
              v-model="mysqlState.queryText"
              type="textarea"
              :rows="6"
              placeholder="SELECT * FROM users LIMIT 50;"
              @keyup="handleMySQLQueryKeyup"
              @keydown="handleMySQLQueryKeydown"
              @focus="handleMySQLQueryFocus"
              @blur="handleMySQLQueryBlur"
            />
            <div v-if="mysqlQueryHints.length" class="sql-hints sql-hints-overlay">
              <button
                v-for="(hint, idx) in mysqlQueryHints"
                :key="hint"
                type="button"
                class="sql-hint"
                :class="{ active: idx === mysqlState.queryHintIndex }"
                @mousedown.prevent="applyMySQLHint(hint)"
              >
                {{ hint }}
              </button>
            </div>
          </div>
          <div v-if="mysqlState.queryRunning" class="muted">Running query...</div>
          <div v-else-if="mysqlState.queryError" class="error">{{ mysqlState.queryError }}</div>
          <div v-else-if="mysqlState.queryResult" class="mysql-query-result">
            <div v-if="mysqlState.queryResult.kind === 'exec'" class="muted">
              Affected rows: {{ mysqlState.queryResult.affectedRows }}, Last insert id:
              {{ mysqlState.queryResult.lastInsertId }}
            </div>
            <div v-else class="muted">
              Returned {{ mysqlState.queryResult.rows.length }} rows in
              {{ mysqlState.queryResult.durationMs }}ms
              <span v-if="mysqlState.queryResult.message"> ({{ mysqlState.queryResult.message }})</span>
            </div>
            <div
              v-if="mysqlState.queryResult.kind === 'rows' && mysqlState.queryResult.columns.length"
              class="mysql-table-wrap"
            >
              <table class="mysql-table">
                <thead>
                  <tr>
                    <th v-for="col in mysqlState.queryResult.columns" :key="col">{{ col }}</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="(row, idx) in mysqlState.queryResult.rows" :key="idx">
                    <td v-for="(cell, cidx) in row" :key="cidx">{{ cell }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>
      </section>

      <section id="activity" class="panel events" v-show="activeTab === 'activity'">
        <div class="panel-head">
          <h2>Events</h2>
          <el-button plain size="small" @click="clearEvents">Clear</el-button>
        </div>
        <div v-if="transfers.length" class="transfer-block">
          <h3>Transfers</h3>
          <div v-for="task in transfers" :key="task.id" class="transfer-row">
            <div>
              <div class="transfer-title">{{ task.remotePath }} -> {{ task.localPath }}</div>
              <div class="muted">{{ transferProgress(task) }}</div>
              <div v-if="task.message" class="muted">{{ task.message }}</div>
            </div>
            <div class="transfer-actions">
              <el-button
                plain
                size="small"
                @click="cancelTransfer(task)"
                :disabled="!(task.state === 'queued' || task.state === 'running')"
              >
                Cancel
              </el-button>
              <el-button
                plain
                size="small"
                @click="retryTransfer(task)"
                :disabled="!(task.state === 'error' || task.state === 'done')"
              >
                Retry
              </el-button>
              <span class="chip" :class="task.state">{{ task.state }}</span>
            </div>
          </div>
        </div>
        <div class="event-list" v-if="events.length">
          <div v-for="event in events" :key="event.id" class="event-row">
            <div class="event-meta">{{ event.time }} - {{ event.type }}</div>
            <div class="event-payload">{{ event.payload }}</div>
          </div>
        </div>
        <div v-else class="muted">No events yet.</div>
      </section>
    </div>

    <div v-if="hostKeyPrompt" class="modal">
      <div class="modal-card">
        <h3>Unknown host key</h3>
        <p class="muted">Host: {{ hostKeyPrompt.host }}</p>
        <p class="muted">Fingerprint: {{ hostKeyPrompt.fingerprint }}</p>
        <div class="form-actions">
          <el-button plain @click="respondHostKey(false)">Reject</el-button>
          <el-button type="primary" @click="respondHostKey(true)">Allow</el-button>
        </div>
      </div>
    </div>
    <div v-if="quickPanelVisible" class="modal" @click.self="quickPanelVisible = false">
      <div class="modal-card modal-card-wide">
        <div class="modal-head">
          <h3>Quick Ops</h3>
          <el-button text size="small" @click="quickPanelVisible = false">Close</el-button>
        </div>
        <div class="terminal-quick">
          <div class="terminal-quick-row">
            <span class="terminal-quick-title">Systemctl</span>
            <el-select v-model="quickServiceName" size="small" placeholder="Service">
              <el-option
                v-for="item in systemctlServices"
                :key="item.value"
                :label="item.label"
                :value="item.value"
              />
            </el-select>
            <el-button plain size="small" @click="runSystemctl('start')" :disabled="!activeTermId">
              Start
            </el-button>
            <el-button plain size="small" @click="runSystemctl('restart')" :disabled="!activeTermId">
              Restart
            </el-button>
            <el-button plain size="small" @click="runSystemctl('stop')" :disabled="!activeTermId">
              Stop
            </el-button>
          </div>
          <div class="terminal-quick-row">
            <span class="terminal-quick-title">Journal Tail</span>
            <el-select v-model="quickJournalService" size="small" placeholder="Service">
              <el-option
                v-for="item in journalServices"
                :key="item.value"
                :label="item.label"
                :value="item.value"
              />
            </el-select>
            <el-button plain size="small" @click="runJournalTail(quickJournalService)" :disabled="!activeTermId">
              Follow
            </el-button>
          </div>
          <div class="terminal-quick-row">
            <span class="terminal-quick-title">Journal Range</span>
            <el-select v-model="quickRangeService" size="small" placeholder="Service">
              <el-option
                v-for="item in journalServices"
                :key="item.value"
                :label="item.label"
                :value="item.value"
              />
            </el-select>
            <el-date-picker
              v-model="journalRange"
              type="datetimerange"
              range-separator="to"
              start-placeholder="Start time"
              end-placeholder="End time"
              size="small"
            />
            <el-button type="primary" size="small" @click="runJournalRange" :disabled="!activeTermId">
              Run
            </el-button>
          </div>
          <p v-if="quickCommandError" class="error">{{ quickCommandError }}</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { reactive, ref, computed, onMounted, onBeforeUnmount, watch, nextTick } from "vue";
import {
  ClipboardGetText,
  ClipboardSetText,
  EventsOn,
  OnFileDrop,
  OnFileDropOff
} from "./wailsjs/wailsjs/runtime/runtime";
import { User, Monitor, DataLine, List } from "@element-plus/icons-vue";
import { Terminal } from "xterm";
import { FitAddon } from "xterm-addon-fit";
import { SearchAddon } from "xterm-addon-search";
import * as api from "./services/api";

const backendReady = ref(false);
const loading = ref(false);
const error = ref("");
const profiles = ref([]);
const mysqlProfiles = ref([]);
const mysqlLoading = ref(false);
const mysqlError = ref("");
const events = ref([]);
const transfers = ref([]);
const hostKeyPrompt = ref(null);
const systemStats = ref(null);
const metricsError = ref("");
const cpuHistory = ref([]);
const memHistory = ref([]);
let metricsTimer = null;

const sessionByProfile = reactive({});
const sessionStateById = reactive({});
const sessionErrorById = reactive({});
const connectingByProfile = reactive({});
const mysqlStatusById = reactive({});

const navItems = [
  { id: "profiles", label: "Profiles", icon: User },
  { id: "terminal", label: "Terminal", icon: Monitor },
  { id: "mysql", label: "MySQL", icon: DataLine },
  { id: "activity", label: "Activity", icon: List }
];

const sidebarItems = computed(() => navItems.filter((item) => item.id !== "terminal"));

const openTabs = ref([navItems[0]]);
const activeTab = ref("profiles");

const terminals = ref([]);
const activeTermId = ref("");
const terminalSessionId = ref("");
const terminalFontSize = ref(12);
const terminalSearchVisible = ref(false);
const terminalSearchQuery = ref("");
const terminalSearchInput = ref(null);
const quickServiceName = ref("supply");
const quickJournalService = ref("supply.service");
const quickRangeService = ref("supply.service");
const journalRange = ref([]);
const quickCommandError = ref("");
const quickPanelVisible = ref(false);

const terminalContainers = new Map();
const terminalInstances = new Map();
const terminalFitAddons = new Map();
const terminalSearchAddons = new Map();
const terminalPending = new Map();
const terminalContextHandlers = new Map();

const filesSessionId = ref("");
const filesPath = ref("/");
const filesEntries = ref([]);
const fileSortKey = ref("name");
const fileSortDir = ref("asc");
const filesLoading = ref(false);
const selectedEntry = ref(null);
const downloadPath = ref("");
const newFolderName = ref("");
const renameName = ref("");
const uploadLocalPath = ref("");
const uploadRemotePath = ref("");
const dropZoneRef = ref(null);
const fileContextMenu = ref({
  visible: false,
  x: 0,
  y: 0,
  entry: null
});
const mysqlTableListRef = ref(null);
const mysqlQueryInputRef = ref(null);
let mysqlHintBlurTimer = null;

const form = reactive({
  id: "",
  name: "",
  group: "",
  host: "",
  port: 22,
  username: "",
  authType: "password",
  privateKeyPath: "",
  useKeyring: true,
  knownHostsPolicy: "ask",
  password: "",
  passphrase: ""
});

const mysqlForm = reactive({
  id: "",
  name: "",
  host: "",
  port: 3306,
  username: "",
  database: "",
  connectionType: "direct",
  sshProfileId: "",
  useKeyring: true,
  useTls: false,
  tlsCaPath: "",
  tlsCertPath: "",
  tlsKeyPath: "",
  tlsSkipVerify: false,
  password: ""
});

const mysqlTabs = ref([]);
const activeMySQLTab = ref("");
const mysqlTabStateById = reactive({});
const mysqlDialogVisible = ref(false);
const mysqlDialogTitle = computed(() =>
  mysqlForm.id ? "Edit MySQL profile" : "New MySQL profile"
);
const emptyMySQLState = reactive({
  profileId: "",
  databases: [],
  tables: [],
  activeDatabase: "",
  activeTable: "",
  newDatabase: "",
  tableSearch: "",
  preview: null,
  previewLoading: false,
  previewError: "",
  originalRows: [],
  rowEdits: {},
  editingCell: { row: -1, col: -1 },
  editSubmitting: false,
  filter: "",
  sortColumn: "",
  sortDirection: "asc",
  limit: 200,
  offset: 0,
  queryText: "",
  queryCursor: 0,
  queryTokenStart: 0,
  queryTokenEnd: 0,
  queryHintToken: "",
  queryHintIndex: 0,
  queryHintVisible: false,
  queryResult: null,
  queryError: "",
  queryRunning: false,
  queryDatabase: "",
  schemaColumns: [],
  schemaLoading: false,
  schemaError: "",
  schemaDialogVisible: false,
  schemaMode: "add",
  schemaForm: {
    originalName: "",
    name: "",
    type: "",
    nullable: true,
    defaultValue: "",
    extra: ""
  }
});
const mysqlState = ref(emptyMySQLState);
const mysqlFilteredTables = computed(() => {
  const state = mysqlState.value;
  const search = state.tableSearch.trim().toLowerCase();
  if (!search) {
    return state.tables;
  }
  return state.tables.filter((name) => name.toLowerCase().includes(search));
});
const mysqlDefaultDatabase = computed(() => {
  const state = mysqlState.value;
  if (!state?.profileId) {
    return "";
  }
  const profile = mysqlProfiles.value.find((item) => item.id === state.profileId);
  return profile?.database || "";
});
const mysqlHintKeywords = [
  "SELECT",
  "FROM",
  "WHERE",
  "INSERT",
  "INTO",
  "VALUES",
  "UPDATE",
  "DELETE",
  "JOIN",
  "LEFT",
  "RIGHT",
  "INNER",
  "OUTER",
  "GROUP",
  "BY",
  "ORDER",
  "LIMIT",
  "OFFSET",
  "DISTINCT",
  "AND",
  "OR",
  "NOT",
  "NULL",
  "LIKE",
  "IN",
  "BETWEEN",
  "DESC",
  "ASC",
  "CREATE",
  "ALTER",
  "DROP",
  "TABLE",
  "DATABASE",
  "SHOW"
];
const mysqlQueryHints = computed(() => {
  const state = mysqlState.value;
  if (!state || !state.queryHintVisible) {
    return [];
  }
  const token = state.queryHintToken.trim();
  if (!token) {
    return [];
  }
  const lower = token.toLowerCase();
  const columns =
    state.schemaColumns?.map((col) => col.name) || [];
  const previewColumns = state.preview?.columns || [];
  const candidates = [...mysqlHintKeywords, ...state.tables, ...previewColumns, ...columns];
  const seen = new Set();
  const results = [];
  for (const item of candidates) {
    if (!item) {
      continue;
    }
    const key = String(item);
    if (seen.has(key)) {
      continue;
    }
    if (key.toLowerCase().startsWith(lower)) {
      results.push(key);
      seen.add(key);
    }
    if (results.length >= 8) {
      break;
    }
  }
  return results;
});
const mysqlHasEdits = computed(() => {
  const state = mysqlState.value;
  return Object.keys(state.rowEdits || {}).length > 0;
});
const fileDownloadTransfers = computed(() => {
  const map = {};
  const sessionId = filesSessionId.value;
  transfers.value.forEach((task) => {
    if (task.direction !== "download") {
      return;
    }
    if (sessionId && task.sessionId !== sessionId) {
      return;
    }
    if (!task.remotePath) {
      return;
    }
    const existing = map[task.remotePath];
    if (!existing || transferPriority(task.state) > transferPriority(existing.state)) {
      map[task.remotePath] = task;
    }
  });
  return map;
});
const fileSortedEntries = computed(() => {
  const entries = [...filesEntries.value];
  const key = fileSortKey.value;
  const dir = fileSortDir.value === "desc" ? -1 : 1;
  entries.sort((a, b) => {
    if (a.isDir !== b.isDir) {
      return a.isDir ? -1 : 1;
    }
    if (key === "mtime") {
      return (a.mtime - b.mtime) * dir;
    }
    return String(a.name).localeCompare(String(b.name)) * dir;
  });
  return entries;
});
const systemctlServices = [
  { label: "supply", value: "supply" }
];
const journalServices = [
  { label: "supply.service", value: "supply.service" },
  { label: "mq.service", value: "mq.service" },
  { label: "cron.service", value: "cron.service" }
];
const cpuTotal = computed(() => systemStats.value?.cpu?.total ?? 0);
const memUsedPercent = computed(() => systemStats.value?.memory?.usedPercent ?? 0);
const memUsed = computed(() => systemStats.value?.memory?.used ?? 0);
const memTotal = computed(() => systemStats.value?.memory?.total ?? 0);
const cpuSparkline = computed(() => sparklinePoints(cpuHistory.value, 160, 48));
const memSparkline = computed(() => sparklinePoints(memHistory.value, 160, 48));

const defaultForm = () => ({
  id: "",
  name: "",
  group: "",
  host: "",
  port: 22,
  username: "",
  authType: "password",
  privateKeyPath: "",
  useKeyring: true,
  knownHostsPolicy: "ask",
  password: "",
  passphrase: ""
});

const connectedProfiles = computed(() => {
  return profiles.value
    .map((profile) => {
      const sessionId = sessionByProfile[profile.id];
      return {
        sessionId,
        label: profile.name || `${profile.username}@${profile.host}`,
        profile
      };
    })
    .filter((item) => item.sessionId && sessionStateById[item.sessionId] === "connected");
});

watch(
  connectedProfiles,
  (list) => {
    const sessionIds = list.map((item) => item.sessionId);
    if (!sessionIds.includes(terminalSessionId.value)) {
      terminalSessionId.value = sessionIds[0] || "";
    }
    if (!sessionIds.includes(filesSessionId.value)) {
      filesSessionId.value = sessionIds[0] || "";
    }
  },
  { immediate: true }
);

watch(activeTermId, async (termId) => {
  if (!termId) {
    return;
  }
  await nextTick();
  initTerminal(termId);
  fitTerminal(termId);
  const term = terminalInstances.get(termId);
  if (term) {
    term.focus();
  }
});

watch(terminalSessionId, (sessionId) => {
  if (!sessionId) {
    return;
  }
  if (filesSessionId.value !== sessionId) {
    filesSessionId.value = sessionId;
  }
});

watch(filesSessionId, (sessionId) => {
  if (!sessionId) {
    filesEntries.value = [];
    return;
  }
  loadFiles(filesPath.value || "/");
});

watch(uploadLocalPath, (value) => {
  if (!value) {
    return;
  }
  if (!uploadRemotePath.value) {
    const baseName = value.split(/[/\\]/).pop() || value;
    uploadRemotePath.value = joinRemotePath(filesPath.value || "/", baseName);
  }
});

watch(mysqlQueryHints, (list) => {
  const state = getActiveMySQLState();
  if (!state) {
    return;
  }
  if (!list.length) {
    state.queryHintIndex = 0;
    return;
  }
  if (state.queryHintIndex >= list.length) {
    state.queryHintIndex = 0;
  }
});

watch(activeMySQLTab, () => {
  mysqlError.value = "";
  const state = getActiveMySQLState();
  mysqlState.value = state || emptyMySQLState;
  if (!state) {
    return;
  }
  if (state.profileId && mysqlStatusLabel(state.profileId) === "connected" && state.databases.length === 0) {
    loadMySQLDatabases(state);
  }
});

function setTermRef(termId, el) {
  const previous = terminalContainers.get(termId);
  if (previous && terminalContextHandlers.has(termId)) {
    previous.removeEventListener("contextmenu", terminalContextHandlers.get(termId));
    terminalContextHandlers.delete(termId);
  }
  if (!el) {
    terminalContainers.delete(termId);
    return;
  }
  terminalContainers.set(termId, el);
  const handler = (event) => {
    event.preventDefault();
    const term = terminalInstances.get(termId);
    if (!term) {
      return;
    }
    const selection = term.getSelection();
    if (selection) {
      ClipboardSetText(selection).catch(() => {});
      return;
    }
    pasteClipboard();
  };
  el.addEventListener("contextmenu", handler);
  terminalContextHandlers.set(termId, handler);
}

function initTerminal(termId) {
  if (terminalInstances.has(termId)) {
    return;
  }
  const container = terminalContainers.get(termId);
  if (!container) {
    return;
  }

  const term = new Terminal({
    convertEol: true,
    cursorBlink: true,
    scrollback: 2000,
    fontFamily: "JetBrains Mono, IBM Plex Mono, Menlo, Consolas, monospace",
    fontSize: terminalFontSize.value,
    theme: {
      background: "#0f1720",
      foreground: "#e2e8f0",
      cursor: "#38bdf8",
      selectionBackground: "rgba(56, 189, 248, 0.3)",
      black: "#0b1220",
      red: "#ef4444",
      green: "#22c55e",
      yellow: "#f59e0b",
      blue: "#60a5fa",
      magenta: "#a855f7",
      cyan: "#22d3ee",
      white: "#e5e7eb",
      brightBlack: "#334155",
      brightRed: "#f87171",
      brightGreen: "#4ade80",
      brightYellow: "#fbbf24",
      brightBlue: "#93c5fd",
      brightMagenta: "#c084fc",
      brightCyan: "#67e8f9",
      brightWhite: "#f8fafc"
    }
  });

  const fitAddon = new FitAddon();
  term.loadAddon(fitAddon);
  const searchAddon = new SearchAddon();
  term.loadAddon(searchAddon);
  term.open(container);
  term.focus();

  term.attachCustomKeyEventHandler((event) => {
    if ((event.ctrlKey || event.metaKey) && event.key.toLowerCase() === "f") {
      openTerminalSearch();
      return false;
    }
    return true;
  });

  const pending = terminalPending.get(termId);
  if (pending) {
    term.write(pending);
    terminalPending.delete(termId);
  }

  term.onData((data) => {
    api.terminalWrite(termId, data).catch((err) => {
      error.value = err.message || String(err);
    });
  });

  terminalInstances.set(termId, term);
  terminalFitAddons.set(termId, fitAddon);
  terminalSearchAddons.set(termId, searchAddon);
  fitTerminal(termId);
}

function fitTerminal(termId) {
  const term = terminalInstances.get(termId);
  const fitAddon = terminalFitAddons.get(termId);
  if (!term || !fitAddon) {
    return;
  }
  fitAddon.fit();
  api.terminalResize(termId, term.cols, term.rows).catch((err) => {
    error.value = err.message || String(err);
  });
}

function disposeTerminal(termId) {
  const term = terminalInstances.get(termId);
  if (term) {
    term.dispose();
    terminalInstances.delete(termId);
  }
  terminalFitAddons.delete(termId);
  terminalSearchAddons.delete(termId);
  terminalPending.delete(termId);
  const container = terminalContainers.get(termId);
  const handler = terminalContextHandlers.get(termId);
  if (container && handler) {
    container.removeEventListener("contextmenu", handler);
  }
  terminalContextHandlers.delete(termId);
  terminalContainers.delete(termId);
}

function handleResize() {
  if (activeTermId.value) {
    fitTerminal(activeTermId.value);
  }
}

function getActiveTerminal() {
  if (!activeTermId.value) {
    return null;
  }
  return terminalInstances.get(activeTermId.value) || null;
}

function statusLabel(profileId) {
  if (connectingByProfile[profileId]) {
    return "connecting";
  }
  const sessionId = sessionByProfile[profileId];
  if (!sessionId) {
    return "idle";
  }
  return sessionStateById[sessionId] || "connecting";
}

function statusClass(profileId) {
  const state = statusLabel(profileId);
  if (state === "connected") {
    return "ok";
  }
  if (state === "disconnected" || state === "connecting") {
    return "warn";
  }
  return "idle";
}

function openTab(id) {
  const item = navItems.find((tab) => tab.id === id);
  if (!item) {
    return;
  }
  if (!openTabs.value.find((tab) => tab.id === id)) {
    openTabs.value.push(item);
  }
  activeTab.value = id;
}

function closeTab(name) {
  const index = openTabs.value.findIndex((tab) => tab.id === name);
  if (index === -1) {
    return;
  }
  openTabs.value.splice(index, 1);
  if (activeTab.value === name) {
    const next = openTabs.value[index] || openTabs.value[index - 1] || openTabs.value[0];
    activeTab.value = next ? next.id : "";
  }
}

async function openTerminalForProfile(profile) {
  if (!profile) {
    return;
  }
  if (statusLabel(profile.id) !== "connected") {
    await connectProfile(profile);
  }
  const sessionId = sessionByProfile[profile.id];
  if (!sessionId) {
    error.value = "Session not connected.";
    return;
  }
  terminalSessionId.value = sessionId;
  openTab("terminal");
  await openTerminal();
}

function mysqlTabLabel(profile) {
  if (!profile) {
    return "MySQL";
  }
  return profile.name || profile.host || "MySQL";
}

function createMySQLTabState(profile) {
  const database = profile?.database || "";
  return reactive({
    profileId: profile?.id || "",
    databases: [],
    tables: [],
    activeDatabase: database,
    activeTable: "",
    newDatabase: "",
    tableSearch: "",
    preview: null,
    previewLoading: false,
    previewError: "",
    originalRows: [],
    rowEdits: {},
    editingCell: { row: -1, col: -1 },
    editSubmitting: false,
    filter: "",
    sortColumn: "",
    sortDirection: "asc",
    limit: 200,
    offset: 0,
    queryText: "",
    queryCursor: 0,
    queryTokenStart: 0,
    queryTokenEnd: 0,
    queryHintToken: "",
    queryHintIndex: 0,
    queryHintVisible: false,
    queryResult: null,
    queryError: "",
    queryRunning: false,
    queryDatabase: database,
    schemaColumns: [],
    schemaLoading: false,
    schemaError: "",
    schemaDialogVisible: false,
    schemaMode: "add",
    schemaForm: {
      originalName: "",
      name: "",
      type: "",
      nullable: true,
      defaultValue: "",
      extra: ""
    }
  });
}

function getActiveMySQLState() {
  return mysqlTabStateById[activeMySQLTab.value] || null;
}

function openMySQLTab(profile) {
  if (!profile) {
    return;
  }
  const existing = mysqlTabs.value.find((tab) => tab.profileId === profile.id);
  if (existing) {
    existing.label = mysqlTabLabel(profile);
    activeMySQLTab.value = existing.id;
    openTab("mysql");
    return;
  }
  const id = profile.id;
  mysqlTabs.value.push({ id, profileId: profile.id, label: mysqlTabLabel(profile) });
  mysqlTabStateById[id] = createMySQLTabState(profile);
  activeMySQLTab.value = id;
  openTab("mysql");
}

function closeMySQLTab(tabId) {
  const index = mysqlTabs.value.findIndex((tab) => tab.id === tabId);
  if (index === -1) {
    return;
  }
  mysqlTabs.value.splice(index, 1);
  delete mysqlTabStateById[tabId];
  if (activeMySQLTab.value === tabId) {
    const next = mysqlTabs.value[index] || mysqlTabs.value[index - 1] || mysqlTabs.value[0];
    activeMySQLTab.value = next ? next.id : "";
  }
}

function quoteMySQLIdent(value) {
  return `\`${String(value).replace(/`/g, "``")}\``;
}

function buildMySQLColumnDefinition(form) {
  const name = quoteMySQLIdent(form.name);
  const type = form.type || "VARCHAR(255)";
  let clause = `${name} ${type}`;
  clause += form.nullable ? " NULL" : " NOT NULL";
  if (form.defaultValue !== "") {
    clause += ` DEFAULT ${form.defaultValue}`;
  }
  if (form.extra) {
    clause += ` ${form.extra}`;
  }
  return clause;
}

function toggleMySQLSort(column) {
  const state = getActiveMySQLState();
  if (!state) {
    return;
  }
  if (state.sortColumn === column) {
    state.sortDirection = state.sortDirection === "asc" ? "desc" : "asc";
  } else {
    state.sortColumn = column;
    state.sortDirection = "asc";
  }
  loadMySQLPreview(state);
}

function mysqlSqlValue(value) {
  if (value === null || value === undefined) {
    return "NULL";
  }
  const text = String(value);
  if (text.trim().toUpperCase() === "NULL") {
    return "NULL";
  }
  return `'${text.replace(/'/g, "''")}'`;
}

function isEditingCell(rowIndex, colIndex) {
  const state = getActiveMySQLState();
  if (!state) {
    return false;
  }
  return state.editingCell.row === rowIndex && state.editingCell.col === colIndex;
}

function startEditCell(rowIndex, colIndex) {
  const state = getActiveMySQLState();
  if (!state || !state.preview) {
    return;
  }
  if (state.editingCell.row === rowIndex && state.editingCell.col === colIndex) {
    return;
  }
  if (state.editingCell.row !== -1) {
    commitEditCell(state.editingCell.row, state.editingCell.col);
  }
  state.editingCell = { row: rowIndex, col: colIndex };
}

function commitEditCell(rowIndex, colIndex) {
  const state = getActiveMySQLState();
  if (!state || !state.preview) {
    return;
  }
  const columns = state.preview.columns || [];
  const columnName = columns[colIndex];
  if (!columnName) {
    state.editingCell = { row: -1, col: -1 };
    return;
  }
  const row = state.preview.rows[rowIndex] || [];
  const newValue = row[colIndex];
  const originalRow = state.originalRows[rowIndex] || [];
  const originalValue = originalRow[colIndex];
  if (newValue === originalValue) {
    if (state.rowEdits[rowIndex]) {
      delete state.rowEdits[rowIndex][columnName];
      if (Object.keys(state.rowEdits[rowIndex]).length === 0) {
        delete state.rowEdits[rowIndex];
      }
    }
  } else {
    if (!state.rowEdits[rowIndex]) {
      state.rowEdits[rowIndex] = {};
    }
    state.rowEdits[rowIndex][columnName] = newValue;
  }
  state.editingCell = { row: -1, col: -1 };
}

async function submitMySQLEdits() {
  const state = getActiveMySQLState();
  if (!state) {
    return;
  }
  if (!state.profileId || !state.activeDatabase || !state.activeTable) {
    mysqlError.value = "Select a table before submitting edits.";
    return;
  }
  if (state.editingCell.row !== -1) {
    commitEditCell(state.editingCell.row, state.editingCell.col);
  }
  const edits = state.rowEdits || {};
  const rowIndexes = Object.keys(edits);
  if (rowIndexes.length === 0) {
    return;
  }
  const columns = state.preview?.columns || [];
  const primaryKeys = (state.schemaColumns || [])
    .filter((col) => String(col.key).toUpperCase() === "PRI")
    .map((col) => col.name);
  state.editSubmitting = true;
  mysqlError.value = "";
  try {
    for (const rowIndexText of rowIndexes) {
      const rowIndex = Number(rowIndexText);
      const rowEdits = edits[rowIndexText];
      if (!rowEdits || Object.keys(rowEdits).length === 0) {
        continue;
      }
      const originalRow = state.originalRows[rowIndex] || [];
      const setParts = Object.keys(rowEdits).map(
        (colName) => `${quoteMySQLIdent(colName)} = ${mysqlSqlValue(rowEdits[colName])}`
      );
      const whereColumns = primaryKeys.length ? primaryKeys : columns;
      const whereParts = whereColumns
        .map((colName) => {
          const idx = columns.indexOf(colName);
          if (idx < 0) {
            return null;
          }
          return `${quoteMySQLIdent(colName)} = ${mysqlSqlValue(originalRow[idx])}`;
        })
        .filter(Boolean);
      if (whereParts.length === 0) {
        throw new Error("No columns available to build WHERE clause.");
      }
      const tableRef = `${quoteMySQLIdent(state.activeDatabase)}.${quoteMySQLIdent(state.activeTable)}`;
      const sql = `UPDATE ${tableRef} SET ${setParts.join(", ")} WHERE ${whereParts.join(
        " AND "
      )} LIMIT 1`;
      await api.mysqlQuery(state.profileId, state.activeDatabase, sql);
    }
    state.rowEdits = {};
    await loadMySQLPreview(state);
    await loadMySQLSchema(state);
  } catch (err) {
    mysqlError.value = err.message || String(err);
  } finally {
    state.editSubmitting = false;
  }
}

async function loadMySQLSchema(stateOverride) {
  const state = stateOverride || getActiveMySQLState();
  if (!state || !state.profileId || !state.activeDatabase || !state.activeTable) {
    if (state) {
      state.schemaColumns = [];
      state.schemaError = "";
    }
    return;
  }
  state.schemaLoading = true;
  state.schemaError = "";
  try {
    state.schemaColumns = await api.mysqlTableSchema(
      state.profileId,
      state.activeDatabase,
      state.activeTable
    );
  } catch (err) {
    state.schemaError = err.message || String(err);
  } finally {
    state.schemaLoading = false;
  }
}

function openMySQLSchemaDialog(mode, column) {
  const state = getActiveMySQLState();
  if (!state) {
    return;
  }
  state.schemaMode = mode;
  state.schemaError = "";
  if (mode === "edit" && column) {
    Object.assign(state.schemaForm, {
      originalName: column.name,
      name: column.name,
      type: column.type,
      nullable: String(column.nullable).toLowerCase() === "yes",
      defaultValue: column.default || "",
      extra: column.extra || ""
    });
  } else {
    Object.assign(state.schemaForm, {
      originalName: "",
      name: "",
      type: "",
      nullable: true,
      defaultValue: "",
      extra: ""
    });
  }
  state.schemaDialogVisible = true;
}

async function saveMySQLSchemaChange() {
  const state = getActiveMySQLState();
  if (!state) {
    return;
  }
  if (!state.profileId || !state.activeDatabase || !state.activeTable) {
    state.schemaError = "Select a table first.";
    return;
  }
  if (!state.schemaForm.name || !state.schemaForm.type) {
    state.schemaError = "Column name and type are required.";
    return;
  }
  const tableRef = `${quoteMySQLIdent(state.activeDatabase)}.${quoteMySQLIdent(state.activeTable)}`;
  let clause = "";
  if (state.schemaMode === "add") {
    clause = `ADD COLUMN ${buildMySQLColumnDefinition(state.schemaForm)}`;
  } else {
    const definition = buildMySQLColumnDefinition(state.schemaForm);
    if (state.schemaForm.originalName && state.schemaForm.originalName !== state.schemaForm.name) {
      clause = `CHANGE COLUMN ${quoteMySQLIdent(state.schemaForm.originalName)} ${definition}`;
    } else {
      clause = `MODIFY COLUMN ${definition}`;
    }
  }
  const sql = `ALTER TABLE ${tableRef} ${clause}`;
  state.schemaError = "";
  try {
    await api.mysqlQuery(state.profileId, state.activeDatabase, sql);
    state.schemaDialogVisible = false;
    await loadMySQLSchema(state);
    await loadMySQLPreview(state);
    await loadMySQLTables(state);
  } catch (err) {
    state.schemaError = err.message || String(err);
  }
}

async function dropMySQLColumn(column) {
  const state = getActiveMySQLState();
  if (!state || !column) {
    return;
  }
  if (!state.profileId || !state.activeDatabase || !state.activeTable) {
    state.schemaError = "Select a table first.";
    return;
  }
  if (!confirm(`Drop column ${column.name}?`)) {
    return;
  }
  const tableRef = `${quoteMySQLIdent(state.activeDatabase)}.${quoteMySQLIdent(state.activeTable)}`;
  const sql = `ALTER TABLE ${tableRef} DROP COLUMN ${quoteMySQLIdent(column.name)}`;
  state.schemaError = "";
  try {
    await api.mysqlQuery(state.profileId, state.activeDatabase, sql);
    await loadMySQLSchema(state);
    await loadMySQLPreview(state);
  } catch (err) {
    state.schemaError = err.message || String(err);
  }
}

function mysqlStatusLabel(profileId) {
  const state = mysqlStatusById[profileId]?.state;
  return state || "disconnected";
}

function mysqlStatusClass(profileId) {
  const state = mysqlStatusLabel(profileId);
  if (state === "connected") {
    return "ok";
  }
  if (state === "error") {
    return "warn";
  }
  return "idle";
}

function mysqlConnectionLabel(profile) {
  if (!profile) {
    return "";
  }
  return profile.connectionType === "ssh" ? "SSH tunnel" : "Direct";
}

function pushEvent(type, payload) {
  const entry = {
    id: `${Date.now()}-${Math.random()}`,
    time: new Date().toLocaleTimeString(),
    type,
    payload: JSON.stringify(payload)
  };
  events.value.unshift(entry);
  if (events.value.length > 120) {
    events.value.pop();
  }
}

function sparklinePoints(values, width, height) {
  if (!values || values.length === 0) {
    return "";
  }
  const max = 100;
  const len = values.length;
  const points = [];
  const stepX = len > 1 ? width / (len - 1) : width;
  for (let i = 0; i < len; i += 1) {
    const value = Math.max(0, Math.min(max, Number(values[i]) || 0));
    const x = i * stepX;
    const y = height - (value / max) * height;
    points.push(`${x.toFixed(2)},${y.toFixed(2)}`);
  }
  return points.join(" ");
}

function pushHistory(listRef, value, limit = 60) {
  const list = listRef.value;
  list.push(value);
  if (list.length > limit) {
    list.splice(0, list.length - limit);
  }
}

function transferPriority(state) {
  if (state === "running") {
    return 3;
  }
  if (state === "queued") {
    return 2;
  }
  if (state === "error") {
    return 1;
  }
  return 0;
}

function isActiveTransfer(task) {
  if (!task) {
    return false;
  }
  return task.state === "running" || task.state === "queued" || task.state === "error";
}

function transferPercent(task) {
  if (!task || !task.totalBytes) {
    return 0;
  }
  const value = (task.doneBytes / task.totalBytes) * 100;
  return Math.min(100, Math.max(0, value));
}

function formatFileSize(value) {
  const size = Number(value) || 0;
  const mb = size / (1024 * 1024);
  return `${mb.toFixed(1)}M`;
}

function registerTransferStub(taskId, sessionId, remotePath, localPath, totalBytes, direction) {
  if (!taskId) {
    return;
  }
  const existing = transfers.value.find((item) => item.id === taskId);
  if (existing) {
    return;
  }
  transfers.value.push({
    id: taskId,
    sessionId: sessionId || "",
    localPath: localPath || "",
    remotePath: remotePath || "",
    direction: direction || "",
    totalBytes: totalBytes || 0,
    doneBytes: 0,
    speedBytes: 0,
    state: "queued",
    message: ""
  });
}

function formatBytes(value) {
  const size = Number(value) || 0;
  if (size <= 0) {
    return "0 B";
  }
  const units = ["B", "KB", "MB", "GB", "TB"];
  let index = 0;
  let current = size;
  while (current >= 1024 && index < units.length - 1) {
    current /= 1024;
    index += 1;
  }
  return `${current.toFixed(1)} ${units[index]}`;
}

async function loadSystemStats() {
  if (!backendReady.value) {
    return;
  }
  metricsError.value = "";
  try {
    const stats = await api.systemStats();
    systemStats.value = stats;
    pushHistory(cpuHistory, stats?.cpu?.total ?? 0);
    pushHistory(memHistory, stats?.memory?.usedPercent ?? 0);
  } catch (err) {
    metricsError.value = err.message || String(err);
  }
}

function startMetrics() {
  if (metricsTimer) {
    clearInterval(metricsTimer);
  }
  if (!backendReady.value) {
    return;
  }
  loadSystemStats();
  metricsTimer = setInterval(loadSystemStats, 2000);
}

function stopMetrics() {
  if (metricsTimer) {
    clearInterval(metricsTimer);
    metricsTimer = null;
  }
}

function clearEvents() {
  events.value = [];
}

async function reloadProfiles() {
  if (!backendReady.value) {
    return;
  }
  loading.value = true;
  error.value = "";
  try {
    profiles.value = await api.profilesList();
  } catch (err) {
    error.value = err.message || String(err);
  } finally {
    loading.value = false;
  }
}

function newProfile() {
  Object.assign(form, defaultForm());
}

function editProfile(profile) {
  Object.assign(form, {
    id: profile.id,
    name: profile.name,
    group: profile.group,
    host: profile.host,
    port: profile.port || 22,
    username: profile.username,
    authType: profile.authType,
    privateKeyPath: profile.privateKeyPath,
    useKeyring: profile.useKeyring,
    knownHostsPolicy: profile.knownHostsPolicy,
    password: "",
    passphrase: ""
  });
}

async function saveProfile() {
  error.value = "";
  let profileId = "";
  try {
    const payload = {
      id: form.id,
      name: form.name,
      group: form.group,
      host: form.host,
      port: form.port || 22,
      username: form.username,
      authType: form.authType,
      privateKeyPath: form.privateKeyPath,
      useKeyring: form.useKeyring,
      knownHostsPolicy: form.knownHostsPolicy
    };
    profileId = await api.profilesSave(payload);
    form.id = profileId;
  } catch (err) {
    error.value = err.message || String(err);
    return;
  }

  let credentialError = "";
  try {
    if (!form.useKeyring) {
      await api.credentialsDelete(profileId);
    } else if (form.authType === "password" && form.password) {
      await api.credentialsSetPassword(profileId, form.password);
    } else if (form.authType === "privateKey" && form.passphrase) {
      await api.credentialsSetPrivateKeyPassphrase(profileId, form.passphrase);
    }
  } catch (err) {
    credentialError = err.message || String(err);
  }

  form.password = "";
  form.passphrase = "";

  await reloadProfiles();

  if (credentialError) {
    error.value = `Profile saved, but credentials failed: ${credentialError}`;
  }
}

async function deleteProfile(profile) {
  if (!confirm(`Delete profile ${profile.name || profile.host}?`)) {
    return;
  }
  error.value = "";
  try {
    await api.profilesDelete(profile.id);
    try {
      await api.credentialsDelete(profile.id);
    } catch (err) {
      error.value = err.message || String(err);
    }
    if (form.id === profile.id) {
      newProfile();
    }
    await reloadProfiles();
  } catch (err) {
    error.value = err.message || String(err);
  }
}

async function clearCredentials() {
  if (!form.id) {
    return;
  }
  try {
    await api.credentialsDelete(form.id);
    form.password = "";
    form.passphrase = "";
  } catch (err) {
    error.value = err.message || String(err);
  }
}

async function connectProfile(profile) {
  if (connectingByProfile[profile.id]) {
    return;
  }
  error.value = "";
  connectingByProfile[profile.id] = true;
  pushEvent("session:connect", { profileId: profile.id });
  try {
    const sessionId = await api.sessionConnect(profile.id);
    sessionByProfile[profile.id] = sessionId;
  } catch (err) {
    error.value = err.message || String(err);
    pushEvent("session:connect:error", { profileId: profile.id, error: error.value });
  } finally {
    connectingByProfile[profile.id] = false;
  }
}

async function disconnectProfile(profile) {
  const sessionId = sessionByProfile[profile.id];
  if (!sessionId) {
    return;
  }
  try {
    await api.sessionDisconnect(sessionId);
  } catch (err) {
    error.value = err.message || String(err);
  }
}

async function reloadMySQLProfiles() {
  if (!backendReady.value) {
    return;
  }
  mysqlLoading.value = true;
  mysqlError.value = "";
  try {
    const list = await api.mysqlProfilesList();
    mysqlProfiles.value = list || [];
    const statuses = await Promise.all(
      mysqlProfiles.value.map((profile) =>
        api.mysqlStatus(profile.id).catch(() => ({ state: "disconnected", lastError: "" }))
      )
    );
    statuses.forEach((status, index) => {
      mysqlStatusById[mysqlProfiles.value[index].id] = status;
    });
    const profileById = new Map(mysqlProfiles.value.map((profile) => [profile.id, profile]));
    mysqlTabs.value = mysqlTabs.value.filter((tab) => profileById.has(tab.profileId));
    mysqlTabs.value.forEach((tab) => {
      const profile = profileById.get(tab.profileId);
      tab.label = mysqlTabLabel(profile);
      if (!mysqlTabStateById[tab.id]) {
        mysqlTabStateById[tab.id] = createMySQLTabState(profile);
      }
    });
    Object.keys(mysqlTabStateById).forEach((tabId) => {
      if (!mysqlTabs.value.find((tab) => tab.id === tabId)) {
        delete mysqlTabStateById[tabId];
      }
    });
    if (activeMySQLTab.value && !mysqlTabs.value.find((tab) => tab.id === activeMySQLTab.value)) {
      activeMySQLTab.value = mysqlTabs.value[0]?.id || "";
    }
    if (mysqlForm.id && !mysqlProfiles.value.find((item) => item.id === mysqlForm.id)) {
      newMySQLProfile();
    }
  } catch (err) {
    mysqlError.value = err.message || String(err);
  } finally {
    mysqlLoading.value = false;
  }
}

function newMySQLProfile() {
  Object.assign(mysqlForm, {
    id: "",
    name: "",
    host: "",
    port: 3306,
    username: "",
    database: "",
    connectionType: "direct",
    sshProfileId: "",
    useKeyring: true,
    useTls: false,
    tlsCaPath: "",
    tlsCertPath: "",
    tlsKeyPath: "",
    tlsSkipVerify: false,
    password: ""
  });
}

function editMySQLProfile(profile) {
  Object.assign(mysqlForm, {
    id: profile.id,
    name: profile.name,
    host: profile.host,
    port: profile.port || 3306,
    username: profile.username,
    database: profile.database,
    connectionType: profile.connectionType || "direct",
    sshProfileId: profile.sshProfileId || "",
    useKeyring: profile.useKeyring,
    useTls: profile.useTls,
    tlsCaPath: profile.tlsCaPath || "",
    tlsCertPath: profile.tlsCertPath || "",
    tlsKeyPath: profile.tlsKeyPath || "",
    tlsSkipVerify: profile.tlsSkipVerify,
    password: ""
  });
}

function showMySQLDialog(profile) {
  if (profile) {
    editMySQLProfile(profile);
  } else {
    newMySQLProfile();
  }
  mysqlDialogVisible.value = true;
}

async function saveMySQLProfile() {
  mysqlError.value = "";
  let profileId = "";
  try {
    const payload = {
      id: mysqlForm.id,
      name: mysqlForm.name,
      host: mysqlForm.host,
      port: mysqlForm.port || 3306,
      username: mysqlForm.username,
      database: mysqlForm.database,
      connectionType: mysqlForm.connectionType,
      sshProfileId: mysqlForm.sshProfileId,
      useKeyring: mysqlForm.useKeyring,
      useTls: mysqlForm.useTls,
      tlsCaPath: mysqlForm.tlsCaPath,
      tlsCertPath: mysqlForm.tlsCertPath,
      tlsKeyPath: mysqlForm.tlsKeyPath,
      tlsSkipVerify: mysqlForm.tlsSkipVerify
    };
    profileId = await api.mysqlProfilesSave(payload);
    mysqlForm.id = profileId;
  } catch (err) {
    mysqlError.value = err.message || String(err);
    return;
  }

  let credentialError = "";
  try {
    if (!mysqlForm.useKeyring) {
      await api.mysqlCredentialsDelete(profileId);
    } else if (mysqlForm.password) {
      await api.mysqlCredentialsSetPassword(profileId, mysqlForm.password);
    }
  } catch (err) {
    credentialError = err.message || String(err);
  }

  mysqlForm.password = "";
  await reloadMySQLProfiles();
  mysqlDialogVisible.value = false;

  if (credentialError) {
    mysqlError.value = `Profile saved, but credentials failed: ${credentialError}`;
  }
}

async function deleteMySQLProfile(profile) {
  if (!confirm(`Delete MySQL profile ${profile.name || profile.host}?`)) {
    return;
  }
  mysqlError.value = "";
  try {
    await api.mysqlProfilesDelete(profile.id);
    try {
      await api.mysqlCredentialsDelete(profile.id);
    } catch (err) {
      mysqlError.value = err.message || String(err);
    }
    if (mysqlForm.id === profile.id) {
      newMySQLProfile();
    }
    const tab = mysqlTabs.value.find((item) => item.profileId === profile.id);
    if (tab) {
      closeMySQLTab(tab.id);
    }
    await reloadMySQLProfiles();
  } catch (err) {
    mysqlError.value = err.message || String(err);
  }
}

async function connectMySQLProfile(profile) {
  if (!profile) {
    return;
  }
  mysqlError.value = "";
  const currentStatus = mysqlStatusLabel(profile.id);
  openMySQLTab(profile);
  if (currentStatus === "connected") {
    return;
  }
  const state = getActiveMySQLState();
  try {
    const nextStatus = await api.mysqlConnect(profile.id);
    mysqlStatusById[profile.id] = nextStatus;
    if (state && state.profileId === profile.id) {
      await loadMySQLDatabases(state);
    }
  } catch (err) {
    mysqlError.value = err.message || String(err);
    mysqlStatusById[profile.id] = { state: "error", lastError: mysqlError.value };
  }
}

async function disconnectMySQLProfile(profile) {
  if (!profile) {
    return;
  }
  try {
    await api.mysqlDisconnect(profile.id);
    mysqlStatusById[profile.id] = { state: "disconnected", lastError: "" };
    const tab = mysqlTabs.value.find((item) => item.profileId === profile.id);
    const state = tab ? mysqlTabStateById[tab.id] : null;
    if (state) {
      state.databases = [];
      state.tables = [];
      state.activeDatabase = "";
      state.activeTable = "";
      state.tableSearch = "";
      state.preview = null;
      state.previewError = "";
      state.previewLoading = false;
      state.originalRows = [];
      state.rowEdits = {};
      state.editingCell = { row: -1, col: -1 };
      state.editSubmitting = false;
      state.sortColumn = "";
      state.sortDirection = "asc";
      state.schemaColumns = [];
      state.schemaError = "";
      state.schemaLoading = false;
      state.schemaDialogVisible = false;
      state.queryResult = null;
      state.queryError = "";
      state.queryRunning = false;
    }
  } catch (err) {
    mysqlError.value = err.message || String(err);
  }
}

async function loadMySQLDatabases(stateOverride) {
  const state = stateOverride || getActiveMySQLState();
  if (!state || !state.profileId) {
    return;
  }
  mysqlError.value = "";
  try {
    const list = await api.mysqlListDatabases(state.profileId);
    state.databases = list || [];
    if (!state.activeDatabase) {
      const profile = mysqlProfiles.value.find((item) => item.id === state.profileId);
      if (profile?.database) {
        state.activeDatabase = profile.database;
      }
    }
    if (!state.databases.includes(state.activeDatabase)) {
      state.activeDatabase = state.databases[0] || "";
    }
    state.queryDatabase = state.activeDatabase;
    if (state.activeDatabase) {
      await loadMySQLTables(state);
    } else {
      state.tables = [];
      state.activeTable = "";
      state.preview = null;
      state.previewError = "";
      state.originalRows = [];
      state.rowEdits = {};
      state.editingCell = { row: -1, col: -1 };
      state.sortColumn = "";
      state.sortDirection = "asc";
      state.schemaColumns = [];
      state.schemaError = "";
    }
  } catch (err) {
    mysqlError.value = err.message || String(err);
  }
}

async function loadMySQLTables(stateOverride) {
  const state = stateOverride || getActiveMySQLState();
  if (!state || !state.profileId || !state.activeDatabase) {
    if (state) {
      state.tables = [];
      state.activeTable = "";
    }
    return;
  }
  mysqlError.value = "";
  try {
    const list = await api.mysqlListTables(state.profileId, state.activeDatabase);
    state.tables = list || [];
    if (!state.tables.includes(state.activeTable)) {
      state.activeTable = "";
      state.preview = null;
      state.previewError = "";
      state.originalRows = [];
      state.rowEdits = {};
      state.editingCell = { row: -1, col: -1 };
      state.sortColumn = "";
      state.sortDirection = "asc";
      state.schemaColumns = [];
      state.schemaError = "";
    }
  } catch (err) {
    mysqlError.value = err.message || String(err);
  }
}

async function selectMySQLDatabase(name) {
  const state = getActiveMySQLState();
  if (!state) {
    return;
  }
  state.activeDatabase = name;
  state.activeTable = "";
  state.preview = null;
  state.previewError = "";
  state.sortColumn = "";
  state.sortDirection = "asc";
  state.schemaColumns = [];
  state.schemaError = "";
  state.originalRows = [];
  state.rowEdits = {};
  state.editingCell = { row: -1, col: -1 };
  state.queryDatabase = name;
  await loadMySQLTables(state);
}

async function saveMySQLDefaultDatabase() {
  const state = getActiveMySQLState();
  if (!state || !state.profileId || !state.activeDatabase) {
    return;
  }
  const profile = mysqlProfiles.value.find((item) => item.id === state.profileId);
  if (!profile) {
    return;
  }
  mysqlError.value = "";
  try {
    const payload = {
      id: profile.id,
      name: profile.name,
      host: profile.host,
      port: profile.port,
      username: profile.username,
      database: state.activeDatabase,
      connectionType: profile.connectionType,
      sshProfileId: profile.sshProfileId,
      useKeyring: profile.useKeyring,
      useTls: profile.useTls,
      tlsCaPath: profile.tlsCaPath,
      tlsCertPath: profile.tlsCertPath,
      tlsKeyPath: profile.tlsKeyPath,
      tlsSkipVerify: profile.tlsSkipVerify
    };
    await api.mysqlProfilesSave(payload);
    await reloadMySQLProfiles();
  } catch (err) {
    mysqlError.value = err.message || String(err);
  }
}

function focusMySQLTableList() {
  mysqlTableListRef.value?.focus();
}

function handleMySQLTableKeydown(event) {
  if (event.key !== "ArrowDown" && event.key !== "ArrowUp") {
    return;
  }
  const state = getActiveMySQLState();
  if (!state) {
    return;
  }
  const tables = mysqlFilteredTables.value;
  if (!tables.length) {
    return;
  }
  event.preventDefault();
  const currentIndex = tables.indexOf(state.activeTable);
  const baseIndex =
    currentIndex === -1
      ? event.key === "ArrowDown"
        ? -1
        : tables.length
      : currentIndex;
  const nextIndex =
    event.key === "ArrowDown"
      ? Math.min(tables.length - 1, baseIndex + 1)
      : Math.max(0, baseIndex - 1);
  const nextName = tables[nextIndex];
  if (!nextName) {
    return;
  }
  selectMySQLTable(nextName);
  nextTick(() => {
    scrollMySQLTableIntoView(nextIndex);
  });
}

function scrollMySQLTableIntoView(index) {
  const container = mysqlTableListRef.value;
  if (!container) {
    return;
  }
  const item = container.querySelector(`[data-index="${index}"]`);
  if (item) {
    item.scrollIntoView({ block: "nearest" });
  }
}

async function selectMySQLTable(name) {
  const state = getActiveMySQLState();
  if (!state) {
    return;
  }
  state.activeTable = name;
  state.sortColumn = "";
  state.sortDirection = "asc";
  await loadMySQLPreview(state);
  await loadMySQLSchema(state);
}

async function loadMySQLPreview(stateOverride) {
  const state = stateOverride || getActiveMySQLState();
  if (!state || !state.profileId || !state.activeDatabase || !state.activeTable) {
    return;
  }
  state.previewLoading = true;
  state.previewError = "";
  try {
    state.preview = await api.mysqlPreviewTable(
      state.profileId,
      state.activeDatabase,
      state.activeTable,
      state.filter,
      state.sortColumn,
      state.sortDirection,
      state.limit,
      state.offset
    );
    state.originalRows = (state.preview?.rows || []).map((row) => row.slice());
    state.rowEdits = {};
    state.editingCell = { row: -1, col: -1 };
    state.editSubmitting = false;
  } catch (err) {
    state.previewError = err.message || String(err);
  } finally {
    state.previewLoading = false;
  }
}

function handleMySQLQueryFocus() {
  const state = getActiveMySQLState();
  if (!state) {
    return;
  }
  if (mysqlHintBlurTimer) {
    clearTimeout(mysqlHintBlurTimer);
    mysqlHintBlurTimer = null;
  }
  state.queryHintVisible = !!state.queryHintToken;
}

function handleMySQLQueryBlur() {
  const state = getActiveMySQLState();
  if (!state) {
    return;
  }
  if (mysqlHintBlurTimer) {
    clearTimeout(mysqlHintBlurTimer);
  }
  mysqlHintBlurTimer = setTimeout(() => {
    state.queryHintVisible = false;
  }, 120);
}

function handleMySQLQueryKeydown(event) {
  const state = getActiveMySQLState();
  if (!state || !state.queryHintVisible) {
    return;
  }
  const hints = mysqlQueryHints.value;
  if (!hints.length) {
    return;
  }
  if (event.key === "Tab") {
    event.preventDefault();
    const hint = hints[state.queryHintIndex] || hints[0];
    applyMySQLHint(hint);
    return;
  }
  if (event.key === "ArrowDown") {
    event.preventDefault();
    state.queryHintIndex = (state.queryHintIndex + 1) % hints.length;
    return;
  }
  if (event.key === "ArrowUp") {
    event.preventDefault();
    state.queryHintIndex = (state.queryHintIndex - 1 + hints.length) % hints.length;
  }
}

function handleMySQLQueryKeyup(event) {
  const state = getActiveMySQLState();
  if (!state) {
    return;
  }
  const cursor = event.target?.selectionStart ?? state.queryText.length;
  state.queryCursor = cursor;
  const info = getMySQLQueryToken(state.queryText, cursor);
  state.queryTokenStart = info.start;
  state.queryTokenEnd = info.end;
  state.queryHintToken = info.token;
  state.queryHintIndex = 0;
  state.queryHintVisible = info.token.length > 0;
}

function getMySQLQueryToken(text, cursor) {
  const safeText = text || "";
  const safeCursor = Math.max(0, Math.min(cursor ?? safeText.length, safeText.length));
  const left = safeText.slice(0, safeCursor);
  const match = left.match(/[A-Za-z0-9_]+$/);
  if (!match) {
    return { token: "", start: safeCursor, end: safeCursor };
  }
  const token = match[0];
  const start = safeCursor - token.length;
  return { token, start, end: safeCursor };
}

function applyMySQLHint(value) {
  const state = getActiveMySQLState();
  if (!state || !value) {
    return;
  }
  const text = state.queryText || "";
  const start = Math.max(0, state.queryTokenStart);
  const end = Math.max(start, state.queryTokenEnd);
  state.queryText = `${text.slice(0, start)}${value}${text.slice(end)}`;
  state.queryCursor = start + value.length;
  state.queryHintToken = "";
  state.queryHintVisible = false;
  state.queryHintIndex = 0;
  nextTick(() => {
    const inputEl = mysqlQueryInputRef.value?.textarea || mysqlQueryInputRef.value?.input;
    if (inputEl) {
      inputEl.focus();
      inputEl.setSelectionRange(state.queryCursor, state.queryCursor);
    }
  });
}

async function runMySQLQuery() {
  const state = getActiveMySQLState();
  if (!state) {
    return;
  }
  state.queryError = "";
  if (!state.profileId) {
    state.queryError = "Select a MySQL tab first.";
    return;
  }
  state.queryRunning = true;
  try {
    state.queryResult = await api.mysqlQuery(
      state.profileId,
      state.queryDatabase,
      state.queryText
    );
  } catch (err) {
    state.queryError = err.message || String(err);
  } finally {
    state.queryRunning = false;
  }
}

async function createMySQLDatabase() {
  const state = getActiveMySQLState();
  if (!state || !state.profileId || !state.newDatabase) {
    return;
  }
  mysqlError.value = "";
  try {
    await api.mysqlCreateDatabase(state.profileId, state.newDatabase);
    state.newDatabase = "";
    await loadMySQLDatabases(state);
  } catch (err) {
    mysqlError.value = err.message || String(err);
  }
}

async function dropMySQLDatabase() {
  const state = getActiveMySQLState();
  if (!state || !state.profileId || !state.activeDatabase) {
    return;
  }
  if (!confirm(`Drop database ${state.activeDatabase}?`)) {
    return;
  }
  mysqlError.value = "";
  try {
    await api.mysqlDropDatabase(state.profileId, state.activeDatabase);
    state.activeDatabase = "";
    state.activeTable = "";
    await loadMySQLDatabases(state);
  } catch (err) {
    mysqlError.value = err.message || String(err);
  }
}

async function dropMySQLTable() {
  const state = getActiveMySQLState();
  if (!state || !state.profileId || !state.activeDatabase || !state.activeTable) {
    return;
  }
  if (!confirm(`Drop table ${state.activeTable}?`)) {
    return;
  }
  mysqlError.value = "";
  try {
    await api.mysqlDropTable(state.profileId, state.activeDatabase, state.activeTable);
    state.activeTable = "";
    await loadMySQLTables(state);
  } catch (err) {
    mysqlError.value = err.message || String(err);
  }
}

async function openTerminal() {
  if (!terminalSessionId.value) {
    return;
  }
  error.value = "";
  try {
    const termId = await api.terminalOpen(terminalSessionId.value, 120, 32);
    const title = connectedProfiles.value.find((item) => item.sessionId === terminalSessionId.value)?.label;
    terminals.value.push({
      id: termId,
      sessionId: terminalSessionId.value,
      title: title || `Session ${termId.slice(0, 6)}`
    });
    activeTermId.value = termId;
    await nextTick();
    initTerminal(termId);
  } catch (err) {
    error.value = err.message || String(err);
  }
}

async function closeTerminal(termId) {
  try {
    await api.terminalClose(termId);
  } catch (err) {
    error.value = err.message || String(err);
  }
  terminals.value = terminals.value.filter((term) => term.id !== termId);
  disposeTerminal(termId);
  if (activeTermId.value === termId) {
    activeTermId.value = terminals.value[0]?.id || "";
  }
}

async function copySelection() {
  const term = getActiveTerminal();
  if (!term) {
    return;
  }
  const selection = term.getSelection();
  if (!selection) {
    return;
  }
  try {
    await ClipboardSetText(selection);
  } catch (err) {
    error.value = err.message || String(err);
  }
}

async function pasteClipboard() {
  if (!activeTermId.value) {
    return;
  }
  try {
    const text = await ClipboardGetText();
    if (!text) {
      return;
    }
    const normalized = text.replace(/\r\n/g, "\n");
    await api.terminalWrite(activeTermId.value, normalized);
  } catch (err) {
    error.value = err.message || String(err);
  }
}

function clearTerminal() {
  const term = getActiveTerminal();
  if (!term) {
    return;
  }
  term.clear();
}

function setTerminalFontSize(nextSize) {
  terminalFontSize.value = Math.min(20, Math.max(10, nextSize));
  terminalInstances.forEach((term, termId) => {
    term.setOption("fontSize", terminalFontSize.value);
    fitTerminal(termId);
  });
}

function openTerminalSearch() {
  if (!activeTermId.value) {
    return;
  }
  terminalSearchVisible.value = true;
  nextTick(() => {
    terminalSearchInput.value?.focus();
  });
}

function closeTerminalSearch() {
  terminalSearchVisible.value = false;
  terminalSearchQuery.value = "";
}

function findTerminalNext() {
  const termId = activeTermId.value;
  if (!termId || !terminalSearchQuery.value) {
    return;
  }
  const addon = terminalSearchAddons.get(termId);
  if (addon) {
    addon.findNext(terminalSearchQuery.value, { caseSensitive: false });
  }
}

function findTerminalPrev() {
  const termId = activeTermId.value;
  if (!termId || !terminalSearchQuery.value) {
    return;
  }
  const addon = terminalSearchAddons.get(termId);
  if (addon) {
    addon.findPrevious(terminalSearchQuery.value, { caseSensitive: false });
  }
}

function handleTerminalSearchKey(event) {
  if (event.key === "Escape") {
    closeTerminalSearch();
    return;
  }
  if (event.key === "Enter") {
    if (event.shiftKey) {
      findTerminalPrev();
    } else {
      findTerminalNext();
    }
  }
}

function formatDateTime(value) {
  if (!value) {
    return "";
  }
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return "";
  }
  const pad = (num) => String(num).padStart(2, "0");
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())} ${pad(
    date.getHours()
  )}:${pad(date.getMinutes())}:${pad(date.getSeconds())}`;
}

async function runTerminalCommand(command) {
  if (!activeTermId.value) {
    quickCommandError.value = "Open a terminal first.";
    return;
  }
  quickCommandError.value = "";
  try {
    await api.terminalWrite(activeTermId.value, `${command}\n`);
  } catch (err) {
    quickCommandError.value = err.message || String(err);
  }
}

function runSystemctl(action) {
  if (!quickServiceName.value) {
    return;
  }
  runTerminalCommand(`systemctl ${action} ${quickServiceName.value}`);
}

function runJournalTail(service) {
  if (!service) {
    return;
  }
  runTerminalCommand(`journalctl -u ${service} -f`);
}

function runJournalRange() {
  const range = journalRange.value || [];
  if (range.length !== 2) {
    quickCommandError.value = "Select start and end time.";
    return;
  }
  const since = formatDateTime(range[0]);
  const until = formatDateTime(range[1]);
  if (!since || !until) {
    quickCommandError.value = "Invalid time range.";
    return;
  }
  runTerminalCommand(
    `journalctl -u ${quickRangeService.value} -f --since "${since}" --until "${until}" --all`
  );
}

function toggleFileSort(key) {
  if (fileSortKey.value === key) {
    fileSortDir.value = fileSortDir.value === "asc" ? "desc" : "asc";
    return;
  }
  fileSortKey.value = key;
  fileSortDir.value = key === "mtime" ? "desc" : "asc";
}

function fileSortIndicator(key) {
  if (fileSortKey.value !== key) {
    return "";
  }
  return fileSortDir.value === "asc" ? "^" : "v";
}

function openFileContextMenu(entry, event) {
  event.preventDefault();
  selectEntry(entry);
  fileContextMenu.value = {
    visible: true,
    x: event.clientX,
    y: event.clientY,
    entry
  };
}

function closeFileContextMenu() {
  if (!fileContextMenu.value.visible) {
    return;
  }
  fileContextMenu.value = {
    visible: false,
    x: 0,
    y: 0,
    entry: null
  };
}

async function downloadEntry(entry) {
  if (!entry || entry.isDir) {
    return;
  }
  if (!filesSessionId.value) {
    return;
  }
  error.value = "";
  try {
    const path = await api.dialogSaveFile("Save download as", entry.name);
    if (!path) {
      return;
    }
    const taskId = await api.transferDownload(filesSessionId.value, entry.path, path);
    registerTransferStub(taskId, filesSessionId.value, entry.path, path, entry.size, "download");
  } catch (err) {
    error.value = err.message || String(err);
  }
}

async function deleteEntry(entry) {
  if (!entry) {
    return;
  }
  selectEntry(entry);
  await removeSelected();
}

async function renameEntry(entry) {
  if (!entry) {
    return;
  }
  if (!filesSessionId.value) {
    return;
  }
  const nextName = prompt("Rename to", entry.name || "");
  if (!nextName || nextName === entry.name) {
    return;
  }
  error.value = "";
  try {
    const toPath = joinRemotePath(filesPath.value || "/", nextName);
    await api.filesRename(filesSessionId.value, entry.path, toPath);
    await loadFiles(filesPath.value || "/");
  } catch (err) {
    error.value = err.message || String(err);
  }
}

async function loadFiles(path) {
  if (!filesSessionId.value) {
    return;
  }
  filesLoading.value = true;
  error.value = "";
  try {
    const items = await api.filesList(filesSessionId.value, path);
    filesEntries.value = items || [];
    filesPath.value = path;
    selectedEntry.value = null;
  } catch (err) {
    error.value = err.message || String(err);
  } finally {
    filesLoading.value = false;
  }
}

function reloadFiles() {
  loadFiles(filesPath.value || "/");
}

function goUp() {
  if (!filesPath.value || filesPath.value === "/") {
    return;
  }
  const trimmed = filesPath.value.replace(/\/+$/, "");
  const index = trimmed.lastIndexOf("/");
  const parent = index <= 0 ? "/" : trimmed.slice(0, index);
  loadFiles(parent);
}

function selectEntry(entry) {
  selectedEntry.value = entry;
  renameName.value = entry?.name || "";
}

function openEntry(entry) {
  if (entry.isDir) {
    loadFiles(entry.path);
  }
}

async function downloadSelected() {
  if (!selectedEntry.value || selectedEntry.value.isDir) {
    return;
  }
  if (!downloadPath.value) {
    error.value = "Provide a local path for download.";
    return;
  }
  error.value = "";
  try {
    const taskId = await api.transferDownload(
      filesSessionId.value,
      selectedEntry.value.path,
      downloadPath.value
    );
    registerTransferStub(
      taskId,
      filesSessionId.value,
      selectedEntry.value.path,
      downloadPath.value,
      selectedEntry.value.size,
      "download"
    );
  } catch (err) {
    error.value = err.message || String(err);
  }
}

async function createFolder() {
  if (!filesSessionId.value || !newFolderName.value) {
    return;
  }
  error.value = "";
  try {
    const target = joinRemotePath(filesPath.value || "/", newFolderName.value);
    await api.filesMkdir(filesSessionId.value, target);
    newFolderName.value = "";
    await loadFiles(filesPath.value || "/");
  } catch (err) {
    error.value = err.message || String(err);
  }
}

async function renameSelected() {
  if (!filesSessionId.value || !selectedEntry.value || !renameName.value) {
    return;
  }
  const fromPath = selectedEntry.value.path;
  const toPath = joinRemotePath(filesPath.value || "/", renameName.value);
  if (fromPath === toPath) {
    return;
  }
  error.value = "";
  try {
    await api.filesRename(filesSessionId.value, fromPath, toPath);
    selectedEntry.value = null;
    renameName.value = "";
    await loadFiles(filesPath.value || "/");
  } catch (err) {
    error.value = err.message || String(err);
  }
}

async function removeSelected() {
  if (!filesSessionId.value || !selectedEntry.value) {
    return;
  }
  const entry = selectedEntry.value;
  const message = entry.isDir
    ? `Delete directory ${entry.name} recursively?`
    : `Delete file ${entry.name}?`;
  if (!confirm(message)) {
    return;
  }
  error.value = "";
  try {
    await api.filesRemove(filesSessionId.value, entry.path, entry.isDir);
    selectedEntry.value = null;
    await loadFiles(filesPath.value || "/");
  } catch (err) {
    error.value = err.message || String(err);
  }
}

async function uploadFile() {
  if (!filesSessionId.value || !uploadLocalPath.value || !uploadRemotePath.value) {
    error.value = "Provide both local and remote paths for upload.";
    return;
  }
  error.value = "";
  try {
    await api.transferUpload(filesSessionId.value, uploadLocalPath.value, uploadRemotePath.value);
  } catch (err) {
    error.value = err.message || String(err);
  }
}

async function browseUploadLocal() {
  error.value = "";
  try {
    const path = await api.dialogOpenFile("Select file to upload");
    if (path) {
      uploadLocalPath.value = path;
    }
  } catch (err) {
    error.value = err.message || String(err);
  }
}

async function browseDownloadPath() {
  if (!selectedEntry.value || selectedEntry.value.isDir) {
    error.value = "Select a file to download.";
    return;
  }
  error.value = "";
  try {
    const path = await api.dialogSaveFile("Save download as", selectedEntry.value.name);
    if (path) {
      downloadPath.value = path;
    }
  } catch (err) {
    error.value = err.message || String(err);
  }
}

function formatTime(epochSeconds) {
  if (!epochSeconds) {
    return "";
  }
  const date = new Date(epochSeconds * 1000);
  return date.toLocaleString();
}

function joinRemotePath(base, name) {
  if (!base || base === "/") {
    return `/${name}`;
  }
  return `${base.replace(/[\\/]+$/, "")}/${name}`;
}

function updateTransfer(payload) {
  const taskId = payload.taskId || payload.id;
  if (!taskId) {
    return;
  }
  const index = transfers.value.findIndex((task) => task.id === taskId);
  const existing = index === -1 ? {} : transfers.value[index];
  const entry = {
    id: taskId,
    sessionId: payload.sessionId ?? existing.sessionId ?? "",
    localPath: payload.localPath ?? existing.localPath ?? "",
    remotePath: payload.remotePath ?? existing.remotePath ?? "",
    direction: payload.direction ?? existing.direction ?? "",
    totalBytes: payload.totalBytes ?? existing.totalBytes ?? 0,
    doneBytes: payload.doneBytes ?? existing.doneBytes ?? 0,
    speedBytes: payload.speedBytes ?? existing.speedBytes ?? 0,
    state: payload.state ?? existing.state ?? "running",
    message: payload.message ?? existing.message ?? ""
  };

  if (index === -1) {
    transfers.value.push(entry);
  } else {
    transfers.value[index] = entry;
  }
}

function transferProgress(task) {
  if (!task.totalBytes) {
    return `${task.doneBytes} bytes`;
  }
  const percent = Math.round((task.doneBytes / task.totalBytes) * 100);
  return `${percent}% (${task.doneBytes}/${task.totalBytes})`;
}

async function cancelTransfer(task) {
  if (!task) {
    return;
  }
  error.value = "";
  try {
    await api.transferCancel(task.id);
  } catch (err) {
    error.value = err.message || String(err);
  }
}

async function retryTransfer(task) {
  if (!task || !task.sessionId) {
    error.value = "Missing session info for retry.";
    return;
  }
  if (!task.direction) {
    error.value = "Missing transfer direction for retry.";
    return;
  }
  error.value = "";
  try {
    if (task.direction === "download") {
      await api.transferDownload(task.sessionId, task.remotePath, task.localPath);
    } else if (task.direction === "upload") {
      await api.transferUpload(task.sessionId, task.localPath, task.remotePath);
    } else {
      error.value = `Unknown transfer direction: ${task.direction}`;
    }
  } catch (err) {
    error.value = err.message || String(err);
  }
}

async function respondHostKey(allow) {
  if (!hostKeyPrompt.value) {
    return;
  }
  const current = hostKeyPrompt.value;
  hostKeyPrompt.value = null;
  try {
    await api.hostKeyRespond(current.id, allow);
  } catch (err) {
    error.value = err.message || String(err);
  }
}

function bindEvents() {
  EventsOn("session:state", (payload) => {
    sessionByProfile[payload.profileId] = payload.sessionId;
    sessionStateById[payload.sessionId] = payload.state;
    sessionErrorById[payload.sessionId] = payload.error;
    connectingByProfile[payload.profileId] = false;
    pushEvent("session:state", payload);
  });

  EventsOn("hostkey:prompt", (payload) => {
    hostKeyPrompt.value = payload;
    pushEvent("hostkey:prompt", payload);
  });

  EventsOn("terminal:data", (payload) => {
    const chunk = payload.chunk || "";
    const term = terminalInstances.get(payload.termId);
    if (term) {
      term.write(chunk);
    } else {
      const prev = terminalPending.get(payload.termId) || "";
      terminalPending.set(payload.termId, prev + chunk);
    }
    pushEvent("terminal:data", { termId: payload.termId, chunk: chunk.slice(0, 120) });
  });

  EventsOn("terminal:exit", (payload) => {
    const message = `\r\n[exit ${payload.code}]\r\n`;
    const term = terminalInstances.get(payload.termId);
    if (term) {
      term.write(message);
    } else {
      const prev = terminalPending.get(payload.termId) || "";
      terminalPending.set(payload.termId, prev + message);
    }
    pushEvent("terminal:exit", payload);
  });

  EventsOn("transfer:progress", (payload) => {
    updateTransfer(payload);
    pushEvent("transfer:progress", payload);
  });

  EventsOn("transfer:done", (payload) => {
    updateTransfer({ ...payload, state: "done" });
    pushEvent("transfer:done", payload);
  });

  EventsOn("transfer:error", (payload) => {
    updateTransfer({ ...payload, state: "error" });
    pushEvent("transfer:error", payload);
  });
}

onMounted(async () => {
  backendReady.value = api.isAvailable();
  if (!backendReady.value) {
    pushEvent("runtime", { message: "Wails runtime not available" });
    return;
  }

  window.addEventListener("resize", handleResize);
  window.addEventListener("click", closeFileContextMenu);
  window.addEventListener("contextmenu", closeFileContextMenu);
  bindEvents();
  await reloadProfiles();
  await reloadMySQLProfiles();
  startMetrics();

  OnFileDrop((x, y, paths) => {
    if (!filesSessionId.value || !paths || paths.length === 0) {
      return;
    }
    const zone = dropZoneRef.value;
    if (zone) {
      const rect = zone.getBoundingClientRect();
      const inside = x >= rect.left && x <= rect.right && y >= rect.top && y <= rect.bottom;
      if (!inside) {
        return;
      }
    }
    uploadLocalPath.value = paths[0];
  }, false);

  try {
    const list = await api.transferListTasks();
    transfers.value = list || [];
  } catch (err) {
    pushEvent("transfer:list", { error: err.message || String(err) });
  }
});

onBeforeUnmount(() => {
  window.removeEventListener("resize", handleResize);
  window.removeEventListener("click", closeFileContextMenu);
  window.removeEventListener("contextmenu", closeFileContextMenu);
  OnFileDropOff();
  terminals.value.forEach((term) => disposeTerminal(term.id));
  stopMetrics();
});
</script>
