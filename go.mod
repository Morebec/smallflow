module github.com/morebec/smallflow

go 1.24.8

replace github.com/morebec/go-misas => ./../go-misas-back

require (
	github.com/alitto/pond/v2 v2.5.0
	github.com/morebec/go-misas v0.0.0-00010101000000-000000000000
	github.com/samber/lo v1.49.1
)

require (
	github.com/google/uuid v1.6.0 // indirect
	github.com/lib/pq v1.10.9 // indirect
	go.opentelemetry.io/otel v1.35.0 // indirect
	go.opentelemetry.io/otel/trace v1.35.0 // indirect
	golang.org/x/text v0.24.0 // indirect
)
