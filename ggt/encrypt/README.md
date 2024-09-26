# 对称加解密

## 文件加解密

```sh
$ ggt encrypt -i /Users/bingoo/Downloads/1.jpg --out /Users/bingoo/Downloads/1.jpg.aes                                                                                        
2024-09-24 22:20:03.828 [INFO ] 42816 --- [1     ] [-] : rand --key 4149fd5f0e8ee10d52155c56984b1761:hex
2024-09-24 22:20:03.829 [INFO ] 42816 --- [1     ] [-] : rand --iv 4781d2c35e619aaba39211ed31af4b63:hex
2024-09-24 22:20:03.831 [INFO ] 42816 --- [1     ] [-] : AES/GCM/NoPadding Encrypt result written to file /Users/bingoo/Downloads/1.jpg.aes

$ ggt encrypt -d --key 4149fd5f0e8ee10d52155c56984b1761:hex --iv 4781d2c35e619aaba39211ed31af4b63:hex -i /Users/bingoo/Downloads/1.jpg.aes -o /Users/bingoo/Downloads/2.jpg
2024-09-24 22:20:52.599 [INFO ] 42912 --- [1     ] [-] : AES/GCM/NoPadding Decrypt result written to file /Users/bingoo/Downloads/2.jpg

$ diff -s /Users/bingoo/Downloads/1.jpg /Users/bingoo/Downloads/2.jpg                                                                                                     
Files /Users/bingoo/Downloads/1.jpg and /Users/bingoo/Downloads/2.jpg are identical
```

## 文本加解密

```sh
$ ggt encrypt -i bingoohuang --base64
2024-09-24 22:19:04.780 [INFO ] 42690 --- [1     ] [-] : rand --key f4b1b49188227518f96e2e8c9214d9e4:hex
2024-09-24 22:19:04.783 [INFO ] 42690 --- [1     ] [-] : rand --iv fed6e50a238ebed2821e7abd4df94f51:hex
2024-09-24 22:19:04.783 [INFO ] 42690 --- [1     ] [-] : AES/GCM/NoPadding Encrypt result: KMiSPR7DE5j127FZOm9SctyIi9QhNn/Kx3N3

$ ggt encrypt -d --key f4b1b49188227518f96e2e8c9214d9e4:hex --iv fed6e50a238ebed2821e7abd4df94f51:hex -i KMiSPR7DE5j127FZOm9SctyIi9QhNn/Kx3N3:b64
2024-09-24 22:19:38.747 [INFO ] 42773 --- [1     ] [-] : AES/GCM/NoPadding Decrypt result: bingoohuang
```

## 文本 sm4 加解密

```sh
$ ggt encrypt --sm4 -i bingoohuang --base64                                                                                                                     
2024-09-24 22:24:23.202 [INFO ] 43995 --- [1     ] [-] : rand --key 2b49c80e2d1a47b18775aeccebb64ee4:hex
2024-09-24 22:24:23.205 [INFO ] 43995 --- [1     ] [-] : rand --iv e010af29b4aaae3e94a58615e04ab473:hex
2024-09-24 22:24:23.205 [INFO ] 43995 --- [1     ] [-] : SM4/GCM/NoPadding Encrypt result: x5Gf1/dhXTVdVL5wLsH/EihIyFpxTPI7lGSZ

$ ggt encrypt -d --sm4 --key 2b49c80e2d1a47b18775aeccebb64ee4:hex --iv e010af29b4aaae3e94a58615e04ab473:hex -i x5Gf1/dhXTVdVL5wLsH/EihIyFpxTPI7lGSZ:b64
2024-09-24 22:24:49.821 [INFO ] 44073 --- [1     ] [-] : SM4/GCM/NoPadding Decrypt result: bingoohuang
```

## sm2 签名验签

SM2 签名时，默认的 Hash 就是 SM3

```sh
$ ggt sm2 key --dir .
2024-09-26 08:14:45.433 [INFO ] 32783 --- [1     ] [-] : private key: 4CPliYoiw4/uo/7mJuV74OnrT90omaHJY6uu44UwrZo=
2024-09-26 08:14:45.435 [INFO ] 32783 --- [1     ] [-] : public key: BGN03rpdKTmGRTDmq9m0kXr+sVJB8k237zsCRmcN2pBFRd/2/7CDXnV19KuSttmgu/33BEjP66mL/TDag/QqltU=
2024-09-26 08:14:45.436 [INFO ] 32783 --- [1     ] [-] : key file sm2_pri.pem created!
2024-09-26 08:14:45.436 [INFO ] 32783 --- [1     ] [-] : key file sm2_pub.pem created!

$ ggt sm2 sign -i bingoohuang -k sm2_pri.pem
MEQCIEq/b+KzaC6jzUM/HI1oRfKZec2Wq+mW4xKY4E49aH1PAiAabAbA+0AA7hfPfuC4NHY6B/9q3F7yGMGp+wM5KHAq3A==
                                                                                                                           
$ ggt sm2 verify -i bingoohuang -K sm2_pub.pem --sign MEQCIEq/b+KzaC6jzUM/HI1oRfKZec2Wq+mW4xKY4E49aH1PAiAabAbA+0AA7hfPfuC4NHY6B/9q3F7yGMGp+wM5KHAq3A==
true
```


