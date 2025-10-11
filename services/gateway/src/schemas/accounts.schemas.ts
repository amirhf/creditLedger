import { z } from 'zod';

export const CreateAccountSchema = z.object({
  currency: z.string().length(3).regex(/^[A-Z]{3}$/, 'Currency must be 3 uppercase letters (e.g., USD, EUR)')
});

export type CreateAccountDto = z.infer<typeof CreateAccountSchema>;

export const AccountResponseSchema = z.object({
  account_id: z.string().uuid()
});

export type AccountResponse = z.infer<typeof AccountResponseSchema>;

export const GetAccountResponseSchema = z.object({
  id: z.string().uuid(),
  currency: z.string(),
  status: z.string(),
  created_at: z.string()
});

export type GetAccountResponse = z.infer<typeof GetAccountResponseSchema>;
