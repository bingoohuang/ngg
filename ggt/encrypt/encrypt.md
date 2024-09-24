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
