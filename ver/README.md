# ver

1. 提供统一的 go 工程构建的 Makefile
2. 提供统一的版本号信息

| \# | 变量名称       | 变量含义                                             |
|----|------------|--------------------------------------------------|
| 1  | GitCommit  | Git 提交信息                                         |
| 2  | BuildTime  | 构建时间                                             |
| 3  | BuildHost  | 构建机器名称                                           |
| 4  | BuildIP    | 构建机器 IP                                          |
| 5  | BuildCI    | 构建的 CI 信息，e.g. Javis_V1.0.0_BuiltID_20240813.131 |
| 6  | GoVersion  | 构建 go版本号                                         |
| 7  | AppVersion | 应用版本号                                            |

