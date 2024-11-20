package ipfsKeystoneTest

// #cgo LDFLAGS: -L/usr/local/ipfs-keystone -lipfs_keystone -lstdc++
// #cgo CFLAGS: -I/usr/local/ipfs-keystone/include -I/usr/local/ipfs-keystone/include/host -I/usr/local/ipfs-keystone/include/edge
// #include "ipfs_keystone.h"
import "C"

import (
    "fmt"
    "unsafe"
    "runtime"
    "sync"
)

func Ipfs_keystone_test(isAES int, FileName string) {

    // 打印FileName
    fmt.Println("Processing file:", FileName)

    // Convert Go int to C int
    cIsAES := C.int(isAES)
    // Convert Go string to C char
    cFileName := C.CString(FileName)

    defer C.free(unsafe.Pointer(cFileName))

	// 使用goroutine启动C函数
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
        defer runtime.KeepAlive(nil) // Ensure the goroutine does not return before the C function completes

		C.ipfs_keystone(cIsAES, cFileName)
	}()

	// 等待所有goroutines完成
	wg.Wait()

}
