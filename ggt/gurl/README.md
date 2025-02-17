# gurl

Gurl is a Go implemented CLI cURL-like tool for humans

`gurl [flags] [METHOD] URL [ITEM [ITEM]]`

Go implemented CLI cURL-like tool for humans. gurl can be used for testing, debugging, and generally interacting with
HTTP servers.

Inspired by [Httpie](https://github.com/jakubroztocil/httpie). Thanks to the author, Jakub.

Features:

1. 2023年05月18日 `INTERACTIVE=0 gurl` 禁止交互模式，否则 请求参数值/地址中的注入 @age 将被解析成插值模式，会要求从命令行输入
2. 2023年05月17日 多网卡时，指定通过指定 IP，连接服务端， export LOCAL_IP=ip
3. 2022年11月21日 支持深度美化 JSON 字符串（包括内嵌JSON字符串），例如  `gurl -b '{"xx":"{\"age\":100}"}' -pBf`
4. 2022年09月14日 支持指定 DNS 解析服务器，例如  `gurl http://a.cn:5003/echo -dns 127.0.0.1`
5. 2022年08月31日 文件下载时，支持断点续传
6. 2022年05月24日 支持
   从文件按行读取请求体，发送多次请求，例如 `gurl :9200/person1/_doc/@ksuid -b persons.txt:line -auth ZWxhc3RpYzoxcWF6WkFRIQ -n0 -pucb -ugly`
7. 2022年03月09日 支持 ca

   ```sh
   $ httplive &
   $ gurl https://localhost:5003/v -ca .cert/localhost.pem  -pb
   {
     "build": "2022-03-09T22:51:14+0800",
     "git": "19d2de6@2022-03-09T17:26:53+08:00",
     "go": "go1.17.8_darwin/amd64",
     "version": "1.3.5"
   }
   ```

8. 2022年02月21日 支持 timeout 参数，e.g.  `-timeout=0`
9. 2022年02月09日 支持多
   URL. `gurl 192.168.126.{16,18,182}:15002/kv -pb` `gurl 192.168.126.{16,18,182}:15002/kv -pb POST v==12345`
10. 2022年01月06日 支持查询值从文件中读取 `gurl -raw b.n:10014/query q==@query.sql`

```sh
$ gurl PUT httpbin.org/put hello=world
PUT /put? HTTP/1.1
Host: httpbin.org
Accept: application/json
Accept-Encoding: gzip, deflate
Content-Type: application/json
User-Agent: gurl/0.1.0

{"hello":"world"}

HTTP/1.1 200 OK
Access-Control-Allow-Credentials: true
Date: Thu, 06 Jan 2022 03:38:08 GMT
Content-Type: application/json
Content-Length: 486
Connection: keep-alive
Server: gunicorn/19.9.0
Access-Control-Allow-Origin: *

{
  "args": {},
  "data": "{\"hello\":\"world\"}\n",
  "files": {},
  "form": {},
  "headers": {
    "Accept": "application/json",
    "Accept-Encoding": "gzip, deflate",
    "Content-Length": "18",
    "Content-Type": "application/json",
    "Host": "httpbin.org",
    "User-Agent": "gurl/0.1.0",
    "X-Amzn-Trace-Id": "Root=1-61d66420-230d9c070cfeffbc477fc755"
  },
  "json": {
    "hello": "world"
  },
  "origin": "43.245.222.139",
  "url": "http://httpbin.org/put"
}
```

- [Main Features](#main-features)
- [Installation](#installation)
- [Usage](#usage)
- [HTTP Method](#http-method)
- [Request URL](#request-url)
- [Request Items](#request-items)
- [JSON](#json)
- [Forms](#forms)
- [HTTP Headers](#http-headers)
- [Authentication](#authentication)
- [Proxies](#proxies)

## Docker

    # Build the docker image
	$ docker build -t bingoohuang/ngg/gurl .

	# Run gurl in a container
	$ docker run --rm -it --net=host bingoohuang/ngg/gurl example.org

## Main Features

- Expressive and intuitive syntax
- Built-in JSON support
- Forms and file uploads
- HTTPS, proxies, and authentication
- Arbitrary request data
- Custom headers

## Installation

### Install with Modules - Go 1.11 or higher

If you only want to install the `gurl` tool:

	go get -u github.com/bingoohuang/ngg/gurl

If you want a mutable copy of source code:

	git clone https://github.com/bingoohuang/ngg/gurl ;# clone outside of GOPATH
	cd gurl
	go install

Make sure the `~/go/bin` is added into `$PATH`.

### Install without Modules - Before Go 1.11

	go get -u github.com/bingoohuang/ngg/gurl

Make sure the `$GOPATH/bin` is added into `$PATH`.

## Usage

Hello World:

	$ gurl beego.me

Synopsis:

	gurl [flags] [METHOD] URL [ITEM [ITEM]]

See also `gurl --help`.

### Examples

Basic settings - [HTTP method](#http-method), [HTTP headers](#http-headers) and [JSON](#json) data:

	$ gurl PUT example.org X-API-Token:123 name=John

Any custom HTTP method (such as WebDAV, etc.):

	$ gurl -method=PROPFIND example.org name=John

Submitting forms:

	$ gurl -form=true POST example.org hello=World

See the request that is being sent using one of the output options:

	$ gurl -print="Hhb" example.org

Use Github API to post a comment on an issue with authentication:

	$ gurl -a USERNAME POST https://api.github.com/repos/bingoohuang/ngg/gurl/issues/1/comments body='gurl is awesome!'

Upload a file using redirected input:

	$ gurl example.org < file.json

Download a file and save it via redirected output:

	$ gurl example.org/file > file

Download a file wget style:

	$ gurl -download=true example.org/file

Set a custom Host header to work around missing DNS records:

	$ gurl localhost:8000 Host:example.com

Following is the detailed documentation. It covers the command syntax, advanced usage, and also features additional
examples.

## HTTP Method

The name of the HTTP method comes right before the URL argument:

	$ gurl DELETE example.org/todos/7

which looks similar to the actual Request-Line that is sent:

DELETE /todos/7 HTTP/1.1

When the METHOD argument is omitted from the command, gurl defaults to either GET (if there is no request data) or
POST (
with request data).

## Request URL

The only information gurl needs to perform a request is a URL. The default scheme is, somewhat unsurprisingly, http://,
and can be omitted from the argument – `gurl example.org` works just fine.

Additionally, curl-like shorthand for localhost is supported. This means that, for example :3000 would expand
to http://localhost:3000 If the port is omitted, then port 80 is assumed.

	$ gurl :/foo

	GET /foo HTTP/1.1
	Host: localhost

	$ gurl :3000/bar

	GET /bar HTTP/1.1
	Host: localhost:3000

	$ gurl :

	GET / HTTP/1.1
	Host: localhost

If you find yourself manually constructing URLs with query string parameters on the terminal, you may appreciate
the `param=value` syntax for appending URL parameters so that you don't have to worry about escaping the & separators.
To search for gurl on Google Images you could use this command:

	$ gurl GET www.google.com search=gurl tbm=isch

	GET /?search=gurl&tbm=isch HTTP/1.1

## Request Items

There are a few different request item types that provide a convenient mechanism for specifying HTTP headers, simple
JSON and form data, files, and URL parameters.

They are key/value pairs specified after the URL. All have in common that they become part of the actual request that is
sent and that their type is distinguished only by the separator used: `:`, `=`, `:=`, `@`, `=@`, and `:=@`. The ones
with an `@` expect a file path as value.

| Item Type                                          | Description                                                                                                                                                                 |
|----------------------------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| HTTP Headers `Name:Value`                          | Arbitrary HTTP header, e.g. `X-API-Token:123`.                                                                                                                              |
| Data Fields `field=value`                          | Request data fields to be serialized as a JSON object (default), or to be form-encoded (--form, -f).                                                                        |
| Form File Fields `field@/dir/file`                 | Only available with `-form`, `-f`. For example `screenshot@~/Pictures/img.png`. The presence of a file field results in a `multipart/form-data` request.                    |
| Form Fields from file `field=@file.txt`            | read content from file as value                                                                                                                                             |
| Raw JSON fields `field:=json`, `field:=@file.json` | Useful when sending JSON and one or more fields need to be a Boolean, Number, nested Object, or an Array, e.g., meals:='["ham","spam"]' or pies:=[1,2,3] (note the quotes). |

You can use `\` to escape characters that shouldn't be used as separators (or parts thereof). For instance, foo\==bar
will become a data key/value pair (foo= and bar) instead of a URL parameter.

You can also quote values, e.g. `foo="bar baz"`.

## JSON

JSON is the lingua franca of modern web services and it is also the implicit content type gurl by default uses:

If your command includes some data items, they are serialized as a JSON object by default. gurl also automatically sets
the following headers, both of which can be overridden:

| header       | value            |
|--------------|------------------|
| Content-Type | application/json |
| Accept       | application/json |

You can use --json=true, -j=true to explicitly set `Accept` to `application/json` regardless of whether you are sending
data (it's a shortcut for setting the header via the usual header notation – `gurl url Accept:application/json`).

Simple example:

	$ gurl PUT example.org name=John email=john@example.org
	PUT / HTTP/1.1
	Accept: application/json
	Accept-Encoding: gzip, deflate
	Content-Type: application/json
	Host: example.org

	{
	    "name": "John",
	    "email": "john@example.org"
	}

Even custom/vendored media types that have a json format are getting detected, as long as they implement a json type
response and contain a `json` in their declared form:

	$ gurl GET example.org/user/1 Accept:application/vnd.example.v2.0+json
	GET / HTTP/1.1
	Accept: application/vnd.example.v2.0+json
	Accept-Encoding: gzip, deflate
	Content-Type: application/vnd.example.v2.0+json
	Host: example.org

	{
	    "name": "John",
	    "email": "john@example.org"
	}

Non-string fields use the := separator, which allows you to embed raw JSON into the resulting object. Text and raw JSON
files can also be embedded into fields using =@ and :=@:

	$ gurl PUT api.example.com/person/1 \
    name=John \
    age:=29 married:=false hobbies:='["http", "pies"]' \  # Raw JSON
    description=@about-john.txt \   # Embed text file
    bookmarks:=@bookmarks.json      # Embed JSON file

	PUT /person/1 HTTP/1.1
	Accept: application/json
	Content-Type: application/json
	Host: api.example.com

	{
	    "age": 29,
	    "hobbies": [
	        "http",
	        "pies"
	    ],
	    "description": "John is a nice guy who likes pies.",
	    "married": false,
	    "name": "John",
	    "bookmarks": {
	        "HTTPie": "http://httpie.org",
	    }
	}

Send JSON data stored in a file (see redirected input for more examples):

	$ gurl POST api.example.com/person/1 < person.json

## Forms

Submitting forms are very similar to sending JSON requests. Often the only difference is in adding the `-form=true`
, `-f` option, which ensures that data fields are serialized correctly and Content-Type is set
to, `application/x-www-form-urlencoded; charset=utf-8`.

It is possible to make form data the implicit content type instead of JSON via the config file.

### Regular Forms

	$ gurl -f POST api.example.org/person/1 name='John Smith' \
    email=john@example.org

	POST /person/1 HTTP/1.1
	Content-Type: application/x-www-form-urlencoded; charset=utf-8

	name=John+Smith&email=john%40example.org

### File Upload Forms

If one or more file fields is present, the serialization and content type is `multipart/form-data`:

	$ gurl -f POST example.com/jobs name='John Smith' cv@~/Documents/cv.pdf

The request above is the same as if the following HTML form were submitted:

---

```markdown
<form enctype="multipart/form-data" method="post" action="http://example.com/jobs">
    <input type="text" name="name" />
    <input type="file" name="cv" />
</form>
```

---

Note that `@` is used to simulate a file upload form field.

## HTTP Headers

To set custom headers you can use the Header:Value notation:

	$ gurl example.org  User-Agent:Bacon/1.0  'Cookie:valued-visitor=yes;foo=bar'  \
    X-Foo:Bar  Referer:http://beego.me/

	GET / HTTP/1.1
	Accept: */*
	Accept-Encoding: gzip, deflate
	Cookie: valued-visitor=yes;foo=bar
	Host: example.org
	Referer: http://beego.me/
	User-Agent: Bacon/1.0
	X-Foo: Bar

There are a couple of default headers that gurl sets:

	GET / HTTP/1.1
	Accept: */*
	Accept-Encoding: gzip, deflate
	User-Agent: gurl/<version>
	Host: <taken-from-URL>

Any of the default headers can be overridden.

## Authentication

Basic auth:

	$ gurl -a=username:password example.org

## Proxies

You can specify proxies to be used through the --proxy argument for each protocol (which is included in the value in
case of redirects across protocols):

	$ gurl --proxy=http://10.10.1.10:3128 example.org

With Basic authentication:

	$ gurl --proxy=http://user:pass@10.10.1.10:3128 example.org

You can also configure proxies by environment variables HTTP_PROXY and HTTPS_PROXY, and the underlying Requests library
will pick them up as well. If you want to disable proxies configured through the environment variables for certain
hosts, you can specify them in NO_PROXY.

In your ~/.bash_profile:

	export HTTP_PROXY=http://10.10.1.10:3128
	export HTTPS_PROXY=https://10.10.1.10:1080
	export NO_PROXY=localhost,example.com

## usages examples

```sh
$ gurl POST name=@姓名_1 name2=@姓名_2 name3=@姓名_1 name4=@姓名_2 name5=@姓名 name6=@姓名
POST / HTTP/1.1
Host: dry.run.url
Accept: application/json
Accept-Encoding: gzip, deflate
Content-Type: application/json
Gurl-Date: Fri, 06 May 2022 00:45:34 GMT
User-Agent: gurl/1.0.0

{
  "name": "轩辕局徖",
  "name2": "夏嬮耤",
  "name3": "轩辕局徖",
  "name4": "夏嬮耤",
  "name5": "寇碭韅",
  "name6": "尤搮蜃"
}
```

## resources

1. [rs/curlie](https://github.com/rs/curlie) The power of curl, the ease of use of httpie.
2. [Hurl](https://github.com/Orange-OpenSource/hurl) is a command line tool that runs HTTP requests defined in a simple
   plain text format.
3. [httpretty](https://github.com/henvic/httpretty) prints the HTTP requests you make with Go pretty on your terminal.

## tlcp support

```sh
# 生成根证书和服务端证书
$ tlcp -m localhost
# 生成客户端证书
$ tlcp -c -m localhost
# 启动国密 https 服务
$ tlcp -l :8443 --ecdhe --verify-client-cert -r root.pem -C localhost.pem -C localhost.key -C localhost.pem -C localhost.key
listen on :8443
```

```sh
# 调用验证
$ TLS_VERIFY=1 TLCP=1 TLCP_CERTS=localhost-client.pem,localhost-client.key,localhost-client.pem,localhost-client.key CERT=root.pem gurl https://localhost:8443 
Hello GoTLCP!
```
