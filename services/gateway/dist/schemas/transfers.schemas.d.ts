import { z } from 'zod';
export declare const CreateTransferSchema: z.ZodEffects<z.ZodObject<{
    from_account_id: z.ZodString;
    to_account_id: z.ZodString;
    amount_minor: z.ZodNumber;
    currency: z.ZodString;
    idempotency_key: z.ZodString;
}, "strip", z.ZodTypeAny, {
    from_account_id?: string;
    to_account_id?: string;
    amount_minor?: number;
    currency?: string;
    idempotency_key?: string;
}, {
    from_account_id?: string;
    to_account_id?: string;
    amount_minor?: number;
    currency?: string;
    idempotency_key?: string;
}>, {
    from_account_id?: string;
    to_account_id?: string;
    amount_minor?: number;
    currency?: string;
    idempotency_key?: string;
}, {
    from_account_id?: string;
    to_account_id?: string;
    amount_minor?: number;
    currency?: string;
    idempotency_key?: string;
}>;
export type CreateTransferDto = z.infer<typeof CreateTransferSchema>;
export declare const TransferResponseSchema: z.ZodObject<{
    transfer_id: z.ZodString;
    status: z.ZodString;
}, "strip", z.ZodTypeAny, {
    status?: string;
    transfer_id?: string;
}, {
    status?: string;
    transfer_id?: string;
}>;
export type TransferResponse = z.infer<typeof TransferResponseSchema>;
export declare const GetTransferResponseSchema: z.ZodObject<{
    id: z.ZodString;
    from_account_id: z.ZodString;
    to_account_id: z.ZodString;
    amount_minor: z.ZodNumber;
    currency: z.ZodString;
    status: z.ZodString;
    idempotency_key: z.ZodString;
    created_at: z.ZodString;
}, "strip", z.ZodTypeAny, {
    status?: string;
    from_account_id?: string;
    to_account_id?: string;
    amount_minor?: number;
    currency?: string;
    idempotency_key?: string;
    id?: string;
    created_at?: string;
}, {
    status?: string;
    from_account_id?: string;
    to_account_id?: string;
    amount_minor?: number;
    currency?: string;
    idempotency_key?: string;
    id?: string;
    created_at?: string;
}>;
export type GetTransferResponse = z.infer<typeof GetTransferResponseSchema>;
