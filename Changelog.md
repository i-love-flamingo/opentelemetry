# Changelog

## Version v0.2.1 (2025-02-17)

### Chores and tidying

- **deps:** update opentelemetry-go-contrib monorepo to v0.59.0 (#40) (a66fa445)
- **deps:** update opentelemetry-go monorepo (#43) (94cff2dc)
- **deps:** update module flamingo.me/flamingo/v3 to v3.13.0 (#41) (a1ddcfc7)
- **deps:** update module flamingo.me/dingo to v0.3.0 (#39) (d57842de)
- **deps:** update opentelemetry-go monorepo (#34) (bb02dbd6)
- **deps:** update module go.opentelemetry.io/contrib/instrumentation/runtime to v0.56.0 (#32) (e6469a15)
- **deps:** update module go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp to v0.56.0 (#31) (a6d1d956)
- **deps:** update opentelemetry-go monorepo (#30) (62c03d0d)
- **deps:** update module github.com/prometheus/client_golang to v1.20.5 (#33) (2d2b18ce)
- **deps:** update module github.com/vektra/mockery/v2 to v2.46.3 (#29) (978f8c91)
- **deps:** update module github.com/vektra/mockery/v2 to v2.46.1 (#27) (d9350926)

## Version v0.2.0 (2024-09-24)

### Features

- register url prefix sampler for flamingo, resolves #17 (#24) (9eb27158)
- drop Jaeger support (#16) (3b88eb8d)

### Fixes

- **#18:** implement shutdown for tracer and meter (#21) (20dc9d84)
- rename to ignoreParentDecision (f951332e)

### Ops and CI/CD

- introduce flamingo code standards (#15) (26c22c80)

### Chores and tidying

- bump go version in tests (#26) (578df66a)
- **deps:** update module github.com/prometheus/client_golang to v1.20.4 (#22) (057e7c9f)
- **deps:** update module go.opentelemetry.io/contrib/instrumentation/runtime to v0.55.0 (#20) (6b68d5e6)
- **deps:** update module github.com/vektra/mockery/v2 to v2.46.0 (#23) (8d3feba3)
- **deps:** update module go.opentelemetry.io/otel/bridge/opencensus to v1 (#10) (76ced0f8)
- **deps:** update module go.opentelemetry.io/contrib/instrumentation/runtime to v0.54.0 (#8) (d10f2a65)
- **deps:** update module go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp to v0.54.0 (#7) (cbb0ac4d)
- **deps:** update opentelemetry-go monorepo (#6) (d206aa96)
- **deps:** update module github.com/prometheus/client_golang to v1.20.3 (#13) (05ba6370)
- **deps:** update module flamingo.me/flamingo/v3 to v3.10.1 (#14) (4dc3d21b)
- **deps:** update dependency go to v1.22.6 (#11) (18dceddd)
- **deps:** update actions/setup-go action to v5 (#9) (4df08762)
- **deps:** update module github.com/prometheus/client_golang to v1.19.1 (#4) (b7ff4f15)
- **deps:** update module flamingo.me/flamingo/v3 to v3.9.0 (#3) (d506e74e)

## Version v0.1.0 (2023-11-15)

### Features

- add otlp exporters for traces (d13d92ce)
- connect opencensus bridge for metrics and expose on systemendpoint (81d7ebcc)
- properly set up opentelemetry resource (f5c74719)
- implement opentelemetry module (145d8905)
- initial commit (342e11e8)

### Tests

- add unit tests for sampler (fea832c9)

### Refactoring

- improve configuration naming (ce381600)

### Ops and CI/CD

- add renovate bot (ea52b2c6)
- add semanticore releases (fc97b06f)
- add basic CI pipeline (8ae6d40e)

### Documentation

- add Jaeger v1.35 hint (f24a12ec)
- add Readme.md (c59cc62f)
- add license (8b97c4b7)

### Chores and tidying

- goimports (65b13524)
- clean up go.mod (deee9230)

