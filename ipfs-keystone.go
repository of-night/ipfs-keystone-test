package ipfsKeystoneTest

// #cgo LDFLAGS: -L/usr/local/ipfs-keystone -lipfs_keystone -lstdc++
// #cgo CFLAGS: -I/usr/local/ipfs-keystone/include -I/usr/local/ipfs-keystone/include/host -I/usr/local/ipfs-keystone/include/edge
// #include <stdlib.h>
// #include "ipfs_keystone.h"
import "C"

import (
	"fmt"
	"io"
	"unsafe"
	"sync"
)

// TEEFileReader 结构体封装了环形缓冲区的相关操作
type TEEFileReader struct {
	rb     *C.RingBuffer          // 指向C语言中的RingBuffer结构
	readCh chan struct{}          // 通道用于通知读取完成
	wg     sync.WaitGroup         // 等待组用于等待后台goroutine完成
	mu     sync.Mutex             // 互斥锁，保护共享资源
	closed bool                   // 标记是否已经关闭
}

// NewTEEFileReader 创建一个新的TEEFileReader实例
func NewTEEFileReader(isAES int, FileName string) (*TEEFileReader, error) {
	rb := (*C.RingBuffer)(C.malloc(C.sizeof_RingBuffer))
	if rb == nil { // 检查内存分配是否成功
		return nil, fmt.Errorf("failed to allocate memory for RingBuffer")
	}

	// Convert Go int to C int
	cIsAES := C.int(isAES)

	C.init_ring_buffer(rb)

	reader := &TEEFileReader{
		rb:     rb,
		readCh: make(chan struct{}, 1),
		closed: false,
	}

	reader.wg.Add(1)
	go func() {
		defer reader.wg.Done() // 确保在goroutine结束时调用Done
		C.ipfs_keystone(cIsAES, unsafe.Pointer(C.CString(FileName)), unsafe.Pointer(rb))
		fmt.Println("TEE read file done")
	}()

	return reader, nil
}

// Read 实现io.Reader接口的方法，从缓冲区读取数据到p切片
func (r *TEEFileReader) Read(p []byte) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return 0, io.EOF
	}

	var readLen C.int = 0;
	result := C.ring_buffer_read((*C.RingBuffer)(r.rb), (*C.char)(unsafe.Pointer(&p[0])), C.int(len(p)), &readLen)
	if result == 0 { // 检查ring_buffer_read的结果
		return int(readLen), io.EOF
	}
	return int(readLen), nil
}

// Close 关闭TEEFileReader实例，释放相关资源
func (r *TEEFileReader) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.closed {
		r.closed = true
		C.free(unsafe.Pointer(r.rb))  // 释放C语言分配的内存
		close(r.readCh)  // 确保通道被关闭
		// r.wg.Wait()  // 等待后台goroutine完成
	}
	fmt.Println("TEEFileReader Close")
	return nil
}

func Ipfs_keystone_test(isAES int, FileName string) (TEEFileReader){

	// 打印FileName
	fmt.Println("Processing file:", FileName)

	reader, _ := NewTEEFileReader(isAES, FileName)
	// defer reader.Close()

	// var ior io.ReadCloser = reader

	return *reader
}

