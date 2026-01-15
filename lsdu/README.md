# lsdu (Hybrid of ls, du and df)

`lsdu` 是一个结合了 `ls`、`du` 和 `df` 功能的命令行工具，能够以 Markdown 表格的形式展示文件/目录的大小、占用的磁盘情况（总空间和剩余空间）、修改时间等信息。 fork from [lsdu](https://github.com/ddmls/lsdu)

## 使用方法

```bash
lsdu [OPTION]... [FILE|DIRECTORY]...
```

## 参数说明

- `-s, --sort`: 按大小排序（默认开启）。
- `-d, --dir`: 只列出目录本身，而不是其内容。
- `--si`: 使用 SI 单位（1000 作为基数，而不是 1024）。
- `-h, --help`: 显示帮助信息。

## 输出示例

程序会输出一个包含以下列的表格：

- `#`: 序号。
- `Mode`: 文件权限模式。
- `Apparent`: 表面大小（文件内容的实际长度）。
- `Allocated`: 实际占用的磁盘空间。
- `FreeSpace`: 文件所在分区的剩余空间。
- `TotalSpace`: 文件所在分区的总空间。
- `ModTime`: 最后修改时间。
- `Path`: 文件或目录路径。

最后还会汇总所有列出的文件/目录的总大小。

```sh
# lsdu -d /root/bytefs_root_root/test_1k_files_cipher
| # | Mode       | Apparent           | Allocated          | FreeSpace              | TotalSpace             | ModTime             | Path                                        |
| - | ---------- | ------------------ | ------------------ | ---------------------- | ---------------------- | ------------------- | ------------------------------------------- |
| 1 | drwxr-xr-x | 102 MiB(107401497) | 391 MiB(409604096) | 3.2 TiB(3530643628032) | 3.6 TiB(3948930732032) | 2026-01-13 14:26:01 | /root/bytefs_root_root/test_1k_files_cipher |
```
