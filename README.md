[![GoDoc](https://godoc.org/github.com/eris-ltd/eris-compilers?status.png)](https://godoc.org/github.com/eris-ltd/eris-compilers)

[![Circle CI](https://circleci.com/gh/eris-ltd/eris-compilers.svg?style=svg)](https://circleci.com/gh/eris-ltd/eris-compilers)

eris-compilers
===========

The Lovely Little Language Compiler: A web server and client for compiling smart contract languages.

# Features

- language agnostic (currently supports lll, serpent2.0, solidity)
- returns smart contract abis (for serpent and solidity)
- handles included files recursively with regex matching
- client side and server side caching
- configuration file with per-language options
- local proxy server for compiling from languages other than go
- easily extensible to new languages

Monax Industries' own public facing LLLC-server (at https://compilers.monax.io) is hardcoded into the source,
so you can start compiling smart contract language right out of the box with no extra tools required.

If you want to use your own server, or default to compiling locally, or otherwise adjust configuration settings,
see the config file at `~/.eris/languages/config.json`.

# How to play

## Using the Golang API

```
bytecode, err := compilers.Compile("mycontract.lll")
```

The language is determined automatically from extension. If you want to compile literal expressions,
you must specify the language explicitly, ie.

```
bytecode, err := compilers.CompileLiteral("[0x5](+ 4 @0x3)", "lll")
```

## Using the CLI

#### Compile Remotely

```
eris-compilers compile --host https://compilers.monax.io:9090 test.lll
```

Leave out the `--host` flag to default to the url in the config.

#### Compile Locally
Make sure you have the appropriate compiler installed and configured (you may need to adjust the `cmd` field in the config file)

```
eris-compilers compile --local test.lll
```

#### Run a server yourself

```
eris-compilers --port 9000
```

## Using the json-rpc proxy server

If you are coding in another language and would like to use the eris-compilers client without wrapping the command line, run a proxy server and send it a simple http-json request.

To run the proxy:

```
eris-compilers proxy --port 9099
```

And the JSON request:

```
{
 source:"myfile.se"
}
```

Or, to compile literals:

```
{
 source: "x=5",
 literal: true,
 language: "se"
}
```

The response JSON looks like:

```
{
 bytecode:"600580600b60003960105660056020525b6000f3",
 error:""
}
```

To test, stick one of the above JSON requests into `file.json` and run

```
curl -X POST -d @file.json http://localhost:9099 --header "Content-Type:application/json"
```

# Install

The eris-compilers itself can be installed with

```
go get github.com/eris-ltd/eris-compilers/cmd/eris-compilers
```

Installing the actual compilers is a bit more involved. :(

See [ethereum wiki](http://solidity.readthedocs.org/en/latest/installing-solidity.html#ubuntu) and add the ethereum and ethereum-dev PPA's (no need for qt)

Note, thelonious and its Genesis Doug were build on a previous version of the languages (before the ABI spec) and so currently only support PoC6 LLL and Serpent 1.0.
But epm works fine using Solidity and Serpent on standard ethereum chains.

To install the compilers:

```
sudo add-apt-repository ppa:ethereum/ethereum
sudo add-apt-repository ppa:ethereum/ethereum-dev
sudo apt-get update
sudo apt-get install lllc sc solc
```

Now the final thing is make sure the configuration paths are properly set.
Running `epm init` (assuming epm is installed) should create the config file at `~/.eris/languages/config.json`.
Edit the `cmd` field for each language to have the correct path.

# Support

Run `eris-compilers --help` or `eris-compilers compile --help` for more info, or come talk to us on Slack at https://slack.monax.io

If you are working on a language, and would like to have it supported, please create an issue! Note it is possible to add new languages simply by editing the config file, without having to recompile the eris-compilers source code.

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

