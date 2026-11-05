# service

fork from [service](https://github.com/kardianos/service)

service will install / un-install, start / stop, and run a program as a service (daemon).
Currently supports Windows XP+, Linux/(systemd | Upstart | SysV), and OSX/Launchd.

Windows controls services by setting up callbacks that is non-trivial. This
is very different then other systems. This package provides the same API
despite the substantial differences.
It also can be used to detect how a program is called, from an interactive
terminal or from a service manager.

## BUGS

 * Dependencies field is not implemented for Linux systems and Launchd.
 * OS X when running as a UserService Interactive will not be accurate.

## Usage

### 集成

导入 `autoload` 包，并在 `main()` 开头调用 `autoload.Run()`：

```go
import (
	"github.com/bingoohuang/ngg/service/autoload"
)
```

```go
func main() {
    autoload.Run()
    // 业务逻辑
}
```

Windows 必须调用 `autoload.Run()`，不能仅使用空白导入。
Go 的原生 Windows 回调需要等待包初始化完成，因此服务调度器不能阻塞在 `init()` 中。
正常工作进程中 `Run()` 会直接返回；daemon 进程中它会阻塞到服务停止。
Linux/macOS 保留原有空白导入兼容行为，也可以采用以上统一写法。

### 使用方式

#### 1. 安装为系统服务
```bash
SRV=install your-program
# 或者
SRV=i your-program
```

#### 2. 卸载系统服务
```bash
SRV=uninstall your-program
# 或者
SRV=u your-program
```

#### 3. 以守护进程模式运行
```bash
SRV=daemon your-program
# 或者
SRV=d your-program
```

### 功能特性

- **自动安装**: 自动将程序安装为系统服务，支持 Windows、Linux、macOS
- **守护进程**: 支持守护进程模式，自动重启工作进程
- **日志管理**: 自动创建日志目录和日志文件
- **跨平台**: 支持 Windows XP+、Linux (systemd/Upstart/SysV)、macOS (Launchd)
- **简单集成**: 导入 `autoload` 并在 `main()` 开头调用 `autoload.Run()`

### 安装路径

- **Linux/macOS/FreeBSD**: `/usr/local/{程序名}/`
- **Windows**: `C:\Program Files\{程序名}\`

Windows 安装后的可执行文件保留 `.exe` 后缀，例如
`C:\Program Files\fost\fost.exe`；服务名仍为 `fost`。
安装器会为服务持久化 `SRV=daemon`，无需手工修改注册表或在业务参数中追加 daemon 参数。
工作进程会移除 `SRV` 环境变量，避免再次进入服务模式。

工作进程以可执行文件所在目录为工作目录，自动创建该目录下的 `log` 文件夹。
标准输出和标准错误写入 `log/dump.log.tmp`，进程退出后转存为 `log/dump.log`。
安装、卸载或服务运行失败时，命令以非零状态码退出；已停止的服务也可以重新安装。

### 示例

```go
package main

import (
	"log"
	"time"
	
	"github.com/bingoohuang/ngg/service/autoload"
)

func main() {
	autoload.Run()
	// 您的业务逻辑
	for {
		log.Println("程序正在运行...")
		time.Sleep(5 * time.Second)
	}
}
```

编译后运行：
```bash
# 安装为系统服务
SRV=install ./your-program

# 程序将自动安装并启动为系统服务
```
