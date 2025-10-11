import { Controller, Get, Param, Query, BadRequestException } from '@nestjs/common';
import { ApiTags, ApiOperation, ApiResponse, ApiParam, ApiQuery } from '@nestjs/swagger';
import { ReadModelService } from '../services/readmodel.service';
import { GetStatementsQuerySchema } from '../schemas/balances.schemas';

@ApiTags('balances')
@Controller('accounts')
export class BalancesController {
  constructor(private readonly readModelService: ReadModelService) {}

  @Get(':id/balance')
  @ApiOperation({ summary: 'Get account balance', description: 'Retrieves the current balance for an account' })
  @ApiParam({ name: 'id', description: 'Account UUID', example: '123e4567-e89b-12d3-a456-426614174000' })
  @ApiResponse({ status: 200, description: 'Balance retrieved successfully', schema: {
    type: 'object',
    properties: {
      account_id: { type: 'string', format: 'uuid' },
      balance_minor: { type: 'integer', example: 5000, description: 'Balance in minor units' },
      currency: { type: 'string', example: 'USD' },
      updated_at: { type: 'string', format: 'date-time' }
    }
  }})
  @ApiResponse({ status: 400, description: 'Invalid account ID format' })
  @ApiResponse({ status: 404, description: 'Account not found' })
  async getBalance(@Param('id') id: string) {
    // Basic UUID validation
    const uuidRegex = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;
    if (!uuidRegex.test(id)) {
      throw new BadRequestException({
        error: 'VALIDATION_ERROR',
        message: 'Invalid account ID format (must be UUID)'
      });
    }

    return this.readModelService.getBalance(id);
  }

  @Get(':id/statements')
  @ApiOperation({ summary: 'Get account statements', description: 'Retrieves transaction history for an account' })
  @ApiParam({ name: 'id', description: 'Account UUID', example: '123e4567-e89b-12d3-a456-426614174000' })
  @ApiQuery({ name: 'from', required: false, description: 'Start date (ISO8601)', example: '2024-01-01T00:00:00Z' })
  @ApiQuery({ name: 'to', required: false, description: 'End date (ISO8601)', example: '2024-12-31T23:59:59Z' })
  @ApiQuery({ name: 'limit', required: false, type: Number, description: 'Maximum number of entries (1-1000)', example: 100 })
  @ApiResponse({ status: 200, description: 'Statements retrieved successfully', schema: {
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
  }})
  @ApiResponse({ status: 400, description: 'Invalid parameters' })
  @ApiResponse({ status: 404, description: 'Account not found' })
  async getStatements(@Param('id') id: string, @Query() query: any) {
    // Basic UUID validation
    const uuidRegex = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;
    if (!uuidRegex.test(id)) {
      throw new BadRequestException({
        error: 'VALIDATION_ERROR',
        message: 'Invalid account ID format (must be UUID)'
      });
    }

    // Validate query parameters
    const parseResult = GetStatementsQuerySchema.safeParse(query);
    if (!parseResult.success) {
      throw new BadRequestException({
        error: 'VALIDATION_ERROR',
        message: 'Invalid query parameters',
        details: parseResult.error.flatten()
      });
    }

    return this.readModelService.getStatements(id, parseResult.data);
  }
}
