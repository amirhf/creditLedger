"use strict";
var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
var AllExceptionsFilter_1;
Object.defineProperty(exports, "__esModule", { value: true });
exports.AllExceptionsFilter = void 0;
const common_1 = require("@nestjs/common");
const zod_1 = require("zod");
let AllExceptionsFilter = AllExceptionsFilter_1 = class AllExceptionsFilter {
    constructor() {
        this.logger = new common_1.Logger(AllExceptionsFilter_1.name);
    }
    catch(exception, host) {
        var _a, _b;
        const ctx = host.switchToHttp();
        const response = ctx.getResponse();
        const request = ctx.getRequest();
        let status = common_1.HttpStatus.INTERNAL_SERVER_ERROR;
        let errorResponse = {
            error: {
                code: 'INTERNAL_ERROR',
                message: 'An unexpected error occurred',
                timestamp: new Date().toISOString(),
                path: request.url
            }
        };
        if (exception instanceof common_1.HttpException) {
            status = exception.getStatus();
            const exceptionResponse = exception.getResponse();
            errorResponse = {
                error: {
                    code: typeof exceptionResponse === 'object' && 'error' in exceptionResponse
                        ? exceptionResponse.error
                        : exception.name,
                    message: exception.message,
                    details: typeof exceptionResponse === 'object' ? exceptionResponse : undefined,
                    timestamp: new Date().toISOString(),
                    path: request.url
                }
            };
        }
        else if (exception instanceof zod_1.ZodError) {
            status = common_1.HttpStatus.BAD_REQUEST;
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
        else if (this.isAxiosError(exception)) {
            const axiosError = exception;
            if (axiosError.response) {
                status = axiosError.response.status;
                errorResponse = {
                    error: {
                        code: 'DOWNSTREAM_ERROR',
                        message: axiosError.message,
                        details: axiosError.response.data,
                        service: (_a = axiosError.config) === null || _a === void 0 ? void 0 : _a.baseURL,
                        timestamp: new Date().toISOString(),
                        path: request.url
                    }
                };
            }
            else if (axiosError.request) {
                status = common_1.HttpStatus.SERVICE_UNAVAILABLE;
                errorResponse = {
                    error: {
                        code: 'SERVICE_UNAVAILABLE',
                        message: 'Downstream service is unavailable',
                        service: (_b = axiosError.config) === null || _b === void 0 ? void 0 : _b.baseURL,
                        timestamp: new Date().toISOString(),
                        path: request.url
                    }
                };
            }
        }
        else if (exception instanceof Error) {
            errorResponse.error.message = exception.message;
            errorResponse.error.code = exception.name;
        }
        this.logger.error(`${request.method} ${request.url} - Status: ${status}`, exception instanceof Error ? exception.stack : String(exception));
        response.status(status).json(errorResponse);
    }
    isAxiosError(error) {
        return error.isAxiosError === true;
    }
};
exports.AllExceptionsFilter = AllExceptionsFilter;
exports.AllExceptionsFilter = AllExceptionsFilter = AllExceptionsFilter_1 = __decorate([
    (0, common_1.Catch)()
], AllExceptionsFilter);
//# sourceMappingURL=http-exception.filter.js.map