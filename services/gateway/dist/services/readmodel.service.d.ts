import { GetBalanceResponse, GetStatementsResponse, GetStatementsQuery } from '../schemas/balances.schemas';
export declare class ReadModelService {
    private readonly logger;
    private httpClient;
    constructor();
    getBalance(accountId: string): Promise<GetBalanceResponse>;
    getStatements(accountId: string, query: GetStatementsQuery): Promise<GetStatementsResponse>;
}
