# OpenStackSwift

**Only for development and test purpose.**

OpenStackSwift is a Golang server that responds to the same calls Openstack Swift responds to. It's a convenient way to use Swift out-of-the-box without any fancy dependencies and configuration. It aims to be **portable** and **lightweight**.

OpenStackSwift doesn't support all of the Swift command set, but the basic ones like upload (with TTL), download, list, copy, authentication, and make containers are supported. More coming soon™.

[https://hub.docker.com/r/mdouchement/openstackswift](https://hub.docker.com/r/mdouchement/openstackswift)

## Running
```bash
$ go run cmd/swift/main.go server -b localhost

# http://localhost:5000
# tenant: test
# username: tester
# password: testing

# http://localhost:5000/v3
# tenant: test
# domain: Default
# region: RegionOne
# name: tester
# password: testing

# storage token: tk_tester
```

Environment variables:
```
SWIFT_STORAGE_TENANT
SWIFT_STORAGE_DOMAIN
SWIFT_STORAGE_USERNAME
SWIFT_STORAGE_PASSWORD
```

## License

MIT. See the [LICENSE](https://github.com/mdouchement/openstackswift/blob/master/LICENSE) for more details.

## Development

### Building
```
go build -ldflags "-s -w" -o swift ./cmd/swift/main.go
```

### Testing
Running tests with coverage
```
go test -coverpkg=./internal/database,./internal/model,./internal/scheduler,./internal/storage,./internal/webserver,./internal/webserver/middleware,./internal/webserver/serializer,./internal/webserver/service,./internal/webserver/weberror,./internal/xpath,./tests -coverprofile=cprof.out -v ./tests/
go tool cover -html=cprof.out -o coverage.html

```
### Build docker

```
docker build . -t openstackswift
```
### Contributing

1. Fork it
2. Create your feature branch (git checkout -b my-new-feature)
3. Commit your changes (git commit -am 'Add some feature')
4. Ensure its building and that the tests pass
5. Push to the branch (git push origin my-new-feature)
6. Create new Pull Request
