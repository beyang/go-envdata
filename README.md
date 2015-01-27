go-envdata
==========

go-envdata packages environment variables into Go binaries. Inspired by [go-bindata](https://sourcegraph.com/github.com/jteeuwen/go-bindata)

Currently tested with bash and *nix systems. YMMV with other shells/architectures.

Usage
-----

Capture ambient environment:
```
go-envdata -pkg env -o env/env.go
```

Capture environment in a file:
```
env -i bash -c 'source ~/my-custom-config.sh; go-envdata;
```
