lllc-server
===========

Web Server and client for compiling ethereum languages.

Features
--------

- language agnostic (currently supports lll, serpent)
- client side and server side caching
- handles recursive includes using "include signatures" 

# Run the client

lllc-server -h http://lllc.erisindustries.com -c mycontract.lll

# Run the server

Note: You should have the respective compilers installed if you plan to offer compilation as a service

```
lllc-server -port 80
```

# Use the API

```
bytecode, err := lllcserver.Compile("mycontract.lll")
```

