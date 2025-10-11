import { Injectable, Logger } from '@nestjs/common';
import { HttpClientService } from './http-client.service';
import { CreateTransferDto, TransferResponse, GetTransferResponse } from '../schemas/transfers.schemas';

@Injectable()
export class OrchestratorService {
  private readonly logger = new Logger(OrchestratorService.name);
  private httpClient: HttpClientService;

  constructor() {
    const baseURL = process.env.ORCHESTRATOR_SERVICE_URL || 'http://localhost:7103';
    this.httpClient = new HttpClientService({ baseURL, timeout: 10000 });
    this.logger.log(`OrchestratorService initialized with baseURL: ${baseURL}`);
  }

  async createTransfer(dto: CreateTransferDto): Promise<TransferResponse> {
    this.logger.debug(`Creating transfer from ${dto.from_account_id} to ${dto.to_account_id}`);
    return this.httpClient.post<TransferResponse>('/v1/transfers', dto);
  }

  async getTransfer(transferId: string): Promise<GetTransferResponse> {
    this.logger.debug(`Getting transfer: ${transferId}`);
    return this.httpClient.get<GetTransferResponse>(`/v1/transfers/${transferId}`);
  }

  async listTransfers(query: any): Promise<any> {
    this.logger.debug(`Listing transfers with filters: ${JSON.stringify(query)}`);
    const params = new URLSearchParams();
    if (query.from_account_id) params.append('from_account_id', query.from_account_id);
    if (query.to_account_id) params.append('to_account_id', query.to_account_id);
    if (query.status) params.append('status', query.status);
    if (query.currency) params.append('currency', query.currency);
    if (query.limit) params.append('limit', query.limit.toString());
    if (query.offset) params.append('offset', query.offset.toString());
    
    const url = `/v1/transfers${params.toString() ? '?' + params.toString() : ''}`;
    return this.httpClient.get<any>(url);
  }
}
