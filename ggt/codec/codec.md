# codec

```sh
$ ggt codec -h                             
hash, baseXx, and etc.

Usage:
  ggt codec [flags]

Flags:
  -f, --from enum        from. allowed: string,hex,base32,base45,base58,base62,base64,base85,base91,base100,safeURL.
  -h, --help             help for codec
  -i, --input string     Input string, or filename
  -k, --key string       HMAC key. env: $KEY.
  -r, --raw              print raw bytes
  -t, --to stringArray   to. allowed: string,hex,base32,base45,base58,base62,base64,base85,base91,base100,safeURL,md2,md4,md5,sha1,sha3-224,sha3-256,sha3-384,sha3-512,sha224,sha256,sha384,sha512,sha512-224,sha512-256,shake128-256,shake128-512,shake256,ripemd160,blake2b-256,blake2b-384,blake2b-512,blake2s-256,blake3,sm3,xxhash.
```

```sh
$ ggt codec -i bingoohuang -t sm3                 
2024-09-29 08:30:44.691 [INFO ] 7061 --- [1     ] [-] : sm3 hex: 1ab21d8355cfa17f8e61194831e81a8f22bec8c728fefb747ed035eb5082aa2b (len: 64)
2024-09-29 08:30:44.692 [INFO ] 7061 --- [1     ] [-] : sm3 base64: GrIdg1XPoX+OYRlIMegajyK+yMco/vt0ftA161CCqis= (len: 44)
```

```sh
$ ggt codec -i bingoohuang -t base64           
2024-09-29 08:32:22.581 [INFO ] 7383 --- [1     ] [-] : base64 raw: YmluZ29vaHVhbmc= (len: 16)
```

```sh
$ ggt codec -i bingoohuang -t base100
2024-09-29 08:32:36.498 [INFO ] 7424 --- [1     ] [-] : base100 raw: 👙👠👥👞👦👦👟👬👘👥👞 (len: 44)
```

```sh
$  ggt codec -i /Users/bingoo/Downloads/1.jpeg -t xxhash
2024-09-29 09:12:23.325 [INFO ] 15201 --- [1     ] [-] : sum64: 6855997601766409421, hex: 5f25657e823784cd
2024-09-29 09:12:23.328 [INFO ] 15201 --- [1     ] [-] : xxhash raw: _%e~�7�� (len: 8)
2024-09-29 09:12:23.328 [INFO ] 15201 --- [1     ] [-] : xxhash hex: 5f25657e823784cd (len: 16)
2024-09-29 09:12:23.328 [INFO ] 15201 --- [1     ] [-] : xxhash base64: XyVlfoI3hM0= (len: 12)
```
