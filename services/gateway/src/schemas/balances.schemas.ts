import { z } from 'zod';

export const GetBalanceResponseSchema = z.object({
  account_id: z.string().uuid(),
  balance_minor: z.number(),
  currency: z.string(),
  updated_at: z.string()
});

export type GetBalanceResponse = z.infer<typeof GetBalanceResponseSchema>;

export const GetStatementsQuerySchema = z.object({
  from: z.string().datetime().optional(),
  to: z.string().datetime().optional(),
  limit: z.coerce.number().int().min(1).max(1000).default(100)
});

export type GetStatementsQuery = z.infer<typeof GetStatementsQuerySchema>;

export const StatementEntrySchema = z.object({
  id: z.number(),
  account_id: z.string().uuid(),
  entry_id: z.string().uuid(),
  amount_minor: z.number(),
  side: z.string(),
  ts: z.string()
});

export const GetStatementsResponseSchema = z.object({
  statements: z.array(StatementEntrySchema)
});

export type StatementEntry = z.infer<typeof StatementEntrySchema>;
export type GetStatementsResponse = z.infer<typeof GetStatementsResponseSchema>;
