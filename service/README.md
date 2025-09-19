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

### 一行代码集成

通过导入 `autoload` 包，您的程序将自动支持作为系统服务运行：

```go
import (
	_ "github.com/bingoohuang/ngg/service/autoload"           // 支持安装为系统服务
)
```

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
- **简单集成**: 只需一行 import 即可启用所有功能

### 安装路径

- **Linux/macOS/FreeBSD**: `/usr/local/{程序名}/`
- **Windows**: `C:\Program Files\{程序名}\`

### 示例

```go
package main

import (
	"log"
	"time"
	
	_ "github.com/bingoohuang/ngg/service/autoload"
)

func main() {
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
