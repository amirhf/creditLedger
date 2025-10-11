"use strict";
var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.AppModule = void 0;
const common_1 = require("@nestjs/common");
const health_controller_1 = require("./health.controller");
const transfers_controller_1 = require("./transfers.controller");
const accounts_controller_1 = require("./controllers/accounts.controller");
const balances_controller_1 = require("./controllers/balances.controller");
const accounts_service_1 = require("./services/accounts.service");
const orchestrator_service_1 = require("./services/orchestrator.service");
const readmodel_service_1 = require("./services/readmodel.service");
let AppModule = class AppModule {
};
exports.AppModule = AppModule;
exports.AppModule = AppModule = __decorate([
    (0, common_1.Module)({
        controllers: [
            health_controller_1.HealthController,
            accounts_controller_1.AccountsController,
            transfers_controller_1.TransfersController,
            balances_controller_1.BalancesController
        ],
        providers: [
            accounts_service_1.AccountsService,
            orchestrator_service_1.OrchestratorService,
            readmodel_service_1.ReadModelService
        ],
    })
], AppModule);
//# sourceMappingURL=app.module.js.map