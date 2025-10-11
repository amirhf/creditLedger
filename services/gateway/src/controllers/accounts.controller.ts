import { Controller, Post, Get, Body, Param, Query, HttpCode, HttpStatus, BadRequestException } from '@nestjs/common';
import { ApiTags, ApiOperation, ApiResponse, ApiBody, ApiParam, ApiQuery } from '@nestjs/swagger';
import { AccountsService } from '../services/accounts.service';
import { CreateAccountSchema, CreateAccountDto } from '../schemas/accounts.schemas';

@ApiTags('accounts')
@Controller('accounts')
export class AccountsController {
  constructor(private readonly accountsService: AccountsService) {}

  @Post()
  @HttpCode(HttpStatus.CREATED)
  @ApiOperation({ summary: 'Create a new account', description: 'Creates a new account with the specified currency' })
  @ApiBody({
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
  })
  @ApiResponse({ status: 201, description: 'Account created successfully', schema: {
    type: 'object',
    properties: {
      account_id: { type: 'string', format: 'uuid', example: '123e4567-e89b-12d3-a456-426614174000' },
      currency: { type: 'string', example: 'USD' },
      status: { type: 'string', example: 'ACTIVE' }
    }
  }})
  @ApiResponse({ status: 400, description: 'Invalid request body' })
  async createAccount(@Body() body: any) {
    // Validate request body
    const parseResult = CreateAccountSchema.safeParse(body);
    if (!parseResult.success) {
      throw new BadRequestException({
        error: 'VALIDATION_ERROR',
        message: 'Invalid request body',
        details: parseResult.error.flatten()
      });
    }

    const dto: CreateAccountDto = parseResult.data;
    return this.accountsService.createAccount(dto);
  }

  @Get()
  @ApiOperation({ summary: 'List accounts', description: 'Retrieves a list of accounts with optional filtering' })
  @ApiQuery({ name: 'currency', required: false, description: 'Filter by currency code', example: 'USD' })
  @ApiQuery({ name: 'status', required: false, description: 'Filter by status', example: 'ACTIVE' })
  @ApiQuery({ name: 'limit', required: false, type: Number, description: 'Maximum number of results (1-100)', example: 20 })
  @ApiQuery({ name: 'offset', required: false, type: Number, description: 'Number of results to skip', example: 0 })
  @ApiResponse({ status: 200, description: 'Accounts retrieved successfully', schema: {
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
  }})
  async listAccounts(@Query() query: any) {
    return this.accountsService.listAccounts(query);
  }

  @Get(':id')
  @ApiOperation({ summary: 'Get account details', description: 'Retrieves account information by ID' })
  @ApiParam({ name: 'id', description: 'Account UUID', example: '123e4567-e89b-12d3-a456-426614174000' })
  @ApiResponse({ status: 200, description: 'Account found', schema: {
    type: 'object',
    properties: {
      id: { type: 'string', format: 'uuid' },
      currency: { type: 'string', example: 'USD' },
      status: { type: 'string', example: 'ACTIVE' },
      created_at: { type: 'string', format: 'date-time' }
    }
  }})
  @ApiResponse({ status: 400, description: 'Invalid account ID format' })
  @ApiResponse({ status: 404, description: 'Account not found' })
  async getAccount(@Param('id') id: string) {
    // Basic UUID validation
    const uuidRegex = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;
    if (!uuidRegex.test(id)) {
      throw new BadRequestException({
        error: 'VALIDATION_ERROR',
        message: 'Invalid account ID format (must be UUID)'
      });
    }

    return this.accountsService.getAccount(id);
  }
}