## sm2 公钥加密，私钥解密

```sh
$ ggt sm2 encrypt -i bingoohuang -K sm2_pub.pem 
2024-09-25 22:58:37.031 [INFO ] 67989 --- [1     ] [-] : encrypted: BGK9tMqqVwPjGMKhQKPMFSJFCKrTbOLphcShXtfoEQ+0Yf5hUvu7hzmUIny7nF8gBX2bA8Dv7/iBqqEkPBfW/onrSMPZMVt/dLrT1e6KEmo6j11JPQvUVA8D6fkk110IvbalHbI322eFuG/b

$ ggt sm2 decrypt -i -k sm2_pri.pem  BGK9tMqqVwPjGMKhQKPMFSJFCKrTbOLphcShXtfoEQ+0Yf5hUvu7hzmUIny7nF8gBX2bA8Dv7/iBqqEkPBfW/onrSMPZMVt/dLrT1e6KEmo6j11JPQvUVA8D6fkk110IvbalHbI322eFuG/b:base64
2024-09-25 22:58:56.607 [INFO ] 68048 --- [1     ] [-] : decrypted: bingoohuang
```

```sh
$ ggt sm2 encrypt -i bingoohuang -K BGN03rpdKTmGRTDmq9m0kXr+sVJB8k237zsCRmcN2pBFRd/2/7CDXnV19KuSttmgu/33BEjP66mL/TDag/QqltU=
2024-09-26 08:26:03.658 [INFO ] 34078 --- [1     ] [-] : encrypted: BIVb+XrLg+qHZLlo/2wCm/zVBOuvaUwMd1AE9XJTiiDvIkLJl8eAMX/8IrwcUvTwVO+30oY1hN/mGjB55Cnb5x6hAutwKEvZfzVfkw3V1Jgi5X2pi7/jJnm1y1iRnjfpZmh/OLBeGZNB1gid

$ ggt sm2 decrypt -k 4CPliYoiw4/uo/7mJuV74OnrT90omaHJY6uu44UwrZo= -i BIVb+XrLg+qHZLlo/2wCm/zVBOuvaUwMd1AE9XJTiiDvIkLJl8eAMX/8IrwcUvTwVO+30oY1hN/mGjB55Cnb5x6hAutwKEvZfzVfkw3V1Jgi5X2pi7/jJnm1y1iRnjfpZmh/OLBeGZNB1gid:base64
2024-09-26 08:26:52.685 [INFO ] 34179 --- [1     ] [-] : decrypted: bingoohuang
```

## 私钥公钥证书解析 / Parse PrivateKey or PublicKey/获取 x, y, d 16进制数据

```sh
$ ggt sm2 inspect -k sm2_pri.pem               
2024-09-26 08:28:26.332 [INFO ] 34322 --- [1     ] [-] : private key D data: e023e5898a22c38feea3fee626e57be0e9eb4fdd2899a1c963abaee38530ad9a
2024-09-26 08:28:26.333 [INFO ] 34322 --- [1     ] [-] : private key X data: 6374deba5d2939864530e6abd9b4917afeb15241f24db7ef3b0246670dda9045
2024-09-26 08:28:26.333 [INFO ] 34322 --- [1     ] [-] : private key Y data: 45dff6ffb0835e7575f4ab92b6d9a0bbfdf70448cfeba98bfd30da83f42a96d5

$ ggt sm2 inspect -K sm2_pub.pem 
2024-09-26 08:28:35.162 [INFO ] 34411 --- [1     ] [-] : public key X data: 6374deba5d2939864530e6abd9b4917afeb15241f24db7ef3b0246670dda9045
2024-09-26 08:28:35.163 [INFO ] 34411 --- [1     ] [-] : public key Y data: 45dff6ffb0835e7575f4ab92b6d9a0bbfdf70448cfeba98bfd30da83f42a96d5


$ ggt sm2 inspect -k 4CPliYoiw4/uo/7mJuV74OnrT90omaHJY6uu44UwrZo=           
2024-09-26 08:29:08.735 [INFO ] 34529 --- [1     ] [-] : private key D data: e023e5898a22c38feea3fee626e57be0e9eb4fdd2899a1c963abaee38530ad9a
2024-09-26 08:29:08.736 [INFO ] 34529 --- [1     ] [-] : private key X data: 6374deba5d2939864530e6abd9b4917afeb15241f24db7ef3b0246670dda9045
2024-09-26 08:29:08.736 [INFO ] 34529 --- [1     ] [-] : private key Y data: 45dff6ffb0835e7575f4ab92b6d9a0bbfdf70448cfeba98bfd30da83f42a96d5

$ ggt sm2 inspect -K BGN03rpdKTmGRTDmq9m0kXr+sVJB8k237zsCRmcN2pBFRd/2/7CDXnV19KuSttmgu/33BEjP66mL/TDag/QqltU=           
2024-09-26 08:29:21.859 [INFO ] 34611 --- [1     ] [-] : public key X data: 6374deba5d2939864530e6abd9b4917afeb15241f24db7ef3b0246670dda9045
2024-09-26 08:29:21.861 [INFO ] 34611 --- [1     ] [-] : public key Y data: 45dff6ffb0835e7575f4ab92b6d9a0bbfdf70448cfeba98bfd30da83f42a96d5


$ ggt sm2 inspect -K sm2_pub.pem -k sm2_pri.pem --check_pair
2024-09-26 08:30:16.559 [INFO ] 34816 --- [1     ] [-] : pair checked: true
```

