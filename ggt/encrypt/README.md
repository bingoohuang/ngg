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

```sh
$ ggt sm2 key --dir .
2024-09-25 22:41:50.844 [INFO ] 65201 --- [1     ] [-] : key file sm2_pri.pem created!
2024-09-25 22:41:50.845 [INFO ] 65201 --- [1     ] [-] : key file sm2_pub.pem created!

$ ggt sm2 sign -i bingoohuang -k sm2_pri.pem
MEQCIEq/b+KzaC6jzUM/HI1oRfKZec2Wq+mW4xKY4E49aH1PAiAabAbA+0AA7hfPfuC4NHY6B/9q3F7yGMGp+wM5KHAq3A==
                                                                                                                           
$ ggt sm2 verify -i bingoohuang -k sm2_pub.pem --sign MEQCIEq/b+KzaC6jzUM/HI1oRfKZec2Wq+mW4xKY4E49aH1PAiAabAbA+0AA7hfPfuC4NHY6B/9q3F7yGMGp+wM5KHAq3A==
true
```

## sm2 公钥加密，私钥解密

```sh
$ ggt sm2 encrypt -i bingoohuang -k sm2_pub.pem 
2024-09-25 22:58:37.031 [INFO ] 67989 --- [1     ] [-] : encrypted: BGK9tMqqVwPjGMKhQKPMFSJFCKrTbOLphcShXtfoEQ+0Yf5hUvu7hzmUIny7nF8gBX2bA8Dv7/iBqqEkPBfW/onrSMPZMVt/dLrT1e6KEmo6j11JPQvUVA8D6fkk110IvbalHbI322eFuG/b

$ ggt sm2 decrypt -i BGK9tMqqVwPjGMKhQKPMFSJFCKrTbOLphcShXtfoEQ+0Yf5hUvu7hzmUIny7nF8gBX2bA8Dv7/iBqqEkPBfW/onrSMPZMVt/dLrT1e6KEmo6j11JPQvUVA8D6fkk110IvbalHbI322eFuG/b:base64 -k sm2_pri.pem 
2024-09-25 22:58:56.607 [INFO ] 68048 --- [1     ] [-] : decrypted: bingoohuang
```