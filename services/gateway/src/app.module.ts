import { Module } from '@nestjs/common';
import { HealthController } from './health.controller';
import { TransfersController } from './transfers.controller';
import { AccountsController } from './controllers/accounts.controller';
import { BalancesController } from './controllers/balances.controller';
import { AccountsService } from './services/accounts.service';
import { OrchestratorService } from './services/orchestrator.service';
import { ReadModelService } from './services/readmodel.service';

@Module({
  controllers: [
    HealthController,
    AccountsController,
    TransfersController,
    BalancesController
  ],
  providers: [
    AccountsService,
    OrchestratorService,
    ReadModelService
  ],
})
export class AppModule {}