export namespace metrics {
	
	export class CPUStats {
	    total: number;
	    perCore: number[];
	
	    static createFrom(source: any = {}) {
	        return new CPUStats(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.total = source["total"];
	        this.perCore = source["perCore"];
	    }
	}
	export class MemoryStats {
	    total: number;
	    used: number;
	    free: number;
	    usedPercent: number;
	
	    static createFrom(source: any = {}) {
	        return new MemoryStats(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.total = source["total"];
	        this.used = source["used"];
	        this.free = source["free"];
	        this.usedPercent = source["usedPercent"];
	    }
	}
	export class Stats {
	    timestamp: number;
	    cpu: CPUStats;
	    memory: MemoryStats;
	
	    static createFrom(source: any = {}) {
	        return new Stats(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.timestamp = source["timestamp"];
	        this.cpu = this.convertValues(source["cpu"], CPUStats);
	        this.memory = this.convertValues(source["memory"], MemoryStats);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace mysql {
	
	export class Column {
	    name: string;
	    type: string;
	    nullable: string;
	    key: string;
	    default: string;
	    extra: string;
	
	    static createFrom(source: any = {}) {
	        return new Column(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.type = source["type"];
	        this.nullable = source["nullable"];
	        this.key = source["key"];
	        this.default = source["default"];
	        this.extra = source["extra"];
	    }
	}
	export class PreviewResult {
	    columns: string[];
	    rows: string[][];
	    truncated: boolean;
	
	    static createFrom(source: any = {}) {
	        return new PreviewResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.columns = source["columns"];
	        this.rows = source["rows"];
	        this.truncated = source["truncated"];
	    }
	}
	export class Profile {
	    id: string;
	    name: string;
	    host: string;
	    port: number;
	    username: string;
	    database: string;
	    connectionType: string;
	    sshProfileId: string;
	    useKeyring: boolean;
	    useTls: boolean;
	    tlsCaPath: string;
	    tlsCertPath: string;
	    tlsKeyPath: string;
	    tlsSkipVerify: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Profile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.host = source["host"];
	        this.port = source["port"];
	        this.username = source["username"];
	        this.database = source["database"];
	        this.connectionType = source["connectionType"];
	        this.sshProfileId = source["sshProfileId"];
	        this.useKeyring = source["useKeyring"];
	        this.useTls = source["useTls"];
	        this.tlsCaPath = source["tlsCaPath"];
	        this.tlsCertPath = source["tlsCertPath"];
	        this.tlsKeyPath = source["tlsKeyPath"];
	        this.tlsSkipVerify = source["tlsSkipVerify"];
	    }
	}
	export class QueryResult {
	    kind: string;
	    columns: string[];
	    rows: string[][];
	    affectedRows: number;
	    lastInsertId: number;
	    durationMs: number;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new QueryResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.kind = source["kind"];
	        this.columns = source["columns"];
	        this.rows = source["rows"];
	        this.affectedRows = source["affectedRows"];
	        this.lastInsertId = source["lastInsertId"];
	        this.durationMs = source["durationMs"];
	        this.message = source["message"];
	    }
	}
	export class Status {
	    state: string;
	    lastError: string;
	
	    static createFrom(source: any = {}) {
	        return new Status(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.state = source["state"];
	        this.lastError = source["lastError"];
	    }
	}

}

export namespace profiles {
	
	export class Profile {
	    id: string;
	    name: string;
	    group: string;
	    host: string;
	    port: number;
	    username: string;
	    authType: string;
	    privateKeyPath: string;
	    useKeyring: boolean;
	    knownHostsPolicy: string;
	
	    static createFrom(source: any = {}) {
	        return new Profile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.group = source["group"];
	        this.host = source["host"];
	        this.port = source["port"];
	        this.username = source["username"];
	        this.authType = source["authType"];
	        this.privateKeyPath = source["privateKeyPath"];
	        this.useKeyring = source["useKeyring"];
	        this.knownHostsPolicy = source["knownHostsPolicy"];
	    }
	}

}

export namespace session {
	
	export class Status {
	    state: string;
	    lastError: string;
	
	    static createFrom(source: any = {}) {
	        return new Status(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.state = source["state"];
	        this.lastError = source["lastError"];
	    }
	}

}

export namespace sftp {
	
	export class FileEntry {
	    name: string;
	    path: string;
	    isDir: boolean;
	    size: number;
	    mode: string;
	    mtime: number;
	
	    static createFrom(source: any = {}) {
	        return new FileEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.path = source["path"];
	        this.isDir = source["isDir"];
	        this.size = source["size"];
	        this.mode = source["mode"];
	        this.mtime = source["mtime"];
	    }
	}

}

export namespace transfer {
	
	export class Task {
	    id: string;
	    sessionId: string;
	    localPath: string;
	    remotePath: string;
	    totalBytes: number;
	    doneBytes: number;
	    state: string;
	    direction: string;
	
	    static createFrom(source: any = {}) {
	        return new Task(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.sessionId = source["sessionId"];
	        this.localPath = source["localPath"];
	        this.remotePath = source["remotePath"];
	        this.totalBytes = source["totalBytes"];
	        this.doneBytes = source["doneBytes"];
	        this.state = source["state"];
	        this.direction = source["direction"];
	    }
	}

}

