"use strict";
var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
var __metadata = (this && this.__metadata) || function (k, v) {
    if (typeof Reflect === "object" && typeof Reflect.metadata === "function") return Reflect.metadata(k, v);
};
var HttpClientService_1;
Object.defineProperty(exports, "__esModule", { value: true });
exports.HttpClientService = void 0;
const common_1 = require("@nestjs/common");
const axios_1 = require("axios");
let HttpClientService = HttpClientService_1 = class HttpClientService {
    constructor(config) {
        this.logger = new common_1.Logger(HttpClientService_1.name);
        this.retries = config.retries || 0;
        this.client = axios_1.default.create({
            baseURL: config.baseURL,
            timeout: config.timeout || 5000,
            headers: {
                'Content-Type': 'application/json'
            }
        });
        this.client.interceptors.request.use((config) => {
            var _a;
            this.logger.debug(`${(_a = config.method) === null || _a === void 0 ? void 0 : _a.toUpperCase()} ${config.baseURL}${config.url}`);
            return config;
        }, (error) => {
            this.logger.error('Request error:', error.message);
            return Promise.reject(error);
        });
        this.client.interceptors.response.use((response) => {
            this.logger.debug(`Response ${response.status} from ${response.config.url}`);
            return response;
        }, (error) => {
            var _a, _b;
            if (error.response) {
                this.logger.error(`Response error ${error.response.status} from ${(_a = error.config) === null || _a === void 0 ? void 0 : _a.url}: ${JSON.stringify(error.response.data)}`);
            }
            else if (error.request) {
                this.logger.error(`No response from ${(_b = error.config) === null || _b === void 0 ? void 0 : _b.url}`);
            }
            else {
                this.logger.error(`Request setup error: ${error.message}`);
            }
            return Promise.reject(error);
        });
    }
    async get(url, config) {
        const response = await this.client.get(url, config);
        return response.data;
    }
    async post(url, data, config) {
        const response = await this.client.post(url, data, config);
        return response.data;
    }
    async put(url, data, config) {
        const response = await this.client.put(url, data, config);
        return response.data;
    }
    async delete(url, config) {
        const response = await this.client.delete(url, config);
        return response.data;
    }
    setHeader(key, value) {
        this.client.defaults.headers.common[key] = value;
    }
};
exports.HttpClientService = HttpClientService;
exports.HttpClientService = HttpClientService = HttpClientService_1 = __decorate([
    (0, common_1.Injectable)(),
    __metadata("design:paramtypes", [Object])
], HttpClientService);
//# sourceMappingURL=http-client.service.js.map