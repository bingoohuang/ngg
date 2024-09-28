# rsa 加解密、签名验签

最佳实践密钥安全：

- 私钥必须妥善保存，绝不能以明文形式存储在代码仓库中。可以使用加密密钥库（如 AWS KMS、HashiCorp Vault 等）来保护密钥。
- OAEP Padding：在 RSA 加密中，始终使用 OAEP（Optimal Asymmetric Encryption Padding）以抵御多种攻击。
- 密钥长度：使用至少 2048 位的密钥长度，建议使用 4096 位以增强安全性。
- 错误处理：在加密和解密过程中，注意检查并处理可能出现的错误，避免信息泄露。
- 避免直接加密大数据：RSA 不适合直接加密大量数据。通常应将数据先用对称加密算法（如 AES）加密，再用 RSA 加密对称密钥。

```sh
$ ggt rsa -h 
Usage:
  ggt rsa [command]

Available Commands:
  check-keypair 检测私钥公钥是否匹配 / Check KeyPair
  encrypt       公钥加密/私钥解密 / Encrypt with public key, 或者  私钥加密/公钥解密 / Encrypt with private key
  newkey        生成公钥私钥
  sign          签名
  verify        验签

Flags:
  -h, --help   help for rsa
```

## 生成公钥私钥

```sh
$ ggt rsa newkey --dir .
2024-09-24 09:01:50.542 [INFO ] 24075 --- [1     ] [-] : key file rsa_pri.pem created!
2024-09-24 09:01:50.545 [INFO ] 24075 --- [1     ] [-] : key file rsa_pub.pem created!
```

## 签名验签

```sh
$ ggt rsa sign -k rsa_pri.pem -i bingoohuang
2024-09-24 09:02:37.175 [INFO ] 24156 --- [1     ] [-] : signed: t9YeI71Og9AdCWz2y0yIApBNi9hnbH2iJrK237iLlDiFziFUDrl5asJ5nlyyup4tZweWO7IgxmLB6RoJIEOEaAnheuj2lmdq7oazi8KSy8c2zMbzZl+bX9L40L633nDdKt328gExmh4AyNGUY0EtWiERwbsZvHOxr3d8tNviyAF6pyRxf4aQ91KP0qzTNJ4iH+M/CMje9tXUnUAyTHqQ88J7M6tJEB+XtKRN9M2rX9UPC2f7XvDB362CY7JsSyicFG5FQIltl4n9nmx3z8eH1LWRu5zXtHNaG+VTF1QQ3RHpHuRSjkPnNu8BAfxOk3GmLSw7wgNAKGQgl+mw/B0cIA==

$ ggt rsa verify -k rsa_pub.pem -i bingoohuang --sign t9YeI71Og9AdCWz2y0yIApBNi9hnbH2iJrK237iLlDiFziFUDrl5asJ5nlyyup4tZweWO7IgxmLB6RoJIEOEaAnheuj2lmdq7oazi8KSy8c2zMbzZl+bX9L40L633nDdKt328gExmh4AyNGUY0EtWiERwbsZvHOxr3d8tNviyAF6pyRxf4aQ91KP0qzTNJ4iH+M/CMje9tXUnUAyTHqQ88J7M6tJEB+XtKRN9M2rX9UPC2f7XvDB362CY7JsSyicFG5FQIltl4n9nmx3z8eH1LWRu5zXtHNaG+VTF1QQ3RHpHuRSjkPnNu8BAfxOk3GmLSw7wgNAKGQgl+mw/B0cIA==
2024-09-24 09:02:52.569 [INFO ] 24194 --- [1     ] [-] : verfied: true
```

## 公钥加密、私钥解密

```sh
$ ggt rsa encrypt -P rsa_pub.pem -i bingoohuang
2024-09-24 09:03:50.129 [INFO ] 24342 --- [1     ] [-] : encrypted: Kf+vNzg2NexqjAPWA0PCK+5U2aWewvoQ2KVaLAgcdrmkRDnnTnQfCyqsrWTM3Qu+lvTiBbnxCxuIho5DNpwAb6Y1662/t3gvXKloxKwkKpzVnFurYShTsSf5JsTc4YJ0k0toaHHmY1l5CamC25N/0YqkvBoPqzroCibwgZzSiUY7YmSW6FH1D0Roc6C8fu6pqyCUPLzKm2NVcrLofPQucaaQbQwEqUE2Uj1Xyt0lI4mEbfORdQRaeRSEaNIResjL56ZDnkzeGXrUfFO+SqMg2Ci8vd99ow24H9vkeYcl+djXIb5Nu/c4K0vQYwUdVn6Lv88BbR5NwOsRz+tKlf+ltA==

$ ggt rsa encrypt -p rsa_pri.pem -d -i Kf+vNzg2NexqjAPWA0PCK+5U2aWewvoQ2KVaLAgcdrmkRDnnTnQfCyqsrWTM3Qu+lvTiBbnxCxuIho5DNpwAb6Y1662/t3gvXKloxKwkKpzVnFurYShTsSf5JsTc4YJ0k0toaHHmY1l5CamC25N/0YqkvBoPqzroCibwgZzSiUY7YmSW6FH1D0Roc6C8fu6pqyCUPLzKm2NVcrLofPQucaaQbQwEqUE2Uj1Xyt0lI4mEbfORdQRaeRSEaNIResjL56ZDnkzeGXrUfFO+SqMg2Ci8vd99ow24H9vkeYcl+djXIb5Nu/c4K0vQYwUdVn6Lv88BbR5NwOsRz+tKlf+ltA==:b64
2024-09-24 09:04:30.340 [INFO ] 24419 --- [1     ] [-] : decrypted: bingoohuang
```

