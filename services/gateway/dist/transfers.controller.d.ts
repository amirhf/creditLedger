import { z } from 'zod';
export declare class TransfersController {
    create(body: any): Promise<{
        error: z.typeToFlattenedError<{
            currency?: string;
            from?: string;
            to?: string;
            amount?: number;
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
