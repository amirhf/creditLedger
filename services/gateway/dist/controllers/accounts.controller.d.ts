import { AccountsService } from '../services/accounts.service';
export declare class AccountsController {
    private readonly accountsService;
    constructor(accountsService: AccountsService);
    createAccount(body: any): Promise<{
        account_id?: string;
    }>;
    listAccounts(query: any): Promise<any>;
    getAccount(id: string): Promise<{
        status?: string;
        currency?: string;
        id?: string;
        created_at?: string;
    }>;
}
