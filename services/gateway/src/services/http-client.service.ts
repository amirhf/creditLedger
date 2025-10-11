import { Injectable, Logger } from '@nestjs/common';
import axios, { AxiosInstance, AxiosError, AxiosRequestConfig } from 'axios';
import { context, trace, propagation } from '@opentelemetry/api';

export interface HttpClientConfig {
  baseURL: string;
  timeout?: number;
  retries?: number;
}

@Injectable()
export class HttpClientService {
  private readonly logger = new Logger(HttpClientService.name);
  private client: AxiosInstance;
  private retries: number;

  constructor(config: HttpClientConfig) {
    this.retries = config.retries || 0;
    this.client = axios.create({
      baseURL: config.baseURL,
      timeout: config.timeout || 5000,
      headers: {
        'Content-Type': 'application/json'
      }
    });

    // Request interceptor for logging and trace propagation
    this.client.interceptors.request.use(
      (config) => {
        // Propagate trace context to downstream services
        const activeContext = context.active();
        const span = trace.getActiveSpan();
        
        if (span) {
          // Inject trace context into headers (adds traceparent header)
          propagation.inject(activeContext, config.headers);
          
          const traceId = span.spanContext().traceId;
          const spanId = span.spanContext().spanId;
          
          this.logger.debug(`${config.method?.toUpperCase()} ${config.baseURL}${config.url}`, {
            traceId,
            spanId,
          });
        } else {
          this.logger.debug(`${config.method?.toUpperCase()} ${config.baseURL}${config.url}`);
        }
        
        return config;
      },
      (error) => {
        this.logger.error('Request error:', error.message);
        return Promise.reject(error);
      }
    );

    // Response interceptor for logging
    this.client.interceptors.response.use(
      (response) => {
        this.logger.debug(`Response ${response.status} from ${response.config.url}`);
        return response;
      },
      (error: AxiosError) => {
        if (error.response) {
          this.logger.error(
            `Response error ${error.response.status} from ${error.config?.url}: ${JSON.stringify(error.response.data)}`
          );
        } else if (error.request) {
          this.logger.error(`No response from ${error.config?.url}`);
        } else {
          this.logger.error(`Request setup error: ${error.message}`);
        }
        return Promise.reject(error);
      }
    );
  }

  async get<T>(url: string, config?: AxiosRequestConfig): Promise<T> {
    const response = await this.client.get<T>(url, config);
    return response.data;
  }

  async post<T>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T> {
    const response = await this.client.post<T>(url, data, config);
    return response.data;
  }

  async put<T>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T> {
    const response = await this.client.put<T>(url, data, config);
    return response.data;
  }

  async delete<T>(url: string, config?: AxiosRequestConfig): Promise<T> {
    const response = await this.client.delete<T>(url, config);
    return response.data;
  }

  // Set headers for trace propagation
  setHeader(key: string, value: string): void {
    this.client.defaults.headers.common[key] = value;
  }
}
