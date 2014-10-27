# glc

[![Build Status](https://travis-ci.org/ksoichiro/glc.svg?branch=master)](https://travis-ci.org/ksoichiro/glc)

GitLab command line interface, especially for managing issues.

## Usage

Get issues:

```sh
$ glc issues -token=YOUR_PRIVATE_TOKEN -url=http://localhost:9664 -out=test.csv
(short)
$ glc i -token=YOUR_PRIVATE_TOKEN -url=http://localhost:9664 -out=test.csv
(with .glc)
$ glc i -out=test.csv
```

### Commands

| Command       | Meaning |
| ------------- | ------- |
| `issues`, `i` | Get issues |

### Options

| Option   | Default | Meaning |
| -------- | ------- | ------- |
| `-token` | (none)  | Your private token. |
| `-url`   | (none)  | GitLab root URL.    |
| `-out`   | (none)  | Output CSV file.   |
| `-csvEncoding` | sjis | Output encoding for CSV file.(sjis, utf8) |

## Install

```sh
$ go get code.google.com/p/go.text/encoding/japanese
$ go get code.google.com/p/go.text/transform
$ go get github.com/ksoichiro/glc
```

## Config

You don't have to set some command line options by creating config file named `~/.glc`.

```
token=YOUR_PRIVATE_TOKEN
url=http://YOUR_GITLAB_ROOT_URL
```

## License

Copyright (c) 2014 Soichiro Kashima  
Licensed under MIT license.  
See the bundled [LICENSE](LICENSE) file for details.

