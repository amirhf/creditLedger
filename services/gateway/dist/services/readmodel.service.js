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
var ReadModelService_1;
Object.defineProperty(exports, "__esModule", { value: true });
exports.ReadModelService = void 0;
const common_1 = require("@nestjs/common");
const http_client_service_1 = require("./http-client.service");
let ReadModelService = ReadModelService_1 = class ReadModelService {
    constructor() {
        this.logger = new common_1.Logger(ReadModelService_1.name);
        const baseURL = process.env.READMODEL_SERVICE_URL || 'http://localhost:7104';
        this.httpClient = new http_client_service_1.HttpClientService({ baseURL, timeout: 5000 });
        this.logger.log(`ReadModelService initialized with baseURL: ${baseURL}`);
    }
    async getBalance(accountId) {
        this.logger.debug(`Getting balance for account: ${accountId}`);
        return this.httpClient.get(`/v1/accounts/${accountId}/balance`);
    }
    async getStatements(accountId, query) {
        this.logger.debug(`Getting statements for account: ${accountId}`);
        const params = new URLSearchParams();
        if (query.from)
            params.append('from', query.from);
        if (query.to)
            params.append('to', query.to);
        if (query.limit)
            params.append('limit', query.limit.toString());
        const url = `/v1/accounts/${accountId}/statements${params.toString() ? '?' + params.toString() : ''}`;
        return this.httpClient.get(url);
    }
};
exports.ReadModelService = ReadModelService;
exports.ReadModelService = ReadModelService = ReadModelService_1 = __decorate([
    (0, common_1.Injectable)(),
    __metadata("design:paramtypes", [])
], ReadModelService);
//# sourceMappingURL=readmodel.service.js.map