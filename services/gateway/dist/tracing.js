"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.initTracing = initTracing;
const sdk_node_1 = require("@opentelemetry/sdk-node");
const exporter_trace_otlp_http_1 = require("@opentelemetry/exporter-trace-otlp-http");
const auto_instrumentations_node_1 = require("@opentelemetry/auto-instrumentations-node");
const resources_1 = require("@opentelemetry/resources");
const semantic_conventions_1 = require("@opentelemetry/semantic-conventions");
function initTracing() {
    const otelEndpoint = process.env.OTEL_EXPORTER_OTLP_ENDPOINT || 'http://localhost:4318';
    const sdk = new sdk_node_1.NodeSDK({
        resource: new resources_1.Resource({
            [semantic_conventions_1.SemanticResourceAttributes.SERVICE_NAME]: 'gateway',
            [semantic_conventions_1.SemanticResourceAttributes.SERVICE_VERSION]: '1.0.0',
            [semantic_conventions_1.SemanticResourceAttributes.DEPLOYMENT_ENVIRONMENT]: process.env.NODE_ENV || 'development',
        }),
        traceExporter: new exporter_trace_otlp_http_1.OTLPTraceExporter({
            url: `${otelEndpoint}/v1/traces`,
            headers: {},
        }),
        instrumentations: [
            (0, auto_instrumentations_node_1.getNodeAutoInstrumentations)({
                '@opentelemetry/instrumentation-http': {
                    enabled: true,
                    ignoreIncomingRequestHook: (req) => {
                        return req.url === '/healthz';
                    },
                },
                '@opentelemetry/instrumentation-nestjs-core': {
                    enabled: true,
                },
                '@opentelemetry/instrumentation-fs': {
                    enabled: false,
                },
                '@opentelemetry/instrumentation-dns': {
                    enabled: false,
                },
            }),
        ],
    });
    sdk.start();
    console.log('OpenTelemetry tracing initialized');
    console.log(`- Service: gateway`);
    console.log(`- OTLP Endpoint: ${otelEndpoint}/v1/traces`);
    process.on('SIGTERM', () => {
        sdk
            .shutdown()
            .then(() => console.log('Tracing terminated'))
            .catch((error) => console.error('Error terminating tracing', error))
            .finally(() => process.exit(0));
    });
    return sdk;
}
//# sourceMappingURL=tracing.js.map