lllc-server
===========

Web Server and client for compiling ethereum languages.

Features
--------

- language agnostic (currently supports lll, serpent)
- client side and server side caching

# Use the API

```
bytecode, err := lllcserver.Compile("mycontract.lll")
```

