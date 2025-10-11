import { CreateAccountDto, AccountResponse, GetAccountResponse } from '../schemas/accounts.schemas';
export declare class AccountsService {
    private readonly logger;
    private httpClient;
    constructor();
    createAccount(dto: CreateAccountDto): Promise<AccountResponse>;
    getAccount(accountId: string): Promise<GetAccountResponse>;
    listAccounts(query: any): Promise<any>;
}
