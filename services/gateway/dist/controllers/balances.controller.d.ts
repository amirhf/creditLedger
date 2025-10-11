import { ReadModelService } from '../services/readmodel.service';
export declare class BalancesController {
    private readonly readModelService;
    constructor(readModelService: ReadModelService);
    getBalance(id: string): Promise<{
        currency?: string;
        account_id?: string;
        balance_minor?: number;
        updated_at?: string;
    }>;
    getStatements(id: string, query: any): Promise<{
        statements?: {
            amount_minor?: number;
            id?: number;
            account_id?: string;
            entry_id?: string;
            side?: string;
            ts?: string;
        }[];
    }>;
}
