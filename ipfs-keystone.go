package ipfsKeystoneTest

// #cgo LDFLAGS: -L/usr/local/ipfs-keystone -lipfs_keystone -lstdc++
// #cgo CFLAGS: -I/usr/local/ipfs-keystone/include -I/usr/local/ipfs-keystone/include/host -I/usr/local/ipfs-keystone/include/edge
// #include <stdlib.h>
// #include "ipfs_keystone.h"
// #include "ipfs_aes.h"
import "C"

import (
	"fmt"
	"io"
	"unsafe"
	"sync"
//	"time"
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

func NewTEEFileReaderDe(isAES int, FileName string) (*TEEFileReader, error) {

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
		C.ipfs_keystone_de(cIsAES, unsafe.Pointer(C.CString(FileName)), unsafe.Pointer(rb))
		fmt.Println("TEE read file done")
	}()

	return reader, nil
}

func Ipfs_keystone_test_de(isAES int, FileName string) (TEEFileReader){

	// 打印FileName
	fmt.Println("Get file:", FileName)

	reader, _ := NewTEEFileReaderDe(isAES, FileName)
	// defer reader.Close()

	// var ior io.ReadCloser = reader

	return *reader
}

// Write 实现io.Write接口的方法，从p切片读取数据到缓冲区
func (r *TEEFileReader) Write(p []byte) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return 0, io.EOF
	}

	var wrsult C.int = 0;
	wrsult = C.ring_buffer_write((*C.RingBuffer)(r.rb), (*C.char)(unsafe.Pointer(&p[0])), C.size_t(len(p)))
	if wrsult == 0 { // 检查ring_buffer_write的结果
		return int(wrsult), io.EOF
	}
	return int(wrsult), nil
}

// Close 关闭TEEFileReader实例，释放相关资源
func (r *TEEFileReader) WaClose() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.closed {
		r.closed = true
		C.ring_buffer_stop((*C.RingBuffer)(r.rb));
		// C.free(unsafe.Pointer(r.rb))  // 由c语言程序释放内存
		close(r.readCh)  // 确保通道被关闭
		C.ring_buffer_already_got()
		// time.Sleep(1500 * time.Millisecond)  // 固定等待1.5s
		// r.wg.Wait()  // 等待后台goroutine完成
	}
	fmt.Println("TEEFileReader WaClose")
	return nil
}

// ==================================================================================
//				AES Encrypt
// ==================================================================================

func Rv_AES_Encrypt(pt []byte, ptLen int, ct []byte)(int) {

	ctLen := C.encrypt(unsafe.Pointer(&pt[0]), C.int(ptLen), unsafe.Pointer(&ct[0]))
	// ctLen := C.encrypt(unsafe.Pointer(uintptr(unsafe.Pointer(&pt[0]))), C.int(ptLen), unsafe.Pointer(uintptr(unsafe.Pointer(&ct[0]))))
	// ctLen := C.encrypt((*C.void)(unsafe.Pointer(&pt[0])), C.int(ptLen), (*C.void)(unsafe.Pointer(&ct[0])))
	return int(ctLen)

}

func Rv_AES_Decrypt(ct []byte, ctLen int, pt []byte)(int) {

	ptLen := C.decrypt(unsafe.Pointer(&ct[0]), C.int(ctLen), unsafe.Pointer(&pt[0]))
	// ptLen := C.decrypt(unsafe.Pointer(uintptr(unsafe.Pointer(&ct[0]))), C.int(ctLen), unsafe.Pointer(uintptr(unsafe.Pointer(&pt[0]))))
	// ptLen := C.decrypt((*C.void)(unsafe.Pointer(&ct[0])), C.int(ctLen), (*C.void)(unsafe.Pointer(&pt[0])))
	return int(ptLen)

}

// ==================================================================================
//				MultiThreaded Keystone Encrypt
// ==================================================================================


// TEEFileReader 结构体封装了环形缓冲区的相关操作
type MultiThreadedTEEFileReader struct {
	mtb     *C.MultiThreadedRingBuffer	// 指向C语言中的MultiThreadedRingBuffer结构
	readCh chan struct{}			// 通道用于通知读取完成
	wg     sync.WaitGroup			// 等待组用于等待后台goroutine完成
	mu     sync.Mutex			// 互斥锁，保护共享资源
	closed bool				// 标记是否已经关闭
}


// MultiThreadedTEEFileReader 创建一个新的MultiThreadedTEEFileReader实例
func NewMultiThreadedTEEFileReader(isAES int, FileName string, fileSize int) (*MultiThreadedTEEFileReader, error) {
	mtb := (*C.MultiThreadedTEEFileReader)(C.malloc(C.sizeof_MultiThreadedTEEFileReader))
	if mtb == nil { // 检查内存分配是否成功
		return nil, fmt.Errorf("failed to allocate memory for RingBuffer")
	}

	// Convert Go int to C int
	cIsAES := C.int(isAES)
	cFileSize := C.int(fileSize)

	cFileSize = C.alignedFileSize(cFileSize)
	cAfileSize := C.aFileSize(cFileSize)

	// 为 half part buffer 分配空间，设置两个buffer的运行状态都为running = 1
	C.init_multi_threaded_ring_buffer(mtb, cFileSize, cAfileSize)

	reader := &MultiThreadedTEEFileReader{
		mtb:    mtb,
		readCh: make(chan struct{}, 1),
		closed: false,
	}

	reader.wg.Add(1)
	go func() {
		defer reader.wg.Done() // 确保在goroutine结束时调用Done
		C.multi_ipfs_keystone_ppb_buffer_wrapper(cIsAES, unsafe.Pointer(C.CString(FileName)), unsafe.Pointer(mtb), 0, cAfileSize)
		fmt.Println("TEE ring buffer read file done")
	}()

	reader.wg.Add(1)
	go func() {
		defer reader.wg.Done() // 确保在goroutine结束时调用Done
		C.multi_ipfs_keystone_hpb_buffer_wrapper(cIsAES, unsafe.Pointer(C.CString(FileName)), unsafe.Pointer(mtb), cAfileSize, cFileSize)
		fmt.Println("TEE ring buffer read file done")
	}()

	return reader, nil
}

func MultiThreaded_Ipfs_keystone_test(isAES int, FileName string, fileSize int) (MultiThreadedTEEFileReader){

	// 打印FileName
	fmt.Println("MultiThread Processing file:", FileName)

	reader, _ := NewMultiThreadedTEEFileReader(isAES, FileName, fileSize)


	return *reader
}

func (mtb *MultiThreadedTEEFileReader)Read(p []byte) (int, error)  {
	mtb.mu.Lock()
	defer mtb.mu.Unlock()

	if r.closed {
		return 0, io.EOF
	}

	var readLen C.int = 0;
	result := C.which_pb_buffer_read((*C.MultiThreadedTEEFileReader)(unsafe.Pointer(mtb)), (*C.char)(unsafe.Pointer(&p[0])), C.int(len(p)), &readLen)
	if result == 0 { // 检查ring_buffer_read的结果
		return int(readLen), io.EOF
	}
	return int(readLen), nil
}

