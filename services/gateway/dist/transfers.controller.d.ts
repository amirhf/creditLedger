import { OrchestratorService } from './services/orchestrator.service';
export declare class TransfersController {
    private readonly orchestratorService;
    constructor(orchestratorService: OrchestratorService);
    createTransfer(body: any): Promise<{
        status?: string;
        transfer_id?: string;
    }>;
    listTransfers(query: any): Promise<any>;
    getTransfer(id: string): Promise<{
        status?: string;
        from_account_id?: string;
        to_account_id?: string;
        amount_minor?: number;
        currency?: string;
        idempotency_key?: string;
        id?: string;
        created_at?: string;
    }>;
}
