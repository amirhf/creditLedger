"use strict";
var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
var __metadata = (this && this.__metadata) || function (k, v) {
    if (typeof Reflect === "object" && typeof Reflect.metadata === "function") return Reflect.metadata(k, v);
};
var OrchestratorService_1;
Object.defineProperty(exports, "__esModule", { value: true });
exports.OrchestratorService = void 0;
const common_1 = require("@nestjs/common");
const http_client_service_1 = require("./http-client.service");
let OrchestratorService = OrchestratorService_1 = class OrchestratorService {
    constructor() {
        this.logger = new common_1.Logger(OrchestratorService_1.name);
        const baseURL = process.env.ORCHESTRATOR_SERVICE_URL || 'http://localhost:7103';
        this.httpClient = new http_client_service_1.HttpClientService({ baseURL, timeout: 10000 });
        this.logger.log(`OrchestratorService initialized with baseURL: ${baseURL}`);
    }
    async createTransfer(dto) {
        this.logger.debug(`Creating transfer from ${dto.from_account_id} to ${dto.to_account_id}`);
        return this.httpClient.post('/v1/transfers', dto);
    }
    async getTransfer(transferId) {
        this.logger.debug(`Getting transfer: ${transferId}`);
        return this.httpClient.get(`/v1/transfers/${transferId}`);
    }
    async listTransfers(query) {
        this.logger.debug(`Listing transfers with filters: ${JSON.stringify(query)}`);
        const params = new URLSearchParams();
        if (query.from_account_id)
            params.append('from_account_id', query.from_account_id);
        if (query.to_account_id)
            params.append('to_account_id', query.to_account_id);
        if (query.status)
            params.append('status', query.status);
        if (query.currency)
            params.append('currency', query.currency);
        if (query.limit)
            params.append('limit', query.limit.toString());
        if (query.offset)
            params.append('offset', query.offset.toString());
        const url = `/v1/transfers${params.toString() ? '?' + params.toString() : ''}`;
        return this.httpClient.get(url);
    }
};
exports.OrchestratorService = OrchestratorService;
exports.OrchestratorService = OrchestratorService = OrchestratorService_1 = __decorate([
    (0, common_1.Injectable)(),
    __metadata("design:paramtypes", [])
], OrchestratorService);
//# sourceMappingURL=orchestrator.service.js.map