import { Body, Controller, Post } from '@nestjs/common';
import { z } from 'zod';


const TransferSchema = z.object({
    from: z.string().uuid(),
    to: z.string().uuid(),
    amount: z.number().int().positive(),
    currency: z.string(),
    idempotencyKey: z.string().min(8)
});


type TransferDto = z.infer<typeof TransferSchema>;


@Controller('transfers')
export class TransfersController {
    @Post()
    async create(@Body() body: any) {
        const parsed = TransferSchema.safeParse(body);
        if (!parsed.success) return { error: parsed.error.flatten() };
        const dto: TransferDto = parsed.data;
// TODO: call orchestrator service
        return { accepted: true, transferId: dto.idempotencyKey };
    }
}