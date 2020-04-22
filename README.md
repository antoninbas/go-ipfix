# go-ipfix

## To install (on Linux)

### First, install libipfix
```
git clone https://github.com/antoninbas/libipfix.git
cd libipfix
./configure && make
sudo make install && sudo ldconfig
```

### Then build this module
```
git clone https://github.com/antoninbas/go-ipfix.git
cd go-ipfix
make
```

## Run example

```
./bin/example [--host <addr>] [--port <port>]
```
