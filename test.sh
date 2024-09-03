#!/bin/bash

# 遍历当前目录下的所有子目录
for dir in */ ; do
    # 进入子目录
    cd "$dir"
    
    # 执行 go mod tidy 命令
    echo "Running 'go mod tidy' in $dir"
    rm -fr go.sum
    go get -u -t ./...
    go mod tidy
    
    # 检查命令是否执行成功
    if [ $? -ne 0 ]; then
        echo "Error: 'go mod tidy' failed in $dir"
        continue # 如果 go mod tidy 失败，则跳过 go test
    fi
    
    # 执行 go test 命令
    echo "Running 'go test' in $dir"
    go test ./...
    
    # 检查命令是否执行成功
    if [ $? -ne 0 ]; then
        echo "Error: 'go test' failed in $dir"
    fi
    
    # 返回到原始目录
    cd ..
done

echo "All directories processed."