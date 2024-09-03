# aimg

ai images viewer

1. 从微信公众号链接中下载图片, 例如: `aimg https://mp.weixin.qq.com/s/sXBUmFpedslBoysUCBVcEw https://mp.weixin.qq.com/s/0-MD2r1yZ_QAxPRNojAn-A`
2. 导入本地图片夹子或者压缩包, 例如: `aimg /Users/bingoo/Downloads/xxx  /Users/bingoo/Downloads/yyy.zip /Users/bingoo/Downloads/zzz.7z`
3. 从 wallhaven.cc 随机下载, 例如： `aimg -page 1 https://wallhaven.cc/randoma`
4. 从 pixabay.com 下载, 例如：`aimg -page 1 https://pixabay.com/users/elf-moondance-19728901/`

查看：

1. 随机查看10张图片： `http://127.0.0.1:1100`
2. 随机查看20张图片： `http://127.0.0.1:1100/?n=20`
3. 查看约100K的图片： `http://127.0.0.1:1100/s/500K/?n=20`

![](snapshots/2023-12-14-21-30-04.png)


注意：

1. 直接响应图片，只在于 /x/{xxhash}，或者 ?n=1 两种情况
