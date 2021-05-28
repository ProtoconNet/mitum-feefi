### mitum-account-extension

*mitum-account-extension* is the data management case of mitum model, based on
[*mitum*](https://github.com/spikeekips/mitum) and [*mitum-currency*](https://github.com/spikeekips/mitum-currency).

#### Features,

* account: account address and keypair is not same.
* simple transaction: create document, update document, sign document.
* *mongodb*: as mitum does, *mongodb* is the primary storage.

#### Installation

> NOTE: at this time, *mitum* and *mitum-account-extension* is actively developed, so before building mitum-account-extension, you will be better with building the latest
mitum-account-extension source.
> `$ git clone https://github.com/protoconNet/mitum-account-extension`
>
> and then, add `replace github.com/spikeekips/mitum => <your mitum source directory>` to `go.mod` of *mitum-account-extension*.

Build it from source
```sh
$ cd mitum-account-extension
$ go build -ldflags="-X 'main.Version=v0.0.1'" -o ./mitum-account-extension ./main.go
```

#### Run

At the first time, you can simply start node with example configuration.

> To start, you need to run *mongodb* on localhost(port, 27017).

```
$ ./mitum-account-extension node init ./standalone.yml
$ ./mitum-account-extension node run ./standalone.yml
```

> Please check `$ ./mbs --help` for detailed usage.

#### Test

```sh
$ go clean -testcache; time go test -race -tags 'test' -v -timeout 20m ./... -run .
```
