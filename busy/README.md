# busy

模拟内存消耗和CPU忙碌

```go
// ControlCPULoad run CPU load in specify cores count and percentage
//
//	coresCount: 使用核数
//	percentage: 每核 CPU 百分比 (默认 100), 0 时不开启 CPU 耗用
//	lockOsThread: 是否在 CPU 耗用时锁定 OS 线程
func ControlCPULoad(ctx context.Context, coresCount, percentage int, lockOsThread bool)

// ControlMem 控制内存消耗
func ControlMem(ctx context.Context, totalMem uint64) error
```

## 一键引入

```go
import (
    _ "github.com/bingoohuang/ngg/busy/autoload"
)
```

## Environment

| Name              | Default  | Meaning                              | Usage                         |
| ----------------- | -------- | ------------------------------------ | ----------------------------- |
| DOG_DEBUG         | 0        | Debug mode                           | `export DOG_DEBUG=1`          |
| DOG_DIR           | 当前目录 | 检查 Dog.busy 和生成 Dog.exit 的路径 | `export DOG_DIR=/etc/dog`     |
| DOG_BUSY_INTERVAL | 10s      | 检查 Dog.busy 文件的间隔时间         | `export DOG_BUSY_INTERVAL=1m` |


### 在程序的工作目录(环境变量 DOG_DIR 指定), 生成 Dog.busy 文件

- `echo '{"pprof": "15s"}' > Dog.busy` 在15秒后生成 cpu/mem.pprof 文件，取回本地， 执行命令 `go tool pprof -http=:8080 Dog.xxx.prof` 自动打开浏览器查看
- `echo '{"mem":"20MiB"}' > Dog.busy` 打满 20 MiB 内存 (用于模拟测试)
- `echo '{"cores":3,"cpu":100}' > Dog.busy` 打满 3 个核（用于模拟测试）
