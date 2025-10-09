import { z } from 'zod';
export declare class TransfersController {
    create(body: any): Promise<{
        error: z.typeToFlattenedError<{
            from?: string;
            to?: string;
            amount?: number;
            currency?: string;
            idempotencyKey?: string;
        }, string>;
        accepted?: undefined;
        transferId?: undefined;
    } | {
        accepted: boolean;
        transferId: string;
        error?: undefined;
    }>;
}
