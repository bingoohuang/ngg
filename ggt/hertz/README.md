# herz

example:

1. `ggt hertz -p 'GET /hello' -b '{"State":200}'`
2. `ggt hertz -p 'GET /world' -b @a.json`
3. `ggt hertz -p /upload -u /Users/bingoo/aaa/upload` 开启上传服务
   1. 测试上传 `gurl -F /Users/bingoo/aaa/2.jpeg :12123/upload`
