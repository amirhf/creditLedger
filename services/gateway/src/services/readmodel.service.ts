import { Injectable, Logger } from '@nestjs/common';
import { HttpClientService } from './http-client.service';
import { GetBalanceResponse, GetStatementsResponse, GetStatementsQuery } from '../schemas/balances.schemas';

@Injectable()
export class ReadModelService {
  private readonly logger = new Logger(ReadModelService.name);
  private httpClient: HttpClientService;

  constructor() {
    const baseURL = process.env.READMODEL_SERVICE_URL || 'http://localhost:7104';
    this.httpClient = new HttpClientService({ baseURL, timeout: 5000 });
    this.logger.log(`ReadModelService initialized with baseURL: ${baseURL}`);
  }

  async getBalance(accountId: string): Promise<GetBalanceResponse> {
    this.logger.debug(`Getting balance for account: ${accountId}`);
    return this.httpClient.get<GetBalanceResponse>(`/v1/accounts/${accountId}/balance`);
  }

  async getStatements(accountId: string, query: GetStatementsQuery): Promise<GetStatementsResponse> {
    this.logger.debug(`Getting statements for account: ${accountId}`);
    const params = new URLSearchParams();
    if (query.from) params.append('from', query.from);
    if (query.to) params.append('to', query.to);
    if (query.limit) params.append('limit', query.limit.toString());
    
    const url = `/v1/accounts/${accountId}/statements${params.toString() ? '?' + params.toString() : ''}`;
    return this.httpClient.get<GetStatementsResponse>(url);
  }
}
