import { CreateTransferDto, TransferResponse, GetTransferResponse } from '../schemas/transfers.schemas';
export declare class OrchestratorService {
    private readonly logger;
    private httpClient;
    constructor();
    createTransfer(dto: CreateTransferDto): Promise<TransferResponse>;
    getTransfer(transferId: string): Promise<GetTransferResponse>;
    listTransfers(query: any): Promise<any>;
}
