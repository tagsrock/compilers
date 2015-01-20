lllc-server
===========

The Lovely Little Language Compiler: A web server and client for compiling ethereum languages.

Features
--------

- language agnostic (currently supports lll, serpent)
- client side and server side caching
- handles included files (at least for lll)

# Use the Golang API

```
bytecode, err := lllcserver.Compile("mycontract.lll")
```

# Use the CLI

### Compile Remotely

```
lllc-server compile --host http://lllc.erisindustries.com:8090 test.lll 
```

### Compile Locally 
Make sure you have the appropriate compiler installed and configured in `~/.decerver/languages/config.json`.

```
lllc-server compile --local test.lll
```

### Run a server yourself

```
lllc-server --port 9000
```

# Support

Run `lllc-server --help` or `lllc-server compile --help` for more info, or come talk to us on irc at #erisindustries and #erisindustries-dev.
