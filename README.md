# pair - reading status locally from a PurpleAir sensor

## Overview

The [PurpleAir](https://purpleair.com) sensors provide a local access
JSON API.  This package wraps that in a [Go
API](https://pkg.go.dev/zappem.net/pub/net/pair).

This `"zappem.net/pub/net/pair"` package includes an
[examples/query.go](examples/query.go) program to view some sensor
readings.

An example way to use this tool:
```
$ go run examples/query.go --sensor 192.168.4.51 --poll=5s
2024/07/06 11:24:35 temp=xx.xF(yy.yC) dewPt=zz.zF(uu.uC) hum=vv% pres=www.whPa AQIab=ff.f,gg.g
...
```

## License info

The `pair` package is distributed with the same BSD 3-clause license
as that used by [golang](https://golang.org/LICENSE) itself.

## Reporting bugs and feature requests

The package `pair` has been developed purely out of self-interest. If
you find a bug or want to suggest a feature addition, please use the
[bug tracker](https://github.com/tinkerator/pair/issues).
