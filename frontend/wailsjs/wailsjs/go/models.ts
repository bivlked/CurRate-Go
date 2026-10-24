export namespace app {
	
	export class ConvertRequest {
	    amount: number;
	    currency: string;
	    date: string;
	    refDate: string;
	
	    static createFrom(source: any = {}) {
	        return new ConvertRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.amount = source["amount"];
	        this.currency = source["currency"];
	        this.date = source["date"];
	        this.refDate = source["refDate"];
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
	    hasReference: boolean;
	    refRate: number;
	    refAmountRUB: number;
	    refDiffRUB: number;
	    refRequestedDate: string;
	    refActualDate: string;
	    refError: string;
	
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
	        this.hasReference = source["hasReference"];
	        this.refRate = source["refRate"];
	        this.refAmountRUB = source["refAmountRUB"];
	        this.refDiffRUB = source["refDiffRUB"];
	        this.refRequestedDate = source["refRequestedDate"];
	        this.refActualDate = source["refActualDate"];
	        this.refError = source["refError"];
	    }
	}
	export class RateResponse {
	    success: boolean;
	    rate: number;
	    actualDate: string;
	    error: string;
	
	    static createFrom(source: any = {}) {
	        return new RateResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.rate = source["rate"];
	        this.actualDate = source["actualDate"];
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

