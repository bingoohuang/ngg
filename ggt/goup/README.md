# goup

Utility to upload or download large files through HTTP in broken-point continuously.

This is done by splitting large file into small chunks; whenever the upload of a chunk fails, uploading is retried until
the procedure completes. This allows uploads to automatically resume uploading after a network connection is lost, or
program restarting either locally or to the server.

## Features

1. splitting large file into smaller chunks.
2. uploading/downloading by HTTP.
3. resume-able. chunk's hash will be checked before transfer.
4. security data by AES-GCM based on [PAKE](https://github.com/schollz/pake).
5. support download short path like `goup -path /xx=/xx.zip`, then the client can use `http://127.0.0.1:2001/xx` to
   download the xx.zip file.

| API | Method | Req Content-Gulp         | Rsp Content-Gulp | Other Headers                                          | Function                                           |
|----:|:-------|:-------------------------|------------------|:-------------------------------------------------------|:---------------------------------------------------|
|  1. | POST   | Filename                 |                  |                                                        | 明文上传（文件作为 Body)                                    |
|  2. | POST   | Session, Curve           | Curve,           |                                                        | PAKE 生成会话秘钥                                        |
|  3. | GET /  | Session, Range, Checksum |                  | Req: Content-Disposition                               | 校验分块 checksum，返回 304 或 其它                          |
|  4. | POST / | Session, Range, Salt     |                  | Req: Content-Disposition                               | 分块加密上传（加密分块作为 Body)                                |
|  5. | GET /  |                          |                  |                                                        | HTML JS 上传页面 / 服务端文件列表（Accept: application/json 时） |
|  6. | GET /  | Session, Range, Checksum | Range, Salt      | Rsp: Content-Type , Content-Disposition                | 分块加密下载                                             |
|  7. | GET /  |                          | Salt             | Rsp: Content-Type, Content-Length, Content-Disposition | 明文下载                                               |
|  8. | POST   |                          |                  |                                                        | 明文上传（multipart-form)                               |

![](_doc/img.png)

## Usage

```shell
$ goup -h
Usage of goup:
  -b    string Bearer token for client or server, auto for server to generate a random one
  -c    string Chunk size for client, unit MB (default 10)
  -t    int    Threads (go-routines) for client
  -f    string Upload file path for client
  -p    int    Listening port for server
  -r    string Rename to another filename
  -u    string Server upload url for client to connect to
  -P    string Password for PAKE
  -C    string Cipher AES256: AES-256 GCM, C20P1305: ChaCha20 Poly1305
  -v    bool   Show version
```

1. Installation `go install https://github.com/bingoohuang/ngg/goup`
2. At the server, `goup`
3. At the client, `goup -u http://a.b.c:2110/ -f 246.png`

```sh
$ goup
2022/01/20 17:40:20 Listening on 2110
2022/01/20 17:40:24 recieved file 246.png with sessionID CF79FC6EA88010E1, range bytes 0-10485760/96894303
2022/01/20 17:40:24 recieved file 246.png with sessionID CF79FC6EA88010E1, range bytes 10485760-20971520/96894303
2022/01/20 17:40:24 recieved file 246.png with sessionID CF79FC6EA88010E1, range bytes 20971520-31457280/96894303
2022/01/20 17:40:24 recieved file 246.png with sessionID CF79FC6EA88010E1, range bytes 31457280-41943040/96894303
2022/01/20 17:40:24 recieved file 246.png with sessionID CF79FC6EA88010E1, range bytes 41943040-52428800/96894303
2022/01/20 17:40:47 recieved file 246.png with sessionID 46CB3E789B10DDB5, range bytes 52428800-62914560/96894303
2022/01/20 17:40:47 recieved file 246.png with sessionID 46CB3E789B10DDB5, range bytes 62914560-73400320/96894303
2022/01/20 17:40:47 recieved file 246.png with sessionID 46CB3E789B10DDB5, range bytes 73400320-83886080/96894303
2022/01/20 17:40:47 recieved file 246.png with sessionID 46CB3E789B10DDB5, range bytes 83886080-94371840/96894303
2022/01/20 17:40:47 recieved file 246.png with sessionID 46CB3E789B10DDB5, range bytes 94371840-96894303/96894303

$ md5sum .goup-files/246.png
5849c383ab841e539673d50a49c3f39a  .goup-files/aDrive.dmg
$ sha256sum .goup-files/246.png
32677ca3a5b19ee5590a75d423b3fc2d754bd76900d50c26e0525681985189f8  .goup-files/aDrive.dmg
```

upload example:

```sh
$ goup -u http://127.0.0.1:2110/ -f 246.png
2022/01/20 17:44:08 Upload B821C49E4B1CBDBD started
40.00 MiB / 92.41 MiB [------------------>____________________] 43.29% ? p/s^C
$ diff -s /Users/bingoobjca/Downloads/goup/.goup-files/246.png 246.png
Binary files /Users/bingoobjca/Downloads/goup/.goup-files/246.png and 246.png differ
$ goup -u http://127.0.0.1:2110/ -f 246.png
2022/01/20 17:44:21 Upload 4739FE47D18ADF7F started
92.41 MiB / 92.41 MiB [---------------------------------------] 100.00% 121.54 MiB p/s
2022/01/20 17:44:21 Upload 4739FE47D18ADF7F completed
$ diff -s /Users/bingoobjca/Downloads/goup/.goup-files/246.png 246.png
Files /Users/bingoobjca/Downloads/goup/.goup-files/246.png and 246.png are identical
$ md5 246.png
MD5 (aDrive.dmg) = 5849c383ab841e539673d50a49c3f39a
$ sha256sum 246.png
32677ca3a5b19ee5590a75d423b3fc2d754bd76900d50c26e0525681985189f8  246.png
```

As you can see, even if you terminate the client uploading, you can resume uploading from the last breakpoint.

```sh
$ goup -u http://127.0.0.1:2110/246.png
2022/01/21 13:29:43 download started: .goup-files/246.png
92.41 MiB / 92.41 MiB [---------------------------------------] 100.00% 97.56 MiB p/s
2022/01/21 13:29:44 download complete: .goup-files/246.png
```

## normal upload/download support

upload:

1. Start the server without a chunk: `goup -c0`
2. Test the upload: `curl -F "file:=@a.mp4" localhost:2110`

download:

1. Start the server: `goup`
2. Download
    1. by [gurl](https://github.com/bingoohuang/gurl): `gurl :2110/246.png`
    2. by curl: `curl -LO 127.0.0.1:2110/246.png`
    3. by goup disable chunk: `goup -u :2110/246.png -c0`

## Resources

1. [SRP](https://github.com/posterity/srp) Go implementation of the Secure Remote Password (SRP) protocol.
   > SRP is an authentication method that allows the use of user names and passwords over unencrypted channels without
   revealing the password to an eavesdropper.
   > SRP also supplies a shared secret at the end of the authentication sequence that can be used to generate encryption
   keys.
2. [Encrypted-Content-Encoding HTTP](https://github.com/posterity/ece) Go implementation of Encrypted-Content-Encoding
   for HTTP (RFC 8188).
