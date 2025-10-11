import { NestFactory } from '@nestjs/core';
import { AppModule } from './app.module';
import { AllExceptionsFilter } from './filters/http-exception.filter';
import { Logger } from '@nestjs/common';
import { SwaggerModule, DocumentBuilder } from '@nestjs/swagger';

async function bootstrap() {
  const logger = new Logger('Bootstrap');
  const app = await NestFactory.create(AppModule);
  
  // Global exception filter
  app.useGlobalFilters(new AllExceptionsFilter());
  
  // Enable CORS for development
  app.enableCors();
  
  // Swagger/OpenAPI setup
  const config = new DocumentBuilder()
    .setTitle('Credit Ledger Gateway API')
    .setDescription('Public REST API for the Credit Ledger microservices system')
    .setVersion('1.0')
    .addTag('accounts', 'Account management')
    .addTag('transfers', 'Transfer operations')
    .addTag('balances', 'Balance and statement queries')
    .build();
  const document = SwaggerModule.createDocument(app, config);
  SwaggerModule.setup('api', app, document);
  
  const port = process.env.PORT || 4000;
  await app.listen(port);
  
  logger.log(`Gateway service listening on port ${port}`);
  logger.log(`Swagger UI available at http://localhost:${port}/api`);
  logger.log(`Environment:`);
  logger.log(`  ACCOUNTS_SERVICE_URL: ${process.env.ACCOUNTS_SERVICE_URL || 'http://localhost:7101'}`);
  logger.log(`  ORCHESTRATOR_SERVICE_URL: ${process.env.ORCHESTRATOR_SERVICE_URL || 'http://localhost:7103'}`);
  logger.log(`  READMODEL_SERVICE_URL: ${process.env.READMODEL_SERVICE_URL || 'http://localhost:7104'}`);
}

bootstrap();