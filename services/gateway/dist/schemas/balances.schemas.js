"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.GetStatementsResponseSchema = exports.StatementEntrySchema = exports.GetStatementsQuerySchema = exports.GetBalanceResponseSchema = void 0;
const zod_1 = require("zod");
exports.GetBalanceResponseSchema = zod_1.z.object({
    account_id: zod_1.z.string().uuid(),
    balance_minor: zod_1.z.number(),
    currency: zod_1.z.string(),
    updated_at: zod_1.z.string()
});
exports.GetStatementsQuerySchema = zod_1.z.object({
    from: zod_1.z.string().datetime().optional(),
    to: zod_1.z.string().datetime().optional(),
    limit: zod_1.z.coerce.number().int().min(1).max(1000).default(100)
});
exports.StatementEntrySchema = zod_1.z.object({
    id: zod_1.z.number(),
    account_id: zod_1.z.string().uuid(),
    entry_id: zod_1.z.string().uuid(),
    amount_minor: zod_1.z.number(),
    side: zod_1.z.string(),
    ts: zod_1.z.string()
});
exports.GetStatementsResponseSchema = zod_1.z.object({
    statements: zod_1.z.array(exports.StatementEntrySchema)
});
//# sourceMappingURL=balances.schemas.js.map