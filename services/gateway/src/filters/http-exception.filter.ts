import { ExceptionFilter, Catch, ArgumentsHost, HttpException, HttpStatus, Logger } from '@nestjs/common';
import { Request, Response } from 'express';
import { AxiosError } from 'axios';
import { ZodError } from 'zod';

@Catch()
export class AllExceptionsFilter implements ExceptionFilter {
  private readonly logger = new Logger(AllExceptionsFilter.name);

  catch(exception: unknown, host: ArgumentsHost) {
    const ctx = host.switchToHttp();
    const response = ctx.getResponse<Response>();
    const request = ctx.getRequest<Request>();

    let status = HttpStatus.INTERNAL_SERVER_ERROR;
    let errorResponse: any = {
      error: {
        code: 'INTERNAL_ERROR',
        message: 'An unexpected error occurred',
        timestamp: new Date().toISOString(),
        path: request.url
      }
    };

    // Handle NestJS HttpException
    if (exception instanceof HttpException) {
      status = exception.getStatus();
      const exceptionResponse = exception.getResponse();
      
      errorResponse = {
        error: {
          code: typeof exceptionResponse === 'object' && 'error' in exceptionResponse 
            ? (exceptionResponse as any).error 
            : exception.name,
          message: exception.message,
          details: typeof exceptionResponse === 'object' ? exceptionResponse : undefined,
          timestamp: new Date().toISOString(),
          path: request.url
        }
      };
    }
    // Handle Zod validation errors
    else if (exception instanceof ZodError) {
      status = HttpStatus.BAD_REQUEST;
      errorResponse = {
        error: {
          code: 'VALIDATION_ERROR',
          message: 'Request validation failed',
          details: exception.errors.map(err => ({
            field: err.path.join('.'),
            message: err.message
          })),
          timestamp: new Date().toISOString(),
          path: request.url
        }
      };
    }
    // Handle Axios errors (from downstream services)
    else if (this.isAxiosError(exception)) {
      const axiosError = exception as AxiosError;
      
      if (axiosError.response) {
        // Downstream service returned an error
        status = axiosError.response.status;
        errorResponse = {
          error: {
            code: 'DOWNSTREAM_ERROR',
            message: axiosError.message,
            details: axiosError.response.data,
            service: axiosError.config?.baseURL,
            timestamp: new Date().toISOString(),
            path: request.url
          }
        };
      } else if (axiosError.request) {
        // No response from downstream service
        status = HttpStatus.SERVICE_UNAVAILABLE;
        errorResponse = {
          error: {
            code: 'SERVICE_UNAVAILABLE',
            message: 'Downstream service is unavailable',
            service: axiosError.config?.baseURL,
            timestamp: new Date().toISOString(),
            path: request.url
          }
        };
      }
    }
    // Handle generic errors
    else if (exception instanceof Error) {
      errorResponse.error.message = exception.message;
      errorResponse.error.code = exception.name;
    }

    // Log the error
    this.logger.error(
      `${request.method} ${request.url} - Status: ${status}`,
      exception instanceof Error ? exception.stack : String(exception)
    );

    response.status(status).json(errorResponse);
  }

  private isAxiosError(error: any): error is AxiosError {
    return error.isAxiosError === true;
  }
}
