export namespace log {
	
	export class LogItem {
	    Level: string;
	    Message: string;
	    // Go type: time
	    Time: any;
	
	    static createFrom(source: any = {}) {
	        return new LogItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Level = source["Level"];
	        this.Message = source["Message"];
	        this.Time = this.convertValues(source["Time"], null);
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
	
	export class AppService {
	    Name: string;
	    IP: string;
	    Ports: number[];
	
	    static createFrom(source: any = {}) {
	        return new AppService(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Name = source["Name"];
	        this.IP = source["IP"];
	        this.Ports = source["Ports"];
	    }
	}
	export class AppNsService {
	    Namespace: string;
	    Services: AppService[];
	
	    static createFrom(source: any = {}) {
	        return new AppNsService(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Namespace = source["Namespace"];
	        this.Services = this.convertValues(source["Services"], AppService);
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
	
	export class VPNConfig {
	    Server: string;
	    User: string;
	    Port: number;
	    Password: string;
	    RefreshInterval: number;
	
	    static createFrom(source: any = {}) {
	        return new VPNConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Server = source["Server"];
	        this.User = source["User"];
	        this.Port = source["Port"];
	        this.Password = source["Password"];
	        this.RefreshInterval = source["RefreshInterval"];
	    }
	}

}

