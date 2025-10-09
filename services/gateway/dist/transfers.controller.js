"use strict";
var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
var __metadata = (this && this.__metadata) || function (k, v) {
    if (typeof Reflect === "object" && typeof Reflect.metadata === "function") return Reflect.metadata(k, v);
};
var __param = (this && this.__param) || function (paramIndex, decorator) {
    return function (target, key) { decorator(target, key, paramIndex); }
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.TransfersController = void 0;
const common_1 = require("@nestjs/common");
const zod_1 = require("zod");
const TransferSchema = zod_1.z.object({
    from: zod_1.z.string().uuid(),
    to: zod_1.z.string().uuid(),
    amount: zod_1.z.number().int().positive(),
    currency: zod_1.z.string(),
    idempotencyKey: zod_1.z.string().min(8)
});
let TransfersController = class TransfersController {
    async create(body) {
        const parsed = TransferSchema.safeParse(body);
        if (!parsed.success)
            return { error: parsed.error.flatten() };
        const dto = parsed.data;
        return { accepted: true, transferId: dto.idempotencyKey };
    }
};
exports.TransfersController = TransfersController;
__decorate([
    (0, common_1.Post)(),
    __param(0, (0, common_1.Body)()),
    __metadata("design:type", Function),
    __metadata("design:paramtypes", [Object]),
    __metadata("design:returntype", Promise)
], TransfersController.prototype, "create", null);
exports.TransfersController = TransfersController = __decorate([
    (0, common_1.Controller)('transfers')
], TransfersController);
//# sourceMappingURL=transfers.controller.js.map