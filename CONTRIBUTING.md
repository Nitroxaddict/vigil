## Prerequisites
To contribute code changes to this project you will need the following development kits.
 * [Go](https://golang.org/doc/install)
 * [Docker](https://docs.docker.com/engine/installation/)
 
As watchtower utilizes go modules for vendor locking, you'll need at least Go 1.11.
You can check your current version of the go language as follows:
```bash
  ~ $ go version
  go version go1.12.1 darwin/amd64
```


## Checking out the code
Do not place your code in the go source path.
```bash
git clone git@github.com:<yourfork>/watchtower.git
cd watchtower
```

## Building and testing
Vigil is a Go application and is built with go commands. The following commands assume that you are at the root level of your repo.
```bash
go build                               # compiles and packages an executable binary, vigil
go test ./... -v                       # runs tests with verbose output
./vigil                                # runs the application (outside of a container)
```

If you dont have it enabled, you'll either have to prefix each command with `GO111MODULE=on` or run `export GO111MODULE=on` before running the commands. [You can read more about modules here.](https://github.com/golang/go/wiki/Modules)

To build a Vigil image of your own, use the release Dockerfile in `dockerfiles/`:

```bash
sudo docker build . -f dockerfiles/Dockerfile.release -t nitroxaddict/vigil:dev # to build an image from local files
```