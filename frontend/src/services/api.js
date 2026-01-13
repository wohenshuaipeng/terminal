function getApi() {
  if (window.go && window.go.app && window.go.app.App) {
    return window.go.app.App;
  }
  return null;
}

function requireApi() {
  const api = getApi();
  if (!api) {
    throw new Error("Wails runtime is not available");
  }
  return api;
}

export function isAvailable() {
  return !!getApi();
}

export async function profilesList() {
  return await requireApi().ProfilesList();
}

export async function profilesSave(profile) {
  return await requireApi().ProfilesSave(profile);
}

export async function profilesDelete(id) {
  return await requireApi().ProfilesDelete(id);
}

export async function sessionConnect(profileId) {
  return await requireApi().SessionConnect(profileId);
}

export async function sessionDisconnect(sessionId) {
  return await requireApi().SessionDisconnect(sessionId);
}

export async function sessionStatus(sessionId) {
  return await requireApi().SessionStatus(sessionId);
}

export async function terminalOpen(sessionId, cols, rows) {
  return await requireApi().TerminalOpen(sessionId, cols, rows);
}

export async function terminalWrite(termId, data) {
  return await requireApi().TerminalWrite(termId, data);
}

export async function terminalResize(termId, cols, rows) {
  return await requireApi().TerminalResize(termId, cols, rows);
}

export async function terminalClose(termId) {
  return await requireApi().TerminalClose(termId);
}

export async function filesList(sessionId, path) {
  return await requireApi().FilesList(sessionId, path);
}

export async function filesStat(sessionId, path) {
  return await requireApi().FilesStat(sessionId, path);
}

export async function filesMkdir(sessionId, path) {
  return await requireApi().FilesMkdir(sessionId, path);
}

export async function filesRemove(sessionId, path, recursive) {
  return await requireApi().FilesRemove(sessionId, path, recursive);
}

export async function filesRename(sessionId, fromPath, toPath) {
  return await requireApi().FilesRename(sessionId, fromPath, toPath);
}

export async function dialogOpenFile(title) {
  return await requireApi().DialogOpenFile(title);
}

export async function dialogSaveFile(title, filename) {
  return await requireApi().DialogSaveFile(title, filename);
}

export async function hostKeyRespond(requestId, allow) {
  return await requireApi().HostKeyRespond(requestId, allow);
}

export async function transferDownload(sessionId, remotePath, localPath) {
  return await requireApi().TransferDownload(sessionId, remotePath, localPath);
}

export async function transferUpload(sessionId, localPath, remotePath) {
  return await requireApi().TransferUpload(sessionId, localPath, remotePath);
}

export async function transferCancel(taskId) {
  return await requireApi().TransferCancel(taskId);
}

export async function transferListTasks() {
  return await requireApi().TransferListTasks();
}

export async function credentialsSetPassword(profileId, password) {
  return await requireApi().CredentialsSetPassword(profileId, password);
}

export async function credentialsSetPrivateKeyPassphrase(profileId, passphrase) {
  return await requireApi().CredentialsSetPrivateKeyPassphrase(profileId, passphrase);
}

export async function credentialsDelete(profileId) {
  return await requireApi().CredentialsDelete(profileId);
}

export async function mysqlProfilesList() {
  return await requireApi().MySQLProfilesList();
}

export async function mysqlProfilesSave(profile) {
  return await requireApi().MySQLProfilesSave(profile);
}

export async function mysqlProfilesDelete(id) {
  return await requireApi().MySQLProfilesDelete(id);
}

export async function mysqlConnect(profileId) {
  return await requireApi().MySQLConnect(profileId);
}

export async function mysqlDisconnect(profileId) {
  return await requireApi().MySQLDisconnect(profileId);
}

export async function mysqlStatus(profileId) {
  return await requireApi().MySQLStatus(profileId);
}

export async function mysqlListDatabases(profileId) {
  return await requireApi().MySQLListDatabases(profileId);
}

export async function mysqlListTables(profileId, database) {
  return await requireApi().MySQLListTables(profileId, database);
}

export async function mysqlTableSchema(profileId, database, table) {
  return await requireApi().MySQLTableSchema(profileId, database, table);
}

export async function mysqlPreviewTable(profileId, database, table, filter, orderBy, orderDir, limit, offset) {
  return await requireApi().MySQLPreviewTable(
    profileId,
    database,
    table,
    filter,
    orderBy,
    orderDir,
    limit,
    offset
  );
}

export async function mysqlQuery(profileId, database, query) {
  return await requireApi().MySQLQuery(profileId, database, query);
}

export async function mysqlCreateDatabase(profileId, name) {
  return await requireApi().MySQLCreateDatabase(profileId, name);
}

export async function mysqlDropDatabase(profileId, name) {
  return await requireApi().MySQLDropDatabase(profileId, name);
}

export async function mysqlDropTable(profileId, database, table) {
  return await requireApi().MySQLDropTable(profileId, database, table);
}

export async function mysqlCredentialsSetPassword(profileId, password) {
  return await requireApi().MySQLCredentialsSetPassword(profileId, password);
}

export async function mysqlCredentialsDelete(profileId) {
  return await requireApi().MySQLCredentialsDelete(profileId);
}

export async function systemStats() {
  return await requireApi().SystemStats();
}
