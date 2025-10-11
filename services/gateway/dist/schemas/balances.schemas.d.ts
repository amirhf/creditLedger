import { z } from 'zod';
export declare const GetBalanceResponseSchema: z.ZodObject<{
    account_id: z.ZodString;
    balance_minor: z.ZodNumber;
    currency: z.ZodString;
    updated_at: z.ZodString;
}, "strip", z.ZodTypeAny, {
    currency?: string;
    account_id?: string;
    balance_minor?: number;
    updated_at?: string;
}, {
    currency?: string;
    account_id?: string;
    balance_minor?: number;
    updated_at?: string;
}>;
export type GetBalanceResponse = z.infer<typeof GetBalanceResponseSchema>;
export declare const GetStatementsQuerySchema: z.ZodObject<{
    from: z.ZodOptional<z.ZodString>;
    to: z.ZodOptional<z.ZodString>;
    limit: z.ZodDefault<z.ZodNumber>;
}, "strip", z.ZodTypeAny, {
    limit?: number;
    from?: string;
    to?: string;
}, {
    limit?: number;
    from?: string;
    to?: string;
}>;
export type GetStatementsQuery = z.infer<typeof GetStatementsQuerySchema>;
export declare const StatementEntrySchema: z.ZodObject<{
    id: z.ZodNumber;
    account_id: z.ZodString;
    entry_id: z.ZodString;
    amount_minor: z.ZodNumber;
    side: z.ZodString;
    ts: z.ZodString;
}, "strip", z.ZodTypeAny, {
    amount_minor?: number;
    id?: number;
    account_id?: string;
    entry_id?: string;
    side?: string;
    ts?: string;
}, {
    amount_minor?: number;
    id?: number;
    account_id?: string;
    entry_id?: string;
    side?: string;
    ts?: string;
}>;
export declare const GetStatementsResponseSchema: z.ZodObject<{
    statements: z.ZodArray<z.ZodObject<{
        id: z.ZodNumber;
        account_id: z.ZodString;
        entry_id: z.ZodString;
        amount_minor: z.ZodNumber;
        side: z.ZodString;
        ts: z.ZodString;
    }, "strip", z.ZodTypeAny, {
        amount_minor?: number;
        id?: number;
        account_id?: string;
        entry_id?: string;
        side?: string;
        ts?: string;
    }, {
        amount_minor?: number;
        id?: number;
        account_id?: string;
        entry_id?: string;
        side?: string;
        ts?: string;
    }>, "many">;
}, "strip", z.ZodTypeAny, {
    statements?: {
        amount_minor?: number;
        id?: number;
        account_id?: string;
        entry_id?: string;
        side?: string;
        ts?: string;
    }[];
}, {
    statements?: {
        amount_minor?: number;
        id?: number;
        account_id?: string;
        entry_id?: string;
        side?: string;
        ts?: string;
    }[];
}>;
export type StatementEntry = z.infer<typeof StatementEntrySchema>;
export type GetStatementsResponse = z.infer<typeof GetStatementsResponseSchema>;
