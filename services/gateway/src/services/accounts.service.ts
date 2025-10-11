import { Injectable, Logger } from '@nestjs/common';
import { HttpClientService } from './http-client.service';
import { CreateAccountDto, AccountResponse, GetAccountResponse } from '../schemas/accounts.schemas';

@Injectable()
export class AccountsService {
  private readonly logger = new Logger(AccountsService.name);
  private httpClient: HttpClientService;

  constructor() {
    const baseURL = process.env.ACCOUNTS_SERVICE_URL || 'http://localhost:7101';
    this.httpClient = new HttpClientService({ baseURL, timeout: 5000 });
    this.logger.log(`AccountsService initialized with baseURL: ${baseURL}`);
  }

  async createAccount(dto: CreateAccountDto): Promise<AccountResponse> {
    this.logger.debug(`Creating account with currency: ${dto.currency}`);
    return this.httpClient.post<AccountResponse>('/v1/accounts', dto);
  }

  async getAccount(accountId: string): Promise<GetAccountResponse> {
    this.logger.debug(`Getting account: ${accountId}`);
    return this.httpClient.get<GetAccountResponse>(`/v1/accounts/${accountId}`);
  }

  async listAccounts(query: any): Promise<any> {
    this.logger.debug(`Listing accounts with filters: ${JSON.stringify(query)}`);
    const params = new URLSearchParams();
    if (query.currency) params.append('currency', query.currency);
    if (query.status) params.append('status', query.status);
    if (query.limit) params.append('limit', query.limit.toString());
    if (query.offset) params.append('offset', query.offset.toString());
    
    const url = `/v1/accounts${params.toString() ? '?' + params.toString() : ''}`;
    return this.httpClient.get<any>(url);
  }
}
