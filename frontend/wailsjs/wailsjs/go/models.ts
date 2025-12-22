export namespace app {
	
	export class ConvertRequest {
	    amount: number;
	    currency: string;
	    date: string;
	
	    static createFrom(source: any = {}) {
	        return new ConvertRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.amount = source["amount"];
	        this.currency = source["currency"];
	        this.date = source["date"];
	    }
	}
	export class ConvertResponse {
	    success: boolean;
	    result: string;
	    error: string;
	
	    static createFrom(source: any = {}) {
	        return new ConvertResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.result = source["result"];
	        this.error = source["error"];
	    }
	}
	export class RateResponse {
	    success: boolean;
	    rate: number;
	    error: string;
	
	    static createFrom(source: any = {}) {
	        return new RateResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.rate = source["rate"];
	        this.error = source["error"];
	    }
	}

}

