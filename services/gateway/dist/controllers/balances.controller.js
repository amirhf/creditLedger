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
var __param = (this && this.__param) || function (paramIndex, decorator) {
    return function (target, key) { decorator(target, key, paramIndex); }
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.BalancesController = void 0;
const common_1 = require("@nestjs/common");
const swagger_1 = require("@nestjs/swagger");
const readmodel_service_1 = require("../services/readmodel.service");
const balances_schemas_1 = require("../schemas/balances.schemas");
let BalancesController = class BalancesController {
    constructor(readModelService) {
        this.readModelService = readModelService;
    }
    async getBalance(id) {
        const uuidRegex = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;
        if (!uuidRegex.test(id)) {
            throw new common_1.BadRequestException({
                error: 'VALIDATION_ERROR',
                message: 'Invalid account ID format (must be UUID)'
            });
        }
        return this.readModelService.getBalance(id);
    }
    async getStatements(id, query) {
        const uuidRegex = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;
        if (!uuidRegex.test(id)) {
            throw new common_1.BadRequestException({
                error: 'VALIDATION_ERROR',
                message: 'Invalid account ID format (must be UUID)'
            });
        }
        const parseResult = balances_schemas_1.GetStatementsQuerySchema.safeParse(query);
        if (!parseResult.success) {
            throw new common_1.BadRequestException({
                error: 'VALIDATION_ERROR',
                message: 'Invalid query parameters',
                details: parseResult.error.flatten()
            });
        }
        return this.readModelService.getStatements(id, parseResult.data);
    }
};
exports.BalancesController = BalancesController;
__decorate([
    (0, common_1.Get)(':id/balance'),
    (0, swagger_1.ApiOperation)({ summary: 'Get account balance', description: 'Retrieves the current balance for an account' }),
    (0, swagger_1.ApiParam)({ name: 'id', description: 'Account UUID', example: '123e4567-e89b-12d3-a456-426614174000' }),
    (0, swagger_1.ApiResponse)({ status: 200, description: 'Balance retrieved successfully', schema: {
            type: 'object',
            properties: {
                account_id: { type: 'string', format: 'uuid' },
                balance_minor: { type: 'integer', example: 5000, description: 'Balance in minor units' },
                currency: { type: 'string', example: 'USD' },
                updated_at: { type: 'string', format: 'date-time' }
            }
        } }),
    (0, swagger_1.ApiResponse)({ status: 400, description: 'Invalid account ID format' }),
    (0, swagger_1.ApiResponse)({ status: 404, description: 'Account not found' }),
    __param(0, (0, common_1.Param)('id')),
    __metadata("design:type", Function),
    __metadata("design:paramtypes", [String]),
    __metadata("design:returntype", Promise)
], BalancesController.prototype, "getBalance", null);
__decorate([
    (0, common_1.Get)(':id/statements'),
    (0, swagger_1.ApiOperation)({ summary: 'Get account statements', description: 'Retrieves transaction history for an account' }),
    (0, swagger_1.ApiParam)({ name: 'id', description: 'Account UUID', example: '123e4567-e89b-12d3-a456-426614174000' }),
    (0, swagger_1.ApiQuery)({ name: 'from', required: false, description: 'Start date (ISO8601)', example: '2024-01-01T00:00:00Z' }),
    (0, swagger_1.ApiQuery)({ name: 'to', required: false, description: 'End date (ISO8601)', example: '2024-12-31T23:59:59Z' }),
    (0, swagger_1.ApiQuery)({ name: 'limit', required: false, type: Number, description: 'Maximum number of entries (1-1000)', example: 100 }),
    (0, swagger_1.ApiResponse)({ status: 200, description: 'Statements retrieved successfully', schema: {
            type: 'object',
            properties: {
                statements: {
                    type: 'array',
                    items: {
                        type: 'object',
                        properties: {
                            id: { type: 'integer' },
                            account_id: { type: 'string', format: 'uuid' },
                            entry_id: { type: 'string', format: 'uuid' },
                            amount_minor: { type: 'integer' },
                            side: { type: 'string', enum: ['DEBIT', 'CREDIT'] },
                            ts: { type: 'string', format: 'date-time' }
                        }
                    }
                }
            }
        } }),
    (0, swagger_1.ApiResponse)({ status: 400, description: 'Invalid parameters' }),
    (0, swagger_1.ApiResponse)({ status: 404, description: 'Account not found' }),
    __param(0, (0, common_1.Param)('id')),
    __param(1, (0, common_1.Query)()),
    __metadata("design:type", Function),
    __metadata("design:paramtypes", [String, Object]),
    __metadata("design:returntype", Promise)
], BalancesController.prototype, "getStatements", null);
exports.BalancesController = BalancesController = __decorate([
    (0, swagger_1.ApiTags)('balances'),
    (0, common_1.Controller)('accounts'),
    __metadata("design:paramtypes", [readmodel_service_1.ReadModelService])
], BalancesController);
//# sourceMappingURL=balances.controller.js.map