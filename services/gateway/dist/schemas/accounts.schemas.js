"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.GetAccountResponseSchema = exports.AccountResponseSchema = exports.CreateAccountSchema = void 0;
const zod_1 = require("zod");
exports.CreateAccountSchema = zod_1.z.object({
    currency: zod_1.z.string().length(3).regex(/^[A-Z]{3}$/, 'Currency must be 3 uppercase letters (e.g., USD, EUR)')
});
exports.AccountResponseSchema = zod_1.z.object({
    account_id: zod_1.z.string().uuid()
});
exports.GetAccountResponseSchema = zod_1.z.object({
    id: zod_1.z.string().uuid(),
    currency: zod_1.z.string(),
    status: zod_1.z.string(),
    created_at: zod_1.z.string()
});
//# sourceMappingURL=accounts.schemas.js.map