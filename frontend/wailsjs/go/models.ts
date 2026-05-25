export namespace config {
	
	export class AuthEndpoint {
	    cluster: string;
	    namespace: string;
	    tokenUrl: string;
	    realmResolverUrl: string;
	    realmJsonPath: string;
	
	    static createFrom(source: any = {}) {
	        return new AuthEndpoint(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.cluster = source["cluster"];
	        this.namespace = source["namespace"];
	        this.tokenUrl = source["tokenUrl"];
	        this.realmResolverUrl = source["realmResolverUrl"];
	        this.realmJsonPath = source["realmJsonPath"];
	    }
	}
	export class Config {
	    namespaces: string[];
	    portRangeStart: number;
	    portRangeEnd: number;
	    grpcPorts: number[];
	    discoveryConcurrency: number;
	    nodePortHost: string;
	    authProvider: string;
	    clientId: string;
	    authEndpoints: AuthEndpoint[];
	    serviceExcludePatterns: string[];
	    parentClaimMap: Record<string, string>;
	
	    static createFrom(source: any = {}) {
	        return new Config(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.namespaces = source["namespaces"];
	        this.portRangeStart = source["portRangeStart"];
	        this.portRangeEnd = source["portRangeEnd"];
	        this.grpcPorts = source["grpcPorts"];
	        this.discoveryConcurrency = source["discoveryConcurrency"];
	        this.nodePortHost = source["nodePortHost"];
	        this.authProvider = source["authProvider"];
	        this.clientId = source["clientId"];
	        this.authEndpoints = this.convertValues(source["authEndpoints"], AuthEndpoint);
	        this.serviceExcludePatterns = source["serviceExcludePatterns"];
	        this.parentClaimMap = source["parentClaimMap"];
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

export namespace main {
	
	export class AuthEndpointInfo {
	    found: boolean;
	    cluster: string;
	    namespace: string;
	    tokenUrl: string;
	    realmResolverUrl: string;
	    realmJsonPath: string;
	    needsSubdomain: boolean;
	
	    static createFrom(source: any = {}) {
	        return new AuthEndpointInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.found = source["found"];
	        this.cluster = source["cluster"];
	        this.namespace = source["namespace"];
	        this.tokenUrl = source["tokenUrl"];
	        this.realmResolverUrl = source["realmResolverUrl"];
	        this.realmJsonPath = source["realmJsonPath"];
	        this.needsSubdomain = source["needsSubdomain"];
	    }
	}
	export class AuthState {
	    loggedIn: boolean;
	    username: string;
	    accessToken: string;
	    expiresAt: number;
	    refreshExpiresAt: number;
	
	    static createFrom(source: any = {}) {
	        return new AuthState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.loggedIn = source["loggedIn"];
	        this.username = source["username"];
	        this.accessToken = source["accessToken"];
	        this.expiresAt = source["expiresAt"];
	        this.refreshExpiresAt = source["refreshExpiresAt"];
	    }
	}
	export class ClusterInfo {
	    name: string;
	    context: string;
	    server: string;
	
	    static createFrom(source: any = {}) {
	        return new ClusterInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.context = source["context"];
	        this.server = source["server"];
	    }
	}
	export class GrpcServiceInfo {
	    displayName: string;
	    serviceName: string;
	    localAddress: string;
	    namespace: string;
	    k8sService: string;
	    viaNodePort: boolean;
	
	    static createFrom(source: any = {}) {
	        return new GrpcServiceInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.displayName = source["displayName"];
	        this.serviceName = source["serviceName"];
	        this.localAddress = source["localAddress"];
	        this.namespace = source["namespace"];
	        this.k8sService = source["k8sService"];
	        this.viaNodePort = source["viaNodePort"];
	    }
	}
	export class MethodInfo {
	    name: string;
	    requestType: string;
	    responseType: string;
	
	    static createFrom(source: any = {}) {
	        return new MethodInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.requestType = source["requestType"];
	        this.responseType = source["responseType"];
	    }
	}
	export class NamespaceInfo {
	    name: string;
	    allowed: boolean;
	    reason?: string;
	
	    static createFrom(source: any = {}) {
	        return new NamespaceInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.allowed = source["allowed"];
	        this.reason = source["reason"];
	    }
	}

}

