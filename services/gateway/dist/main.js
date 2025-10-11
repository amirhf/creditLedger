"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const tracing_1 = require("./tracing");
(0, tracing_1.initTracing)();
const core_1 = require("@nestjs/core");
const app_module_1 = require("./app.module");
const http_exception_filter_1 = require("./filters/http-exception.filter");
const common_1 = require("@nestjs/common");
const swagger_1 = require("@nestjs/swagger");
async function bootstrap() {
    const logger = new common_1.Logger('Bootstrap');
    const app = await core_1.NestFactory.create(app_module_1.AppModule);
    app.useGlobalFilters(new http_exception_filter_1.AllExceptionsFilter());
    app.enableCors();
    const config = new swagger_1.DocumentBuilder()
        .setTitle('Credit Ledger Gateway API')
        .setDescription('Public REST API for the Credit Ledger microservices system')
        .setVersion('1.0')
        .addTag('accounts', 'Account management')
        .addTag('transfers', 'Transfer operations')
        .addTag('balances', 'Balance and statement queries')
        .build();
    const document = swagger_1.SwaggerModule.createDocument(app, config);
    swagger_1.SwaggerModule.setup('api', app, document);
    const port = process.env.PORT || 4000;
    await app.listen(port);
    logger.log(`Gateway service listening on port ${port}`);
    logger.log(`Swagger UI available at http://localhost:${port}/api`);
    logger.log(`Environment:`);
    logger.log(`  ACCOUNTS_SERVICE_URL: ${process.env.ACCOUNTS_SERVICE_URL || 'http://localhost:7101'}`);
    logger.log(`  ORCHESTRATOR_SERVICE_URL: ${process.env.ORCHESTRATOR_SERVICE_URL || 'http://localhost:7103'}`);
    logger.log(`  READMODEL_SERVICE_URL: ${process.env.READMODEL_SERVICE_URL || 'http://localhost:7104'}`);
    logger.log(`  OTEL_EXPORTER_OTLP_ENDPOINT: ${process.env.OTEL_EXPORTER_OTLP_ENDPOINT || 'http://localhost:4318'}`);
}
bootstrap();
//# sourceMappingURL=main.js.map