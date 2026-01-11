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
	    sourceAmount: number;
	    targetAmountRUB: number;
	    rate: number;
	    currency: string;
	    currencySymbol: string;
	    requestedDate: string;
	    actualDate: string;
	
	    static createFrom(source: any = {}) {
	        return new ConvertResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.result = source["result"];
	        this.error = source["error"];
	        this.sourceAmount = source["sourceAmount"];
	        this.targetAmountRUB = source["targetAmountRUB"];
	        this.rate = source["rate"];
	        this.currency = source["currency"];
	        this.currencySymbol = source["currencySymbol"];
	        this.requestedDate = source["requestedDate"];
	        this.actualDate = source["actualDate"];
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
	export class SendStarResponse {
	    success: boolean;
	    error: string;
	
	    static createFrom(source: any = {}) {
	        return new SendStarResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.error = source["error"];
	    }
	}

}

