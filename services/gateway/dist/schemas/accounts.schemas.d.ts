import { z } from 'zod';
export declare const CreateAccountSchema: z.ZodObject<{
    currency: z.ZodString;
}, "strip", z.ZodTypeAny, {
    currency?: string;
}, {
    currency?: string;
}>;
export type CreateAccountDto = z.infer<typeof CreateAccountSchema>;
export declare const AccountResponseSchema: z.ZodObject<{
    account_id: z.ZodString;
}, "strip", z.ZodTypeAny, {
    account_id?: string;
}, {
    account_id?: string;
}>;
export type AccountResponse = z.infer<typeof AccountResponseSchema>;
export declare const GetAccountResponseSchema: z.ZodObject<{
    id: z.ZodString;
    currency: z.ZodString;
    status: z.ZodString;
    created_at: z.ZodString;
}, "strip", z.ZodTypeAny, {
    status?: string;
    currency?: string;
    id?: string;
    created_at?: string;
}, {
    status?: string;
    currency?: string;
    id?: string;
    created_at?: string;
}>;
export type GetAccountResponse = z.infer<typeof GetAccountResponseSchema>;
