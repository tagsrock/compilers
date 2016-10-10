[![GoDoc](https://godoc.org/github.com/eris-ltd/eris-compilers?status.png)](https://godoc.org/github.com/eris-ltd/eris-compilers)

[![Circle CI](https://circleci.com/gh/eris-ltd/eris-compilers.svg?style=svg)](https://circleci.com/gh/eris-ltd/eris-compilers)

eris-compilers
===========

A web server and client for compiling smart contract languages.

# Features

- Supports Solidity
- returns smart contract abis
- handles included files recursively with regex matching
- client side and server side caching
- configuration file with per-language options
- easily extensible to new languages

Monax Industries' own public facing compiler server (at https://compilers.monax.io) is hardcoded into the source,
so you can start compiling smart contract language right out of the box with no extra tools required.

If you want to use your own server, look below on how to get set up.

# How to play

## Using the Golang API

```
client "github.com/eris-ltd/eris-compilers/network"

url := "https://compilers.monax.io:9099/compile"
filename := "maSolcFile.sol"
optimize := true
librariesString := "maLibrariez:0x1234567890"

output, err := client.BeginCompile(url, filename, optimize, librariesString)

contractName := output.Objects[0].Objectname // contract C would give you C here
binary := output.Objects[0].Bytecode //gives you binary
abi := output.Objects[0].ABI //gives you the ABI
```

## Using the CLI

#### Compile Remotely

```
eris-compilers compile test.sol
```

Will by default compile directly using the monax servers. You can configure this to call a different server by checking out the ```--help``` option.

#### Compile Locally
Make sure you have the appropriate compiler installed and configured (you may need to adjust the `cmd` field in the config file)

```
eris-compilers compile --local test.sol
```

#### Run a server yourself

```
eris-compilers server --no-ssl
```

will run a simple http server. For encryption, pass in a key with the `--key` flag, or a certificate with the `--cert` flag and drop the `--no-ssl`.

# Install

The eris-compilers itself can be installed with

```
go get github.com/eris-ltd/eris-compilers/cmd/eris-compilers
```

Currently the compilers server supports only solidity, which can be readily and easily installed [here](http://solidity.readthedocs.org/en/latest/installing-solidity.html)

You can also start the compilers service in a docker container via the eris CLI with a simple `eris services start compilers`.

# Support

Run `eris-compilers server --help` or `eris-compilers compile --help` for more info, or come talk to us on [Slack](https://slack.monax.io).

If you are working on a language, and would like to have it supported, please create an issue!

# Contributions

Are Welcome! Before submitting a pull request please:

* read up on [How The Marmots Git](https://github.com/eris-ltd/coding/wiki/How-The-Marmots-Git)
* fork from `develop`
* go fmt your changes
* have tests
* pull request
* be awesome

That's pretty much it. 

See our [CONTRIBUTING.md](.github/CONTRIBUTING.md) and [PULL_REQUEST_TEMPLATE.md](.github/PULL_REQUEST_TEMPLATE.md) for more details.

Please note that this repository is GPLv3.0 per the LICENSE file. Any code which is contributed via pull request shall be deemed to have consented to GPLv3.0 via submission of the code (were such code accepted into the repository).

# Bug Reporting

Found a bug in our stack? Make an issue!

The [issue template](.github/ISSUE_TEMPLATE.md] specifies what needs to be included in your issue and will autopopulate the issue.

# License

[Proudly GPL-3](http://www.gnu.org/philosophy/enforcing-gpl.en.html). See [license file](https://github.com/eris-ltd/eris-cli/blob/master/LICENSE.md).

