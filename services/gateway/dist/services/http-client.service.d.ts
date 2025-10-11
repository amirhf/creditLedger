import { AxiosRequestConfig } from 'axios';
export interface HttpClientConfig {
    baseURL: string;
    timeout?: number;
    retries?: number;
}
export declare class HttpClientService {
    private readonly logger;
    private client;
    private retries;
    constructor(config: HttpClientConfig);
    get<T>(url: string, config?: AxiosRequestConfig): Promise<T>;
    post<T>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T>;
    put<T>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T>;
    delete<T>(url: string, config?: AxiosRequestConfig): Promise<T>;
    setHeader(key: string, value: string): void;
}
