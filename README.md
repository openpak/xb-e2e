# xb-e2e

In-process end-to-end test of the native original-Xbox sign-in slice: provisioning
and enrolment in xb-account (fake core), MACS/AS/TGS in xb-auth, SG-init admission
in xb-gateway, over real sockets and mutual TLS. The console side is a test client
written against the same wire structures; no retail console has run it.

Test-only module (`github.com/openpak/xb-e2e`, no binaries, no image). It imports each
service's public `testkit` package and `xb-gateway/sgwire`, pinned in go.mod to a
released tag of each repo; bump the pins when a service changes its testkit surface.

```sh
GOWORK=off GOPRIVATE='github.com/openpak/*' go test -race ./...
```

CI (`.github/workflows/test.yml`) runs only on pushed `v*.*.*` tags and manual dispatch.
