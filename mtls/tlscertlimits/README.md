# TLS certificates limits

This repository is the code and configuration used to generate TLS certificates for my blog entry [Exploring TLS certificates and their limits](https://0x00.cl/blog/2024/exploring-tls-certs/)

It will generate certificates with validity dates from Jan 1, 1950 to Dec 31, 9999. It will also generate RSA keys with 16384 bits, which depending on your hardware it could take a while.

## Requirements

* Python 3.12 (Only tested with 3.12 but probably works with 3.8 too)
* [Cryptography](https://cryptography.io/en/latest/)
* (Optional) Poetry

## Running

Clone the repository
```bash
$ git clone https://gitlab.com/0x00cl/tlscertlimits.git
$ cd tlscertlimits
```

### Using Poetry

```bash
$ poetry install
$ poetry run python tlscert.py
```

### Using pip
```bash
$ pip install -U 'cryptography>=42.0'
$ python tlscert.py
```

Once it finished running it should generate 7 files between .keys and .pems.

```bash
$ ls -l certs
total 148
-rw-r--r--. 1 tomas tomas 65557 Jul 03 12:00 myCABundle.pem
-rw-r--r--. 1 tomas tomas  1704 Jul 03 12:00 myCAinter.key
-rw-r--r--. 1 tomas tomas  2183 Jul 03 12:00 myCAinter.pem
-rw-r--r--. 1 tomas tomas  1704 Jul 03 12:00 myCA.key
-rw-r--r--. 1 tomas tomas  2122 Jul 03 12:00 myCA.pem
-rw-r--r--. 1 tomas tomas  1704 Jul 03 12:00 myCAweb.key
-rw-r--r--. 1 tomas tomas 61252 Jul 03 12:00 myCAweb.pem
```

## Caddy

Using caddy you can test the generated TLS certificates, the repository includes a Caddyfile that uses ports 8443 and 8080 to serve the content.

### Running

An easy way to get caddy running is simply [downloading the binary](https://github.com/caddyserver/caddy/releases/tag/v2.8.4) and putting it in the caddy directory.
Then you can simply run it with `./caddy run -c Caddyfile`

### Making a request

Using curl you can simply execute

```bash
$ curl -vk https://localhost:8443
* Host localhost:8443 was resolved.
* IPv6: ::1
* IPv4: 127.0.0.1
*   Trying [::1]:8443...
* Connected to localhost (::1) port 8443
* ALPN: curl offers h2,http/1.1
* TLSv1.3 (OUT), TLS handshake, Client hello (1):
* TLSv1.3 (IN), TLS handshake, Server hello (2):
* TLSv1.3 (IN), TLS handshake, Encrypted Extensions (8):
* TLSv1.3 (IN), TLS handshake, Certificate (11):
* TLSv1.3 (IN), TLS handshake, CERT verify (15):
* TLSv1.3 (IN), TLS handshake, Finished (20):
* TLSv1.3 (OUT), TLS change cipher, Change cipher spec (1):
* TLSv1.3 (OUT), TLS handshake, Finished (20):
* SSL connection using TLSv1.3 / TLS_AES_128_GCM_SHA256 / x25519 / RSASSA-PSS
* ALPN: server accepted h2
* Server certificate:
*  subject: C=WW; ST=WWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWW; L=WWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWW; O=WWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWW; CN=0x00.cl Web; C=WW; ST=WWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWW; L=WWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWW; O=WWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWW; CN=0x00.cl Web; C=WW; ST=WWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWW; L=WWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWW; O=WWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWW; CN=0x00.cl Web; C=WW; ST=WWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWW; L=WWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWW; O=WWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWW; CN=0x00.cl Web; C=WW; ST=WWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWW; L=WWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWW; O=WWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWW; CN=0x00.cl Web; C=WW; ST=WWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWW; L=WWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWW
*  start date: Jan  1 00:00:00 1950 GMT
*  expire date: Dec 31 23:59:59 9999 GMT
*  issuer: C=WW; ST=WWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWW; L=WWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWW; O=WWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWW; CN=0x00.cl CA Intermediate
*  SSL certificate verify result: self-signed certificate in certificate chain (19), continuing anyway.
*   Certificate level 0: Public key type RSA (16384/112 Bits/secBits), signed using RSA-SHA3-512
*   Certificate level 1: Public key type RSA (16384/112 Bits/secBits), signed using RSA-SHA3-512
*   Certificate level 2: Public key type RSA (16384/112 Bits/secBits), signed using RSA-SHA3-512
* TLSv1.3 (IN), TLS handshake, Newsession Ticket (4):
* using HTTP/2
* [HTTP/2] [1] OPENED stream for https://localhost:8443/
* [HTTP/2] [1] [:method: GET]
* [HTTP/2] [1] [:scheme: https]
* [HTTP/2] [1] [:authority: localhost:8443]
* [HTTP/2] [1] [:path: /]
* [HTTP/2] [1] [user-agent: curl/8.6.0]
* [HTTP/2] [1] [accept: */*]
> GET / HTTP/2
> Host: localhost:8443
> User-Agent: curl/8.6.0
> Accept: */*
>
< HTTP/2 200
< alt-svc: h3=":8443"; ma=2592000
< content-type: text/plain; charset=utf-8
< server: Caddy
< content-length: 13
< date: Fri, 03 Jul 2024 12:00:00 GMT
<
* Connection #0 to host localhost left intact
Hello, World!
```
