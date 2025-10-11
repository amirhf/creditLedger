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
var AccountsService_1;
Object.defineProperty(exports, "__esModule", { value: true });
exports.AccountsService = void 0;
const common_1 = require("@nestjs/common");
const http_client_service_1 = require("./http-client.service");
let AccountsService = AccountsService_1 = class AccountsService {
    constructor() {
        this.logger = new common_1.Logger(AccountsService_1.name);
        const baseURL = process.env.ACCOUNTS_SERVICE_URL || 'http://localhost:7101';
        this.httpClient = new http_client_service_1.HttpClientService({ baseURL, timeout: 5000 });
        this.logger.log(`AccountsService initialized with baseURL: ${baseURL}`);
    }
    async createAccount(dto) {
        this.logger.debug(`Creating account with currency: ${dto.currency}`);
        return this.httpClient.post('/v1/accounts', dto);
    }
    async getAccount(accountId) {
        this.logger.debug(`Getting account: ${accountId}`);
        return this.httpClient.get(`/v1/accounts/${accountId}`);
    }
    async listAccounts(query) {
        this.logger.debug(`Listing accounts with filters: ${JSON.stringify(query)}`);
        const params = new URLSearchParams();
        if (query.currency)
            params.append('currency', query.currency);
        if (query.status)
            params.append('status', query.status);
        if (query.limit)
            params.append('limit', query.limit.toString());
        if (query.offset)
            params.append('offset', query.offset.toString());
        const url = `/v1/accounts${params.toString() ? '?' + params.toString() : ''}`;
        return this.httpClient.get(url);
    }
};
exports.AccountsService = AccountsService;
exports.AccountsService = AccountsService = AccountsService_1 = __decorate([
    (0, common_1.Injectable)(),
    __metadata("design:paramtypes", [])
], AccountsService);
//# sourceMappingURL=accounts.service.js.map