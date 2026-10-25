# Xbox End-to-End Tests

OpenPak's Xbox End-to-End Tests.

| | |
| --- | --- |
| Network | Original Xbox |
| Language | Go module `github.com/openpak/xb-e2e` |
| Depends on | [`xb-account`](https://github.com/openpak/xb-account), [`xb-auth`](https://github.com/openpak/xb-auth), [`xb-gateway`](https://github.com/openpak/xb-gateway) |
| Licence | AGPL-3.0-only |

## Run

```sh
GOWORK=off GOPRIVATE='github.com/openpak/*' go test -race ./...
```

## Docs

- [Protocol](docs/protocol.md)
- [Evidence](docs/evidence.md)
- [Deploy](docs/deploy.md)
- [Notes](docs/notes.md)

## Licence

AGPL-3.0-only. See [LICENSE](LICENSE).
