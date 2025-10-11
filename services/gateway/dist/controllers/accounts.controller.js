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
exports.AccountsController = void 0;
const common_1 = require("@nestjs/common");
const swagger_1 = require("@nestjs/swagger");
const accounts_service_1 = require("../services/accounts.service");
const accounts_schemas_1 = require("../schemas/accounts.schemas");
let AccountsController = class AccountsController {
    constructor(accountsService) {
        this.accountsService = accountsService;
    }
    async createAccount(body) {
        const parseResult = accounts_schemas_1.CreateAccountSchema.safeParse(body);
        if (!parseResult.success) {
            throw new common_1.BadRequestException({
                error: 'VALIDATION_ERROR',
                message: 'Invalid request body',
                details: parseResult.error.flatten()
            });
        }
        const dto = parseResult.data;
        return this.accountsService.createAccount(dto);
    }
    async listAccounts(query) {
        return this.accountsService.listAccounts(query);
    }
    async getAccount(id) {
        const uuidRegex = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;
        if (!uuidRegex.test(id)) {
            throw new common_1.BadRequestException({
                error: 'VALIDATION_ERROR',
                message: 'Invalid account ID format (must be UUID)'
            });
        }
        return this.accountsService.getAccount(id);
    }
};
exports.AccountsController = AccountsController;
__decorate([
    (0, common_1.Post)(),
    (0, common_1.HttpCode)(common_1.HttpStatus.CREATED),
    (0, swagger_1.ApiOperation)({ summary: 'Create a new account', description: 'Creates a new account with the specified currency' }),
    (0, swagger_1.ApiBody)({
        description: 'Account creation request',
        schema: {
            type: 'object',
            required: ['currency'],
            properties: {
                currency: {
                    type: 'string',
                    description: 'Three-letter ISO currency code (e.g., USD, EUR, GBP)',
                    example: 'USD',
                    pattern: '^[A-Z]{3}$',
                    minLength: 3,
                    maxLength: 3
                }
            }
        }
    }),
    (0, swagger_1.ApiResponse)({ status: 201, description: 'Account created successfully', schema: {
            type: 'object',
            properties: {
                account_id: { type: 'string', format: 'uuid', example: '123e4567-e89b-12d3-a456-426614174000' },
                currency: { type: 'string', example: 'USD' },
                status: { type: 'string', example: 'ACTIVE' }
            }
        } }),
    (0, swagger_1.ApiResponse)({ status: 400, description: 'Invalid request body' }),
    __param(0, (0, common_1.Body)()),
    __metadata("design:type", Function),
    __metadata("design:paramtypes", [Object]),
    __metadata("design:returntype", Promise)
], AccountsController.prototype, "createAccount", null);
__decorate([
    (0, common_1.Get)(),
    (0, swagger_1.ApiOperation)({ summary: 'List accounts', description: 'Retrieves a list of accounts with optional filtering' }),
    (0, swagger_1.ApiQuery)({ name: 'currency', required: false, description: 'Filter by currency code', example: 'USD' }),
    (0, swagger_1.ApiQuery)({ name: 'status', required: false, description: 'Filter by status', example: 'ACTIVE' }),
    (0, swagger_1.ApiQuery)({ name: 'limit', required: false, type: Number, description: 'Maximum number of results (1-100)', example: 20 }),
    (0, swagger_1.ApiQuery)({ name: 'offset', required: false, type: Number, description: 'Number of results to skip', example: 0 }),
    (0, swagger_1.ApiResponse)({ status: 200, description: 'Accounts retrieved successfully', schema: {
            type: 'object',
            properties: {
                accounts: {
                    type: 'array',
                    items: {
                        type: 'object',
                        properties: {
                            id: { type: 'string', format: 'uuid' },
                            currency: { type: 'string', example: 'USD' },
                            status: { type: 'string', example: 'ACTIVE' },
                            created_at: { type: 'string', format: 'date-time' }
                        }
                    }
                },
                total: { type: 'integer', description: 'Total number of accounts matching filters' },
                limit: { type: 'integer' },
                offset: { type: 'integer' }
            }
        } }),
    __param(0, (0, common_1.Query)()),
    __metadata("design:type", Function),
    __metadata("design:paramtypes", [Object]),
    __metadata("design:returntype", Promise)
], AccountsController.prototype, "listAccounts", null);
__decorate([
    (0, common_1.Get)(':id'),
    (0, swagger_1.ApiOperation)({ summary: 'Get account details', description: 'Retrieves account information by ID' }),
    (0, swagger_1.ApiParam)({ name: 'id', description: 'Account UUID', example: '123e4567-e89b-12d3-a456-426614174000' }),
    (0, swagger_1.ApiResponse)({ status: 200, description: 'Account found', schema: {
            type: 'object',
            properties: {
                id: { type: 'string', format: 'uuid' },
                currency: { type: 'string', example: 'USD' },
                status: { type: 'string', example: 'ACTIVE' },
                created_at: { type: 'string', format: 'date-time' }
            }
        } }),
    (0, swagger_1.ApiResponse)({ status: 400, description: 'Invalid account ID format' }),
    (0, swagger_1.ApiResponse)({ status: 404, description: 'Account not found' }),
    __param(0, (0, common_1.Param)('id')),
    __metadata("design:type", Function),
    __metadata("design:paramtypes", [String]),
    __metadata("design:returntype", Promise)
], AccountsController.prototype, "getAccount", null);
exports.AccountsController = AccountsController = __decorate([
    (0, swagger_1.ApiTags)('accounts'),
    (0, common_1.Controller)('accounts'),
    __metadata("design:paramtypes", [accounts_service_1.AccountsService])
], AccountsController);
//# sourceMappingURL=accounts.controller.js.map