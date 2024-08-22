# rt

runtime 相关

1. `pf, err := rt.StartMemProf()` 开始记录 mem profile
2. `pf, err := rt.StartCPUProf()` 开始记录 cpu profile
3. `ExpandHome` 展开 ~ 主目录
4. `WriteTempFile` 写临时文件
5. `Exists` 文件是否存在
6. `Close` 关闭多个 io.Closer 
7. `ReadAll` 从 io.Reader 中读取为字符串
8. `ExpandAtFile` 扩展 @file 字符串，读取文件内容
