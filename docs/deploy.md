# Deploy: xb-e2e

Test-only module (`github.com/openpak/xb-e2e`, no binaries, no image). It imports each service's public `testkit` package and `xb-gateway/sgwire`, pinned in go.mod to a released tag of each repo; bump the pins when a service changes its testkit surface.

```sh
GOWORK=off GOPRIVATE='github.com/openpak/*' go test -race ./...
```

CI (`.github/workflows/test.yml`) runs only on pushed `v*.*.*` tags and manual dispatch.
