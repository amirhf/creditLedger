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
exports.TransfersController = void 0;
const common_1 = require("@nestjs/common");
const swagger_1 = require("@nestjs/swagger");
const orchestrator_service_1 = require("./services/orchestrator.service");
const transfers_schemas_1 = require("./schemas/transfers.schemas");
let TransfersController = class TransfersController {
    constructor(orchestratorService) {
        this.orchestratorService = orchestratorService;
    }
    async createTransfer(body) {
        const parseResult = transfers_schemas_1.CreateTransferSchema.safeParse(body);
        if (!parseResult.success) {
            throw new common_1.BadRequestException({
                error: 'VALIDATION_ERROR',
                message: 'Invalid request body',
                details: parseResult.error.flatten()
            });
        }
        const dto = parseResult.data;
        return this.orchestratorService.createTransfer(dto);
    }
    async listTransfers(query) {
        return this.orchestratorService.listTransfers(query);
    }
    async getTransfer(id) {
        const uuidRegex = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;
        if (!uuidRegex.test(id)) {
            throw new common_1.BadRequestException({
                error: 'VALIDATION_ERROR',
                message: 'Invalid transfer ID format (must be UUID)'
            });
        }
        return this.orchestratorService.getTransfer(id);
    }
};
exports.TransfersController = TransfersController;
__decorate([
    (0, common_1.Post)(),
    (0, common_1.HttpCode)(common_1.HttpStatus.ACCEPTED),
    (0, swagger_1.ApiOperation)({ summary: 'Create a transfer', description: 'Initiates a transfer between two accounts with idempotency support' }),
    (0, swagger_1.ApiBody)({
        description: 'Transfer creation request',
        schema: {
            type: 'object',
            required: ['from_account_id', 'to_account_id', 'amount_minor', 'currency', 'idempotency_key'],
            properties: {
                from_account_id: {
                    type: 'string',
                    format: 'uuid',
                    description: 'Source account UUID',
                    example: '123e4567-e89b-12d3-a456-426614174000'
                },
                to_account_id: {
                    type: 'string',
                    format: 'uuid',
                    description: 'Destination account UUID (must differ from source)',
                    example: '987fcdeb-51a2-43d7-b890-123456789abc'
                },
                amount_minor: {
                    type: 'integer',
                    description: 'Amount in minor units (e.g., cents for USD)',
                    example: 5000,
                    minimum: 1
                },
                currency: {
                    type: 'string',
                    description: 'Three-letter ISO currency code',
                    example: 'USD',
                    pattern: '^[A-Z]{3}$',
                    minLength: 3,
                    maxLength: 3
                },
                idempotency_key: {
                    type: 'string',
                    description: 'Unique key to ensure idempotent processing (8-128 characters)',
                    example: 'transfer-2024-001',
                    minLength: 8,
                    maxLength: 128
                }
            }
        }
    }),
    (0, swagger_1.ApiResponse)({ status: 202, description: 'Transfer accepted for processing', schema: {
            type: 'object',
            properties: {
                transfer_id: { type: 'string', format: 'uuid' },
                status: { type: 'string', example: 'INITIATED' }
            }
        } }),
    (0, swagger_1.ApiResponse)({ status: 400, description: 'Invalid request body or validation error' }),
    __param(0, (0, common_1.Body)()),
    __metadata("design:type", Function),
    __metadata("design:paramtypes", [Object]),
    __metadata("design:returntype", Promise)
], TransfersController.prototype, "createTransfer", null);
__decorate([
    (0, common_1.Get)(),
    (0, swagger_1.ApiOperation)({ summary: 'List transfers', description: 'Retrieves a list of transfers with optional filtering' }),
    (0, swagger_1.ApiQuery)({ name: 'from_account_id', required: false, description: 'Filter by source account', example: '123e4567-e89b-12d3-a456-426614174000' }),
    (0, swagger_1.ApiQuery)({ name: 'to_account_id', required: false, description: 'Filter by destination account', example: '987fcdeb-51a2-43d7-b890-123456789abc' }),
    (0, swagger_1.ApiQuery)({ name: 'status', required: false, description: 'Filter by status', example: 'COMPLETED', enum: ['INITIATED', 'COMPLETED', 'FAILED'] }),
    (0, swagger_1.ApiQuery)({ name: 'currency', required: false, description: 'Filter by currency', example: 'USD' }),
    (0, swagger_1.ApiQuery)({ name: 'limit', required: false, type: Number, description: 'Maximum number of results (1-100)', example: 20 }),
    (0, swagger_1.ApiQuery)({ name: 'offset', required: false, type: Number, description: 'Number of results to skip', example: 0 }),
    (0, swagger_1.ApiResponse)({ status: 200, description: 'Transfers retrieved successfully', schema: {
            type: 'object',
            properties: {
                transfers: {
                    type: 'array',
                    items: {
                        type: 'object',
                        properties: {
                            id: { type: 'string', format: 'uuid' },
                            from_account_id: { type: 'string', format: 'uuid' },
                            to_account_id: { type: 'string', format: 'uuid' },
                            amount_minor: { type: 'integer' },
                            currency: { type: 'string' },
                            status: { type: 'string' },
                            created_at: { type: 'string', format: 'date-time' }
                        }
                    }
                },
                total: { type: 'integer', description: 'Total number of transfers matching filters' },
                limit: { type: 'integer' },
                offset: { type: 'integer' }
            }
        } }),
    __param(0, (0, common_1.Query)()),
    __metadata("design:type", Function),
    __metadata("design:paramtypes", [Object]),
    __metadata("design:returntype", Promise)
], TransfersController.prototype, "listTransfers", null);
__decorate([
    (0, common_1.Get)(':id'),
    (0, swagger_1.ApiOperation)({ summary: 'Get transfer details', description: 'Retrieves transfer information by ID' }),
    (0, swagger_1.ApiParam)({ name: 'id', description: 'Transfer UUID', example: '123e4567-e89b-12d3-a456-426614174000' }),
    (0, swagger_1.ApiResponse)({ status: 200, description: 'Transfer found', schema: {
            type: 'object',
            properties: {
                id: { type: 'string', format: 'uuid' },
                from_account_id: { type: 'string', format: 'uuid' },
                to_account_id: { type: 'string', format: 'uuid' },
                amount_minor: { type: 'integer' },
                currency: { type: 'string' },
                status: { type: 'string', example: 'COMPLETED' },
                idempotency_key: { type: 'string' },
                created_at: { type: 'string', format: 'date-time' }
            }
        } }),
    (0, swagger_1.ApiResponse)({ status: 400, description: 'Invalid transfer ID format' }),
    (0, swagger_1.ApiResponse)({ status: 404, description: 'Transfer not found' }),
    __param(0, (0, common_1.Param)('id')),
    __metadata("design:type", Function),
    __metadata("design:paramtypes", [String]),
    __metadata("design:returntype", Promise)
], TransfersController.prototype, "getTransfer", null);
exports.TransfersController = TransfersController = __decorate([
    (0, swagger_1.ApiTags)('transfers'),
    (0, common_1.Controller)('transfers'),
    __metadata("design:paramtypes", [orchestrator_service_1.OrchestratorService])
], TransfersController);
//# sourceMappingURL=transfers.controller.js.map