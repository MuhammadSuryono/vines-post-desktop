export namespace main {
	
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
	export class  {
	    data: Record<string, string>;
	    use_dash: boolean;
	
	    static createFrom(source: any = {}) {
	        return new (source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.data = source["data"];
	        this.use_dash = source["use_dash"];
	    }
	}
	export class PrinterLine {
	    header_line: HeaderLine;
	    // Go type: struct { Data map[string]string "json:\"data\""; UseDash bool "json:\"use_dash\"" }
	    description_line: any;
	    item_line: ItemLine[];
	    others: [];
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
	        this.others = this.convertValues(source["others"], );
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

}

