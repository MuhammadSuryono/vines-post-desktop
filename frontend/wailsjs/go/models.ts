export namespace main {
	
	export class AppConfig {
	    mode: string;
	    remote_url: string;
	    version: string;
	    print_mode: string;
	
	    static createFrom(source: any = {}) {
	        return new AppConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.mode = source["mode"];
	        this.remote_url = source["remote_url"];
	        this.version = source["version"];
	        this.print_mode = source["print_mode"];
	    }
	}
	export class GitHubAsset {
	    name: string;
	    browser_download_url: string;
	
	    static createFrom(source: any = {}) {
	        return new GitHubAsset(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.browser_download_url = source["browser_download_url"];
	    }
	}
	export class GitHubRelease {
	    tag_name: string;
	    name: string;
	    html_url: string;
	    assets: GitHubAsset[];
	
	    static createFrom(source: any = {}) {
	        return new GitHubRelease(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.tag_name = source["tag_name"];
	        this.name = source["name"];
	        this.html_url = source["html_url"];
	        this.assets = this.convertValues(source["assets"], GitHubAsset);
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
	export class HeaderLine {
	    header: string;
	    address: string;
	    city: string;
	    phone_number: string;
	    portal_code: string;
	    use_dash: boolean;
	
	    static createFrom(source: any = {}) {
	        return new HeaderLine(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.header = source["header"];
	        this.address = source["address"];
	        this.city = source["city"];
	        this.phone_number = source["phone_number"];
	        this.portal_code = source["portal_code"];
	        this.use_dash = source["use_dash"];
	    }
	}
	export class ItemLine {
	    item_name: string;
	    total_unit: string;
	    price: string;
	    total_price: string;
	
	    static createFrom(source: any = {}) {
	        return new ItemLine(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.item_name = source["item_name"];
	        this.total_unit = source["total_unit"];
	        this.price = source["price"];
	        this.total_price = source["total_price"];
	    }
	}
	export class PrinterLine {
	    header_line: HeaderLine;
	    description_line: struct { Data main.;
	    item_line: ItemLine[];
	    others: struct { Data main.receiptFields "json:\"data\""; UseDash bool "json:\"use_dash\"" }[];
	    notes: string;
	    printer_name: string;
	
	    static createFrom(source: any = {}) {
	        return new PrinterLine(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.header_line = this.convertValues(source["header_line"], HeaderLine);
	        this.description_line = this.convertValues(source["description_line"], Object);
	        this.item_line = this.convertValues(source["item_line"], ItemLine);
	        this.others = this.convertValues(source["others"], struct { Data main.receiptFields "json:\"data\""; UseDash bool "json:\"use_dash\"" });
	        this.notes = source["notes"];
	        this.printer_name = source["printer_name"];
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
	export class receiptField {
	    Key: string;
	    Value: string;
	
	    static createFrom(source: any = {}) {
	        return new receiptField(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Key = source["Key"];
	        this.Value = source["Value"];
	    }
	}

}

export namespace struct { Data main {
	
	export class  {
	    data: main.receiptField[];
	    use_dash: boolean;
	
	    static createFrom(source: any = {}) {
	        return new (source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.data = this.convertValues(source["data"], main.receiptField);
	        this.use_dash = source["use_dash"];
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

