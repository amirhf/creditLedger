module github.com/amirhf/credit-ledger/services/accounts

go 1.24

require (
	github.com/amirhf/credit-ledger/proto v0.0.0
	github.com/go-chi/chi/v5 v5.1.0
	github.com/google/uuid v1.6.0
	github.com/lib/pq v1.10.9
	github.com/prometheus/client_golang v1.19.1
	github.com/segmentio/kafka-go v0.4.47
	google.golang.org/protobuf v1.36.3
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/klauspost/compress v1.15.9 // indirect
	github.com/pierrec/lz4/v4 v4.1.15 // indirect
	github.com/prometheus/client_model v0.5.0 // indirect
	github.com/prometheus/common v0.48.0 // indirect
	github.com/prometheus/procfs v0.12.0 // indirect
	golang.org/x/sys v0.17.0 // indirect
)

replace github.com/amirhf/credit-ledger/proto => ../../proto
