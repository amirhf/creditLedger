import { Controller, Post, Get, Body, Param, Query, HttpCode, HttpStatus, BadRequestException } from '@nestjs/common';
import { ApiTags, ApiOperation, ApiResponse, ApiBody, ApiParam, ApiQuery } from '@nestjs/swagger';
import { OrchestratorService } from './services/orchestrator.service';
import { CreateTransferSchema, CreateTransferDto } from './schemas/transfers.schemas';

@ApiTags('transfers')
@Controller('transfers')
export class TransfersController {
  constructor(private readonly orchestratorService: OrchestratorService) {}

  @Post()
  @HttpCode(HttpStatus.ACCEPTED)
  @ApiOperation({ summary: 'Create a transfer', description: 'Initiates a transfer between two accounts with idempotency support' })
  @ApiBody({
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
  })
  @ApiResponse({ status: 202, description: 'Transfer accepted for processing', schema: {
    type: 'object',
    properties: {
      transfer_id: { type: 'string', format: 'uuid' },
      status: { type: 'string', example: 'INITIATED' }
    }
  }})
  @ApiResponse({ status: 400, description: 'Invalid request body or validation error' })
  async createTransfer(@Body() body: any) {
    // Validate request body
    const parseResult = CreateTransferSchema.safeParse(body);
    if (!parseResult.success) {
      throw new BadRequestException({
        error: 'VALIDATION_ERROR',
        message: 'Invalid request body',
        details: parseResult.error.flatten()
      });
    }

    const dto: CreateTransferDto = parseResult.data;
    return this.orchestratorService.createTransfer(dto);
  }

  @Get()
  @ApiOperation({ summary: 'List transfers', description: 'Retrieves a list of transfers with optional filtering' })
  @ApiQuery({ name: 'from_account_id', required: false, description: 'Filter by source account', example: '123e4567-e89b-12d3-a456-426614174000' })
  @ApiQuery({ name: 'to_account_id', required: false, description: 'Filter by destination account', example: '987fcdeb-51a2-43d7-b890-123456789abc' })
  @ApiQuery({ name: 'status', required: false, description: 'Filter by status', example: 'COMPLETED', enum: ['INITIATED', 'COMPLETED', 'FAILED'] })
  @ApiQuery({ name: 'currency', required: false, description: 'Filter by currency', example: 'USD' })
  @ApiQuery({ name: 'limit', required: false, type: Number, description: 'Maximum number of results (1-100)', example: 20 })
  @ApiQuery({ name: 'offset', required: false, type: Number, description: 'Number of results to skip', example: 0 })
  @ApiResponse({ status: 200, description: 'Transfers retrieved successfully', schema: {
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
  }})
  async listTransfers(@Query() query: any) {
    return this.orchestratorService.listTransfers(query);
  }

  @Get(':id')
  @ApiOperation({ summary: 'Get transfer details', description: 'Retrieves transfer information by ID' })
  @ApiParam({ name: 'id', description: 'Transfer UUID', example: '123e4567-e89b-12d3-a456-426614174000' })
  @ApiResponse({ status: 200, description: 'Transfer found', schema: {
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
  }})
  @ApiResponse({ status: 400, description: 'Invalid transfer ID format' })
  @ApiResponse({ status: 404, description: 'Transfer not found' })
  async getTransfer(@Param('id') id: string) {
    // Basic UUID validation
    const uuidRegex = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;
    if (!uuidRegex.test(id)) {
      throw new BadRequestException({
        error: 'VALIDATION_ERROR',
        message: 'Invalid transfer ID format (must be UUID)'
      });
    }

    return this.orchestratorService.getTransfer(id);
  }
}