## SM2 用 x, y 生成公钥，用 d 生成私钥 / use x,y to make public key and use d to make private key

```sh
$ ggt sm2 recover -x 6374deba5d2939864530e6abd9b4917afeb15241f24db7ef3b0246670dda9045 -y 45dff6ffb0835e7575f4ab92b6d9a0bbfdf70448cfeba98bfd30da83f42a96d5
2024-09-26 08:33:17.181 [INFO ] 35285 --- [1     ] [-] : public key: BGN03rpdKTmGRTDmq9m0kXr+sVJB8k237zsCRmcN2pBFRd/2/7CDXnV19KuSttmgu/33BEjP66mL/TDag/QqltU=
2024-09-26 08:33:17.184 [INFO ] 35285 --- [1     ] [-] : key:
-----BEGIN PUBLIC KEY-----
MFkwEwYHKoZIzj0CAQYIKoEcz1UBgi0DQgAEY3Teul0pOYZFMOar2bSRev6xUkHy
TbfvOwJGZw3akEVF3/b/sINedXX0q5K22aC7/fcESM/rqYv9MNqD9CqW1Q==
-----END PUBLIC KEY-----

$ ggt sm2 recover -x 6374deba5d2939864530e6abd9b4917afeb15241f24db7ef3b0246670dda9045 -y 45dff6ffb0835e7575f4ab92b6d9a0bbfdf70448cfeba98bfd30da83f42a96d5 -d e023e5898a22c38feea3fee626e57be0e9eb4fdd2899a1c963abaee38530ad9a
c38feea3fee626e57be0e9eb4fdd2899a1c963abaee38530ad9a
2024-09-26 08:33:57.890 [INFO ] 35345 --- [1     ] [-] : public key: BGN03rpdKTmGRTDmq9m0kXr+sVJB8k237zsCRmcN2pBFRd/2/7CDXnV19KuSttmgu/33BEjP66mL/TDag/QqltU=
2024-09-26 08:33:57.892 [INFO ] 35345 --- [1     ] [-] : key:
-----BEGIN PUBLIC KEY-----
MFkwEwYHKoZIzj0CAQYIKoEcz1UBgi0DQgAEY3Teul0pOYZFMOar2bSRev6xUkHy
TbfvOwJGZw3akEVF3/b/sINedXX0q5K22aC7/fcESM/rqYv9MNqD9CqW1Q==
-----END PUBLIC KEY-----
2024-09-26 08:33:57.892 [INFO ] 35345 --- [1     ] [-] : private key: 4CPliYoiw4/uo/7mJuV74OnrT90omaHJY6uu44UwrZo=
2024-09-26 08:33:57.892 [INFO ] 35345 --- [1     ] [-] : key:
-----BEGIN PRIVATE KEY-----
MIGTAgEAMBMGByqGSM49AgEGCCqBHM9VAYItBHkwdwIBAQQg4CPliYoiw4/uo/7m
JuV74OnrT90omaHJY6uu44UwrZqgCgYIKoEcz1UBgi2hRANCAARjdN66XSk5hkUw
5qvZtJF6/rFSQfJNt+87AkZnDdqQRUXf9v+wg151dfSrkrbZoLv99wRIz+upi/0w
2oP0KpbV
-----END PRIVATE KEY-----

```

## 私钥证书编码格式转换 / Change PrivateKey type

```sh
$ ggt sm2 convert -k sm2_pri.pem --pkcs8
2024-09-26 08:35:51.165 [INFO ] 35588 --- [1     ] [-] : key:
-----BEGIN PRIVATE KEY-----
MIGTAgEAMBMGByqGSM49AgEGCCqBHM9VAYItBHkwdwIBAQQg4CPliYoiw4/uo/7m
JuV74OnrT90omaHJY6uu44UwrZqgCgYIKoEcz1UBgi2hRANCAARjdN66XSk5hkUw
5qvZtJF6/rFSQfJNt+87AkZnDdqQRUXf9v+wg151dfSrkrbZoLv99wRIz+upi/0w
2oP0KpbV
-----END PRIVATE KEY-----
```
