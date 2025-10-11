"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.GetTransferResponseSchema = exports.TransferResponseSchema = exports.CreateTransferSchema = void 0;
const zod_1 = require("zod");
exports.CreateTransferSchema = zod_1.z.object({
    from_account_id: zod_1.z.string().uuid('Invalid from_account_id UUID'),
    to_account_id: zod_1.z.string().uuid('Invalid to_account_id UUID'),
    amount_minor: zod_1.z.number().int().positive('Amount must be a positive integer'),
    currency: zod_1.z.string().length(3).regex(/^[A-Z]{3}$/, 'Currency must be 3 uppercase letters'),
    idempotency_key: zod_1.z.string().min(8, 'Idempotency key must be at least 8 characters').max(128, 'Idempotency key must be at most 128 characters')
}).refine(data => data.from_account_id !== data.to_account_id, {
    message: 'from_account_id and to_account_id must be different',
    path: ['to_account_id']
});
exports.TransferResponseSchema = zod_1.z.object({
    transfer_id: zod_1.z.string().uuid(),
    status: zod_1.z.string()
});
exports.GetTransferResponseSchema = zod_1.z.object({
    id: zod_1.z.string().uuid(),
    from_account_id: zod_1.z.string().uuid(),
    to_account_id: zod_1.z.string().uuid(),
    amount_minor: zod_1.z.number(),
    currency: zod_1.z.string(),
    status: zod_1.z.string(),
    idempotency_key: zod_1.z.string(),
    created_at: zod_1.z.string()
});
//# sourceMappingURL=transfers.schemas.js.map