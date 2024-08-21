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
