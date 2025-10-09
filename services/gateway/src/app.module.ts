import { Module } from '@nestjs/common';
import { HealthController } from './health.controller';
import { TransfersController } from './transfers.controller';


@Module({
    controllers: [HealthController, TransfersController],
})
export class AppModule {}