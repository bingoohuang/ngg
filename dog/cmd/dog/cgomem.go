//go:build cgomem
// +build cgomem

package main

/*
#include <stdlib.h>
#include <stdio.h>
#include <string.h>
#include <time.h>

// C 函数：分配内存并用随机值填充
void *allocate_and_fill_random(size_t size) {
    void *mem = malloc(size);
    if (mem == NULL) {
        perror("Memory allocation failed");
        exit(EXIT_FAILURE);
    }

    // 初始化随机数生成器
    srand((unsigned int)time(NULL));

    // 填充随机值
	size_t i = 0;
    for (; i < size; ++i) {
        ((unsigned char *)mem)[i] = rand() % 256;
    }
    return mem;
}

// C 函数：释放内存
void free_memory(void *mem) {
    free(mem);
}
*/
import "C"
import (
	"flag"
	"log"

	"github.com/bingoohuang/ngg/ss"
)

func cgoMemDemo(cgoMemSize uint64) {
	if mem := C.allocate_and_fill_random(C.size_t(cgoMemSize)); mem == nil {
		log.Fatalf("Failed to allocate memory")
	}
}

var cgoMem = flag.String("cgo-mem", "", "cgo malloc memory size")

func cgoDemo() {
	if *cgoMem == "" {
		return
	}

	if cgoMemSize, err := ss.ParseBytes(*cgoMem); err != nil {
		log.Fatalf("parse cgoMem %s error: %v", *cgoMem, err)
	} else if cgoMemSize > 0 {
		cgoMemDemo(cgoMemSize)
	}
}
