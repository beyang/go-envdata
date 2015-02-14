go-envdata
==========

go-envdata packages environment variables into Go binaries. Inspired by [go-bindata](https://sourcegraph.com/github.com/jteeuwen/go-bindata).

Tries to be shell/OS-agnostic, but currently tested with bash and *nix systems.

Usage
-----

Capture ambient environment:
```
go-envdata -pkg env -o env/env.go
```

Capture environment defined by config file:
```
env -i PATH=$PATH bash -c 'source my-config.sh; go-envdata;
```
or
```
env -i PATH=$PATH bash -c "./setup-environment.sh; go-envdata"
```
where `my-config.sh` and `setup-environment.sh` are files you define to setup the environment variables.
