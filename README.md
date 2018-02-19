# Baton

Baton is a load testing tool written in Go. It currently supports GET, POST, PUT, and DELETE requests. 

### Usage

Baton currently supports the following options:
```
  -b string
    	Body (use instead of -f)
  -c int
    	Number of concurrent requests (default 1)
  -f string
    	File path to file to be used as the body (use instead of -b)
  -i	Ignore TLS/SSL certificate validation
  -m string
    	HTTP Method (GET,POST,PUT,DELETE) (default "GET")
  -o	Supress output, no results will be printed to stdout
  -r int
    	Number of requests (use instead of -t) (default 1)
  -t int
    	Duration of testing in seconds (use instead of -r)
  -u string
    	URL to run against
  -w int
    	Number of seconds to wait before running test
  -z string
    	Read requests from a file
```

A basic example which will use 10 workers to send 200,000 requests is as follows: 

```bash
$ baton -u http://localhost:8080/test -c 10 -r 200000
```

Instead of the number of requests, you can specify the time (in seconds) during which the
requests should be sent. Baton will wait for all the responses to be received before reporting the results.

##### Requests file

When specifying a file to load requests from (`-z filename`), the file should be of CSV format ([RFC-4180](https://tools.ietf.org/html/rfc4180))
```
<method>,<url>,[<body>],[<header-key>:<header-value>, ...]
...
```

You can have one or more headers at the end separated by `,`

For example:

```
POST,http://localhost:8888,body,Accept: application/xml,Content-type: Secret
GET,http://localhost:8888,,,
```

##### Example Output:

```
====================== Results ======================
Total requests:                               1254155
Time taken to complete requests:        10.046739294s
Requests per second:                           124832
===================== Breakdown =====================
Number of connection errors:                        0
Number of 1xx responses:                            0
Number of 2xx responses:                      1254155
Number of 3xx responses:                            0
Number of 4xx responses:                            0
Number of 5xx responses:                            0
=====================================================

```

### Features which are on the horizon...
* Dynamic generation of data based on a template
* Testing REST endpoints with dynamically generated keys

### Installing Baton

Installation with Go is as easy as running `go get`.

```sh
$ go get -u github.com/americanexpress/baton
```

Binary releases are [available](https://github.com/americanexpress/baton/releases).

### Running Baton in docker

To build the image run:
```Bash
$ docker build -t baton .
```

Alternatively, update the docker-compose.yml file to meet your needs and run:
```bash
$ docker-compose up
```



## Dependency Management
[Dep](https://github.com/golang/dep) is currently being utilized as the dependency manager for Baton.
Details of how to use dep can be found on https://golang.github.io/dep/.

Before updating any dependencies, ensure you have fully tested all functionality.



## Contributing
We welcome Your interest in the American Express Open Source Community on Github.
Any Contributor to any Open Source Project managed by the American Express Open
Source Community must accept and sign an Agreement indicating agreement to the
terms below. Except for the rights granted in this Agreement to American Express
and to recipients of software distributed by American Express, You reserve all
right, title, and interest, if any, in and to Your Contributions. Please [fill out the Agreement](https://cla-assistant.io/americanexpress/).

Please feel free to open pull requests and see [CONTRIBUTING.md](./CONTRIBUTING.md) for commit formatting details.

## License
Any contributions made under this project will be governed by the [Apache License 2.0](./LICENSE.md).

## Code of Conduct
This project adheres to the [American Express Community Guidelines](./CODE_OF_CONDUCT.md).
By participating, you are expected to honor these guidelines.
