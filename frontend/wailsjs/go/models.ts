export namespace config {
	
	export class Config {
	    namespaces: string[];
	    portRangeStart: number;
	    portRangeEnd: number;
	    grpcPorts: number[];
	    nodePortHost: string;
	    tokenEndpoint: string;
	    clientId: string;
	    serviceExcludePatterns: string[];
	
	    static createFrom(source: any = {}) {
	        return new Config(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.namespaces = source["namespaces"];
	        this.portRangeStart = source["portRangeStart"];
	        this.portRangeEnd = source["portRangeEnd"];
	        this.grpcPorts = source["grpcPorts"];
	        this.nodePortHost = source["nodePortHost"];
	        this.tokenEndpoint = source["tokenEndpoint"];
	        this.clientId = source["clientId"];
	        this.serviceExcludePatterns = source["serviceExcludePatterns"];
	    }
	}

}

export namespace main {
	
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

