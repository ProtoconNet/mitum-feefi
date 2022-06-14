### mitum-feefi

*mitum-feefi* is the feefi management case of mitum model, based on
[*mitum*](https://github.com/spikeekips/mitum) and [*mitum-currency*](https://github.com/spikeekips/mitum-currency).

#### Features,

* account: account address and keypair is not same.
* simple transaction: feefi pool register, pool policy update, pool deposit, pool withdraw.
* *mongodb*: as mitum does, *mongodb* is the primary storage.

#### Installation

> NOTE: at this time, *mitum* , *mitum-currency* and *mitum-currency-extension* is actively developed, so before building mitum-feefi, you will be better with building the latest mitum,
mitum-currency and mitum-currency-extension source.
> `$ git clone https://github.com/ProtoconNet/mitum-feefi`
>
> and then, add `replace github.com/spikeekips/mitum => <your mitum source directory>` and `replace github.com/spikeekips/mitum-currency => <your mitum-currency source directory>` and `replace github.com/spikeekips/mitum-currency-extension => <your mitum-currency-extension source directory>`to `go.mod` of *mitum-feefi*.

Build it from source
```sh
$ cd mitum-feefi
$ go build -ldflags="-X 'main.Version=v0.0.1'" -o ./mitum-feefi ./main.go
```

#### Run

At the first time, you can simply start node with example configuration.

> To start, you need to run *mongodb* on localhost(port, 27017).

```
$ ./mitum-feefi node init ./standalone.yml
$ ./mitum-feefi node run ./standalone.yml
```

> Please check `$ ./mbs --help` for detailed usage.

#### Test

```sh
$ go clean -testcache; time go test -race -tags 'test' -v -timeout 20m ./... -run .
```
