# glc

GitLab command line interface, especially for managing issues.

## Usage

Get issues:

```sh
$ glc -token=YOUR_PRIVATE_TOKEN -url=http://localhost:9664 -out=test.csv
```

## Install

```sh
$ go get code.google.com/p/go.text/encoding/japanese
$ go get code.google.com/p/go.text/transform
$ go get github.com/ksoichiro/glc
```

## Command line options

| Option   | Meaning |
| -------- | ------- |
| `-token` | Your private token. |
| `-url`   | GitLab root URL.    |
| `-out`   | Output CSV file.   |

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

