# Changelog

## Version v0.1.1 (2024-08-08)

### Fixes

- rename to ignoreParentDecision (f951332e)

### Chores and tidying

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