填充模式OAEP(Optimal Asymmetric Encryption Padding)

```sh
$ ggt rsa encrypt -P rsa_pub.pem -i bingoohuang --oaep
2024-09-24 09:09:49.649 [INFO ] 25256 --- [1     ] [-] : encrypted: oBPQRsodVebjcIv4Ibb10v+uKR9tiHxZRZF4ZZx5vtSOBsPZD0mgxCAemTtfcmHXMql9lfOkk5OYhDHYNeo5fU8FCquvMLLC97/C04gaRSpWQfTVwpgQjuprqclAlbMLgqaJs9GLPITIASurNuSrO3DDBbHCKIRQ/rUfTTuu4MfJRz3RonJbMZs1Jz5wCGiNGGyBIdWOT4Aqrs1fmSH+eGJbElpDh2FK8ziG0CeGuyyHPAX/uzgdzuWGfr8S2AORV0btJK0zP42bgcLCJvK04x/0uiBf5UIu0u/zldu+1PvRKnHhVJVPRtwLRwljaDoB1FugRy5iHAKQAHI/TlgZZg==

$ ggt rsa encrypt -p rsa_pri.pem --oaep -d -i oBPQRsodVebjcIv4Ibb10v+uKR9tiHxZRZF4ZZx5vtSOBsPZD0mgxCAemTtfcmHXMql9lfOkk5OYhDHYNeo5fU8FCquvMLLC97/C04gaRSpWQfTVwpgQjuprqclAlbMLgqaJs9GLPITIASurNuSrO3DDBbHCKIRQ/rUfTTuu4MfJRz3RonJbMZs1Jz5wCGiNGGyBIdWOT4Aqrs1fmSH+eGJbElpDh2FK8ziG0CeGuyyHPAX/uzgdzuWGfr8S2AORV0btJK0zP42bgcLCJvK04x/0uiBf5UIu0u/zldu+1PvRKnHhVJVPRtwLRwljaDoB1FugRy5iHAKQAHI/TlgZZg==:b64
2024-09-24 09:10:28.685 [INFO ] 25375 --- [1     ] [-] : decrypted: bingoohuang
```

## 私钥加密、公钥解密

```sh
$ ggt rsa encrypt -p rsa_pri.pem -i bingoohuang
2024-09-24 09:05:11.540 [INFO ] 24505 --- [1     ] [-] : encrypted: m9FNEytNQ2apeW/PUo2fVsYMU6fJEcCLCp30XivH2Yc1zRV63aLuswtP5zar3MrVdjKBU/o8HpKCOkZgDCrHYEFO8VQQp94uZx/MrrNcTNTI82MRpke2KfVv32ScUPOtVagFMyOEDmDdKACHnI551AbnJG0xHw5Rr0Xfy3m6q4C9KvAnJYFenNXer/Nzerxfzf+q2s7Zzu+z8MSDkTpSkhWt8d24RS2TCAQs4f6azfECIeQeORjB8ObAJ18uqRIce3lw5GTuQC/Yh/gNp3ck15OdEwCTBygcwWsaYERAutideA/2J7DdYHUrbXKYT8Sa1oEe2NHBS0WO1fBLB9FhQQ==
 
$ ggt rsa encrypt -P rsa_pub.pem -d -i  m9FNEytNQ2apeW/PUo2fVsYMU6fJEcCLCp30XivH2Yc1zRV63aLuswtP5zar3MrVdjKBU/o8HpKCOkZgDCrHYEFO8VQQp94uZx/MrrNcTNTI82MRpke2KfVv32ScUPOtVagFMyOEDmDdKACHnI551AbnJG0xHw5Rr0Xfy3m6q4C9KvAnJYFenNXer/Nzerxfzf+q2s7Zzu+z8MSDkTpSkhWt8d24RS2TCAQs4f6azfECIeQeORjB8ObAJ18uqRIce3lw5GTuQC/Yh/gNp3ck15OdEwCTBygcwWsaYERAutideA/2J7DdYHUrbXKYT8Sa1oEe2NHBS0WO1fBLB9FhQQ==:b64
2024-09-24 09:05:43.705 [INFO ] 24642 --- [1     ] [-] : decrypted: bingoohuang
```

## 检测私钥公钥是否匹配 / Check KeyPair

```sh
$ ggt rsa check-keypair -P rsa_pub.pem -p rsa_pri.pem           
2024-09-24 09:33:23.178 [INFO ] 26942 --- [1     ] [-] : check key pair result: true
```
