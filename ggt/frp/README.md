# proxytarget

为 web 访问，启动临时 frp 代理，例如

`proxytarget -p 2000 -t https://172.16.24.101 -p :6001`

访问 http://127.0.0.1:2000 将经过代理 127.0.0.1:6001 指向 https://172.16.24.101

示例配置文件 


```yaml
---
proxyaddr: :6000

proxies:
  - listen: :3003
    targetaddr:  192.168.1.51:8090
    desc: wiki
  - listen: :3002
    targetaddr: 192.168.1.1:80
    desc: gitlab
  - listen: :3001
    targetaddr: http://172.16.2.5:31010
    desc: rigaga
  - listen: :3004
    targetaddr: https://172.16.2.1
    desc: 虚拟机控制台
```
