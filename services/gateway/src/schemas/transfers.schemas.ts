import { z } from 'zod';

export const CreateTransferSchema = z.object({
  from_account_id: z.string().uuid('Invalid from_account_id UUID'),
  to_account_id: z.string().uuid('Invalid to_account_id UUID'),
  amount_minor: z.number().int().positive('Amount must be a positive integer'),
  currency: z.string().length(3).regex(/^[A-Z]{3}$/, 'Currency must be 3 uppercase letters'),
  idempotency_key: z.string().min(8, 'Idempotency key must be at least 8 characters').max(128, 'Idempotency key must be at most 128 characters')
}).refine(data => data.from_account_id !== data.to_account_id, {
  message: 'from_account_id and to_account_id must be different',
  path: ['to_account_id']
});

export type CreateTransferDto = z.infer<typeof CreateTransferSchema>;

export const TransferResponseSchema = z.object({
  transfer_id: z.string().uuid(),
  status: z.string()
});

export type TransferResponse = z.infer<typeof TransferResponseSchema>;

export const GetTransferResponseSchema = z.object({
  id: z.string().uuid(),
  from_account_id: z.string().uuid(),
  to_account_id: z.string().uuid(),
  amount_minor: z.number(),
  currency: z.string(),
  status: z.string(),
  idempotency_key: z.string(),
  created_at: z.string()
});

export type GetTransferResponse = z.infer<typeof GetTransferResponseSchema>;
