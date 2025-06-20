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
	// "time"
	"bytes"

	"os"
	"os/exec"
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
	mtb     *C.MultiThreadedBuffer  	// 指向C语言中的MultiThreadedBuffer结构
	readCh chan struct{}          		// 通道用于通知读取完成
	wg     sync.WaitGroup         		// 等待组用于等待后台goroutine完成
	mu     sync.Mutex             		// 互斥锁，保护共享资源
	closed bool                   		// 标记是否已经关闭
}


// NewMultiThreadedTEEFileReader 创建一个新的MultiThreadedTEEFileReader实例
func NewMultiThreadedTEEFileReader(isAES int, FileName string, fileSize int) (*MultiThreadedTEEFileReader, error) {
	mtb := (*C.MultiThreadedBuffer)(C.malloc(C.sizeof_MultiThreadedBuffer))
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
		fmt.Println("MultiTEE buffer read file done")
	}()

	reader.wg.Add(1)
	go func() {
		defer reader.wg.Done() // 确保在goroutine结束时调用Done
		C.multi_ipfs_keystone_hpb_buffer_wrapper(cIsAES, unsafe.Pointer(C.CString(FileName)), unsafe.Pointer(mtb), cAfileSize + 1, cFileSize)
		fmt.Println("MultiTEE ring buffer read file done")
	}()

	return reader, nil
}

func MultiThreaded_Ipfs_keystone_test(isAES int, FileName string, fileSize int) (MultiThreadedTEEFileReader){

	// 打印FileName
	fmt.Println("MultiThread Processing file:", FileName)

	reader, _ := NewMultiThreadedTEEFileReader(isAES, FileName, fileSize)


	return *reader
}

func (mtbr *MultiThreadedTEEFileReader)Read(p []byte) (int, error)  {
	mtbr.mu.Lock()
	defer mtbr.mu.Unlock()

	if mtbr.closed {
		return 0, io.EOF
	}

	var readLen C.int = 0;
	result := C.which_pb_buffer_read((*C.MultiThreadedBuffer)(unsafe.Pointer(mtbr.mtb)), (*C.char)(unsafe.Pointer(&p[0])), C.int(len(p)), &readLen)
	if result == 0 { // 检查ring_buffer_read的结果
		return int(readLen), io.EOF
	}
	return int(readLen), nil
}

// Close 关闭TMultiThreadedTEEFileReader实例，释放相关资源
func (mtbr *MultiThreadedTEEFileReader) Close() error {
	mtbr.mu.Lock()
	defer mtbr.mu.Unlock()

	if !mtbr.closed {
		mtbr.closed = true
		C.destory_multi_threaded_ring_buffer((*C.MultiThreadedBuffer)(unsafe.Pointer(mtbr.mtb)))
		// C.free(unsafe.Pointer(mtbr.mtb))  // 释放C语言分配的内存
		close(mtbr.readCh)  // 确保通道被关闭
		// mtbr.wg.Wait()  // 等待后台goroutine完成
	}
	fmt.Println("TEEFileReader Close")
	return nil
}



// ==================================================================================
//				Multi-process Keystone Encrypt
// ==================================================================================

const (
	shmKey   = 241227 // 共享内存键值
)

type MultiProcessTEEFileReader struct {
	shmaddr     []byte				  	// 共享内存的地址
	shmsize     int				  		// 共享内存的长度
	mpb		*C.MultiProcessSHMBuffer	// 指向C语言中的 MultiProcessSHMBuffer 结构
	readCh chan struct{}          		// 通道用于通知读取完成
	mu     sync.Mutex             		// 互斥锁，保护共享资源
	closed bool                   		// 标记是否已经关闭
}

// 创建一个新的共享内存段
func createShm(size int) ([]byte, error) {

	shmaddr := C.creat_shareMemory(C.int(size))

	// 错误写法 (*[size]byte)中 size 必须为常量，只是类型转换，并没有分配空间
	// return (*[size]byte)(shmaddr)[:], nil
	// [low:high:max] 获取内存切片low-high 可以索引low-high  数组实际空间大小为max
	// 若不指定 max 则是前面类型的空间，即1 << 32 = 1GB
	return (*[1 << 32]byte)(shmaddr)[:size:size], nil
}

// 连接到现有的共享内存段
func attachShm(size int) ([]byte, error) {

	shmaddr := C.attach_shareMemory(C.int(size))

	// 错误写法 (*[size]byte)中 size 必须为常量，只是类型转换，并没有分配空间
	// return (*[size]byte)(shmaddr)[:], nil
	// [low:high:max] 获取内存切片low-high 可以索引low-high  数组实际空间大小为max
	// 若不指定 max 则是前面类型的空间，即1 << 32 = 1GB
	return (*[1 << 32]byte)(shmaddr)[:size:size], nil
}

// 断开与共享内存段的连接
func detachShm(shm []byte) error {
	C.detach_shareMemory(unsafe.Pointer(&shm[0]))

	return nil
}

// 删除共享内存段
func removeShm(shmsize int) error {
	C.removeShm(C.int(shmsize))

	return nil
}

// NewMultiProcessTEEFileReader MultiProcessTEEFileReader
func NewMultiProcessTEEFileReader(isAES int, FileName string, fileSize int) (*MultiProcessTEEFileReader, error) {

	// 创建共享内存片段
	shmsize := fileSize + C.sizeof_MultiProcessSHMBuffer
	shm, err := createShm(shmsize)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create shared memory: %v\n", err)
		os.Exit(1)
	}

	// Convert Go int to C int
	cFileSize := C.int(fileSize)

	cFileSize = C.alignedFileSize(cFileSize)
	cAfileSize := C.aFileSize(cFileSize)

	reader := &MultiProcessTEEFileReader{
		shmaddr:    shm,
		shmsize:	shmsize,
		readCh: make(chan struct{}, 1),
		closed: false,
	}

	// 启动第一个子进程，读取文件的前半部分
	cmd1 := exec.Command("./child_process", fmt.Sprintf("%d", isAES), fmt.Sprintf("%d", shmsize), FileName, fmt.Sprintf("%d", 0), fmt.Sprintf("%d", cAfileSize))

	// 创建缓冲区来存储标准输出和标准错误
	var stdout1, stderr1 bytes.Buffer

	// 将子进程的标准输出和标准错误重定向到上面创建的缓冲区
	cmd1.Stdout = &stdout1
	cmd1.Stderr = &stderr1

	err = cmd1.Start()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start first child process: %v\n", err)
		os.Exit(1)
	}

	// 启动第二个子进程，读取文件的后半部分
	cmd2 := exec.Command("./child_process", fmt.Sprintf("%d", isAES), fmt.Sprintf("%d", shmsize), FileName, fmt.Sprintf("%d", cAfileSize), fmt.Sprintf("%d", cFileSize))

	// 创建缓冲区来存储标准输出和标准错误
    var stdout2, stderr2 bytes.Buffer

    // 将子进程的标准输出和标准错误重定向到上面创建的缓冲区
    cmd2.Stdout = &stdout2
    cmd2.Stderr = &stderr2

	err = cmd2.Start()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start second child process: %v\n", err)
		os.Exit(1)
	}

	// time.Sleep(1500 * time.Millisecond)

	// // 打印子进程的标准输出
    //     fmt.Printf("Child process output:\n%s\n", stdout1.String())
    //     fmt.Printf("Child process output:\n%s\n", stdout2.String())

	// cmd1.Wait()
	// cmd2.Wait()

	// // 打印子进程的标准输出
    //     fmt.Printf("Child process output:\n%s\n", stdout1.String())
    //     fmt.Printf("Child process output:\n%s\n", stdout2.String())

	// // 打印子进程的输出
    //     fmt.Printf("Child process output:\n%s\n", stderr1.String())
    //     fmt.Printf("Child process output:\n%s\n", stderr2.String())


	C.waitKeystoneReady(unsafe.Pointer(&reader.shmaddr[0]))

	fmt.Printf("Child1 process output:\n%s\n", stdout1.String())
	fmt.Printf("Child2 process output:\n%s\n", stdout2.String())

	return reader, nil
}


func MultiProcess_Ipfs_keystone_test(isAES int, FileName string, fileSize int) (MultiProcessTEEFileReader){

	// 打印FileName
	fmt.Println("MultiProcess Processing file:", FileName)

	reader, _ := NewMultiProcessTEEFileReader(isAES, FileName, fileSize)


	return *reader
}


// Close 关闭MultiProcessTEEFileReader实例，释放相关资源
func (mptr *MultiProcessTEEFileReader) Close() error {
	mptr.mu.Lock()
	defer mptr.mu.Unlock()

	if !mptr.closed {
		mptr.closed = true
		close(mptr.readCh)  // 确保通道被关闭
		defer detachShm(mptr.shmaddr)
		defer removeShm(mptr.shmsize)
	}
	fmt.Println("MultiProcess TEEFileReader Close")
	return nil
}

// 从共享内存中读取数据并打印出来
func parentProcess(shmaddr []byte, shmsize int, p []byte, size int)(int, error) {

	var readLen C.int = 0;
	// 交给c语言函数处理
	result := C.MultiProcessRead(unsafe.Pointer(&shmaddr[0]), C.int(shmsize), unsafe.Pointer(&p[0]), C.int(size), &readLen);

	if result == 0 {
		return int(readLen), io.EOF
	}

	return int(readLen), nil;
}

func (mtbr *MultiProcessTEEFileReader)Read(p []byte) (int, error)  {
	mtbr.mu.Lock()
	defer mtbr.mu.Unlock()

	if mtbr.closed {
		return 0, io.EOF
	}

	return parentProcess(mtbr.shmaddr, mtbr.shmsize, p, len(p))
}


// ==================================================================================
//				Multi-process Cross-read Keystone Encrypt
// ==================================================================================


type MultiProcessCrossTEEFileReader struct {
	shmaddr     []byte				  	// 共享内存的地址
	shmsize     int64				  		// 共享内存的长度
	readCh chan struct{}          		// 通道用于通知读取完成
	mu     sync.Mutex             		// 互斥锁，保护共享资源
	closed bool                   		// 标记是否已经关闭
}

// 创建一个新的共享内存段
func longcreateShm(size int64) ([]byte, error) {

	shmaddr := C.long_create_shareMemory(C.longlong(size))

	// 错误写法 (*[size]byte)中 size 必须为常量，只是类型转换，并没有分配空间
	// return (*[size]byte)(shmaddr)[:], nil
	// [low:high:max] 获取内存切片low-high 可以索引low-high  数组实际空间大小为max
	// 若不指定 max 则是前面类型的空间，即1 << 32 = 1GB
	return (*[1 << 32]byte)(shmaddr)[:size:size], nil
}

// 删除共享内存段
func longremoveShm(shmsize int64) error {
	C.long_removeShm(C.longlong(shmsize))

	return nil
}


// NewMultiProcessCrossTEEFileReader MultiProcessCrossTEEFileReader
func NewMultiProcessCrossTEEFileReader(isAES int, FileName string, fileSize int64) (*MultiProcessCrossTEEFileReader, error) {

	// Convert Go int to C int
	cFileSize := C.longlong(fileSize)

	cFileSize = C.long_alignedFileSize(cFileSize)
	cBlocksNums := C.long_alignedFileSize_blocksnums(cFileSize)
	
	// 创建共享内存片段
	shmsize := C.sizeof_MultiProcessCrossSHMBuffer + (int64(cBlocksNums) * 4) + int64(cFileSize)
	shm, err := longcreateShm(shmsize)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create shared memory: %v\n", err)
		os.Exit(1)
	}

	reader := &MultiProcessCrossTEEFileReader{
		shmaddr:    shm,
		shmsize:	shmsize,
		readCh: make(chan struct{}, 1),
		closed: false,
	}

	// 启动keystone之前先初始化内存空间
	C.crossInitSHM(unsafe.Pointer(&reader.shmaddr[0]), cBlocksNums);
	// fmt.Println("MultiProcess Processing file test")

	// 启动第一个子进程，读取文件的前半部分
	cmd1 := exec.Command("./cross_child_process", fmt.Sprintf("%d", isAES), fmt.Sprintf("%d", shmsize), FileName, fmt.Sprintf("%d", 0))

	// 创建缓冲区来存储标准输出和标准错误
	var stdout1, stderr1 bytes.Buffer

	// 将子进程的标准输出和标准错误重定向到上面创建的缓冲区
	cmd1.Stdout = &stdout1
	cmd1.Stderr = &stderr1

	err = cmd1.Start()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start first child process: %v\n", err)
		os.Exit(1)
	}

	// 启动第二个子进程，读取文件的后半部分
	cmd2 := exec.Command("./cross_child_process", fmt.Sprintf("%d", isAES), fmt.Sprintf("%d", shmsize), FileName, fmt.Sprintf("%d", 1))

	// 创建缓冲区来存储标准输出和标准错误
	var stdout2, stderr2 bytes.Buffer

	// 将子进程的标准输出和标准错误重定向到上面创建的缓冲区
	cmd2.Stdout = &stdout2
	cmd2.Stderr = &stderr2

	err = cmd2.Start()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start second child process: %v\n", err)
		os.Exit(1)
	}

	C.crosswaitKeystoneReady(unsafe.Pointer(&reader.shmaddr[0]))

	// fmt.Println("MultiProcess Processing file wait")

	// cmd1.Wait()
	// cmd2.Wait()

	// // 打印子进程的标准输出
    //     fmt.Printf("Child process output:\n%s\n", stdout1.String())
    //     fmt.Printf("Child process output:\n%s\n", stdout2.String())

	// // 打印子进程的标准错误输出
    //     fmt.Printf("Child process err output:\n%s\n", stderr1.String())
    //     fmt.Printf("Child process err output:\n%s\n", stderr2.String())

	return reader, nil
}


func MultiProcess_Cross_Ipfs_keystone_test(isAES int, FileName string, fileSize int64) (MultiProcessCrossTEEFileReader){

	// 打印FileName
	fmt.Println("MultiProcess Processing file:", FileName)

	reader, _ := NewMultiProcessCrossTEEFileReader(isAES, FileName, fileSize)


	return *reader
}

func (mpcr *MultiProcessCrossTEEFileReader)Read(p []byte) (int, error)  {
	mpcr.mu.Lock()
	defer mpcr.mu.Unlock()

	if mpcr.closed {
		return 0, io.EOF
	}

	// fmt.Println("MultiProcess Processing read start")
	var readLen C.int = 0;
	// 交给c语言函数处理
	result := C.MultiProcessCrossRead(unsafe.Pointer(&mpcr.shmaddr[0]), C.int(mpcr.shmsize), unsafe.Pointer(&p[0]), C.int(len(p)), &readLen);
	if result == 0 {
		return int(readLen), io.EOF
	}

	// fmt.Println("MultiProcess Processing read done")

	return int(readLen), nil;
}

// Close 关闭MultiProcessCrossTEEFileReader实例，释放相关资源
func (mpcr *MultiProcessCrossTEEFileReader) Close() error {
	mpcr.mu.Lock()
	defer mpcr.mu.Unlock()

	if !mpcr.closed {
		mpcr.closed = true
		close(mpcr.readCh)  // 确保通道被关闭
		defer detachShm(mpcr.shmaddr)
		defer longremoveShm(mpcr.shmsize)
	}
	fmt.Println("MultiProcess Cross TEEFileReader Close")
	return nil
}


// ==================================================================================
//				Multi-process Cross-read Flexible Keystone Encrypt
// ==================================================================================

type MultiProcessCrossTEEFileFlexibleReader struct {
	shmaddr     []byte				  	// 共享内存的地址
	shmsize     int64				  	// 共享内存的长度
	readCh chan struct{}          		// 通道用于通知读取完成
	mu     sync.Mutex             		// 互斥锁，保护共享资源
	closed bool                   		// 标记是否已经关闭
}


// NewMultiProcessCrossTEEFileFlexibleReader MultiProcessCrossTEEFileFlexibleReader
func NewMultiProcessCrossTEEFileFlexibleReader(isAES int, FileName string, fileSize int64, flexible int) (*MultiProcessCrossTEEFileFlexibleReader, error) {

	// Convert Go int to C int
	cFileSize := C.longlong(fileSize)

	cFileSize = C.long_alignedFileSize(cFileSize)
	cBlocksNums := C.long_alignedFileSize_blocksnums(cFileSize)
	
	// 创建共享内存片段
	shmsize := C.sizeof_MultiProcessCrossFlexibleSHMBuffer + (int64(cBlocksNums) * 4) + int64(cFileSize)
	shm, err := longcreateShm(shmsize)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create shared memory: %v\n", err)
		os.Exit(1)
	}

	reader := &MultiProcessCrossTEEFileFlexibleReader{
		shmaddr:    shm,
		shmsize:	shmsize,
		readCh: make(chan struct{}, 1),
		closed: false,
	}

	// 启动keystone之前先初始化内存空间
	C.flexiblecrossInitSHM(unsafe.Pointer(&reader.shmaddr[0]), cBlocksNums);
	// fmt.Println("MultiProcess Processing file test")

	var numflexible int = 0
	// MAXNUM 10
	C.fixFlexibleNum(unsafe.Pointer(&flexible))
	for numflexible < flexible {

		// 启动第一个子进程，读取文件的前半部分
		cmd := exec.Command("./flexible_cross_child_process", 
			fmt.Sprintf("%d", isAES), 
			fmt.Sprintf("%d", shmsize), 
			FileName, 
			fmt.Sprintf("%d", numflexible), 
			fmt.Sprintf("%d", flexible),
		)

		err = cmd.Start()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to start %d child process: %v\n", numflexible, err)
			os.Exit(1)
		}
		numflexible++

	}

	C.flexiblecrosswaitKeystoneReady(unsafe.Pointer(&reader.shmaddr[0]), C.int(flexible))

	return reader, nil
}


func MultiProcess_Cross_Flexible_Ipfs_keystone_test(isAES int, FileName string, fileSize int64, flexible int) (MultiProcessCrossTEEFileFlexibleReader){

	// 打印FileName
	fmt.Println("MultiProcess flexible Processing file:", FileName)

	reader, _ := NewMultiProcessCrossTEEFileFlexibleReader(isAES, FileName, fileSize, flexible)


	return *reader
}


func (mpcfr *MultiProcessCrossTEEFileFlexibleReader)Read(p []byte) (int, error)  {
	mpcfr.mu.Lock()
	defer mpcfr.mu.Unlock()

	if mpcfr.closed {
		return 0, io.EOF
	}

	// fmt.Println("MultiProcess Processing read start")
	var readLen C.int = 0;
	// 交给c语言函数处理
	result := C.MultiProcessCrossReadFlexible(unsafe.Pointer(&mpcfr.shmaddr[0]), C.int(mpcfr.shmsize), unsafe.Pointer(&p[0]), C.int(len(p)), &readLen);
	if result == 0 {
		return int(readLen), io.EOF
	}

	// fmt.Println("MultiProcess Processing read done")

	return int(readLen), nil;
}

// Close 关闭 MultiProcessCrossTEEFileFlexibleReader 实例，释放相关资源
func (mpcfr *MultiProcessCrossTEEFileFlexibleReader) Close() error {
	mpcfr.mu.Lock()
	defer mpcfr.mu.Unlock()

	if !mpcfr.closed {
		mpcfr.closed = true
		close(mpcfr.readCh)  // 确保通道被关闭
		defer detachShm(mpcfr.shmaddr)
		defer longremoveShm(mpcfr.shmsize)
	}
	fmt.Println("MultiProcess Cross TEEFileReader Close")
	return nil
}



// ==================================================================================
//				Multi-process Keystone Decrypt
// ==================================================================================

type Shmsm struct {
	shmaddr     []byte				  	// 共享内存的地址
	shmsize     int64				  	// 共享内存的长度
}

type MultiProcessTEEDispatch struct {
	shmsm	[]Shmsm
	blockcount int64
	blockbytes int64
	flexible int
	readCh	chan struct{}          		// 通道用于通知读取完成
	mu		sync.Mutex             		// 互斥锁，保护共享资源
	closed	bool                   		// 标记是否已经关闭
}

func DispathSetLength(size uint64) {
	C.dispathSetLength(C.ulonglong(size))
}

// 获取递增的engine_id函数
func GetDispathEngineSeq() (uint64) {
	return uint64(C.getDispathEngineSeq())
}

// 创建一个新的共享内存段
func dispath_longcreateShm(size int64, en_id int) ([]byte, error) {

	shmaddr := C.dispath_long_create_shareMemory(C.longlong(size), C.int(en_id))

	// 错误写法 (*[size]byte)中 size 必须为常量，只是类型转换，并没有分配空间
	// return (*[size]byte)(shmaddr)[:], nil
	// [low:high:max] 获取内存切片low-high 可以索引low-high  数组实际空间大小为max
	// 若不指定 max 则是前面类型的空间，即1 << 32 = 1GB
	return (*[1 << 32]byte)(shmaddr)[:size:size], nil
}

// 断开连接共享内存
func dispath_detachShm(shm []Shmsm, flexible int) error {

	for i := 0; i < flexible; i++ {
		C.dispath_detach_shareMemory(unsafe.Pointer(&shm[i].shmaddr[0]))
	}

	return nil
}

// 删除共享内存段
func dispath_longremoveShm(shm []Shmsm, flexible int) error {

	for i := 0; i < flexible; i++ {
		C.dispath_long_removeShm(C.longlong(shm[i].shmsize), C.int(i))
	}

	return nil
}

// NewMultiProcessTEEDispatch MultiProcessTEEDispatch
func NewMultiProcessTEEDispatch(isAES int, fileSize uint64, flexible int) (*MultiProcessTEEDispatch, error) {

	// MAXNUM <= 10
	C.fixFlexibleNum(unsafe.Pointer(&flexible))
	
	reader := &MultiProcessTEEDispatch{
		shmsm:  make([]Shmsm, flexible),
		blockcount: 0,
		blockbytes: 0,
		flexible: flexible,
		readCh: make(chan struct{}, 1),
		closed: false,
	}

	// // 需要调度的块总数,对256*1024向上取整
	// cBlocksNums := C.long_alignedFileSize_blocksnums(C.longlong(fileSize))

	// //每一个enclave最少需要接收的块数量和大小
	// eblock := (int64(cBlocksNums)/flexible)
	// esize := eblock * 256 * 1024

	// // 剩下的块数量
	// seblock := int64(cBlocksNums)%flexible

	var eblock int64;
	var seblock int64;
	C.dispath_blocks(C.ulonglong(fileSize), unsafe.Pointer(&eblock), unsafe.Pointer(&seblock), C.int(flexible));

	var shmsize int64;
	// 创建共享内存片段
	if (seblock == 0) {
		for i := 0; i < flexible; i++ {
			// 每个enclave的共享内存的大小，调度器与enclave之间
			if (i == (flexible - 1)) {
				if eblock == 0 {
					shmsize = C.sizeof_MultiProcessTEEDispatchSHMBuffer
				} else {
					shmsize = C.sizeof_MultiProcessTEEDispatchSHMBuffer + (eblock - 1)*(4+262144) + (int64)(4 + (fileSize & 0x3ffff))
				}
				
			} else {
				shmsize = C.sizeof_MultiProcessTEEDispatchSHMBuffer + eblock*(4+262144)
			}
			
			// 每一个enclave与dispath之间都有一个共享内存
			shm, err := dispath_longcreateShm(shmsize, i)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to create shared memory: %v\n", err)
				os.Exit(1)
			}
			reader.shmsm[i] = Shmsm{
				shmaddr: shm,
				shmsize: shmsize,
			}
			// 启动keystone之前先初始化内存空间
			C.dispath_InitSHM(unsafe.Pointer(&reader.shmsm[i].shmaddr[0]), C.longlong(eblock));
		}
	} else {
		for i := 0; i < flexible; i++ {
			var snumber int64;
			var snumber_size int64;
			if seblock > 1 {
				snumber = 1
				snumber_size = 4 + 262144
			} else if seblock == 1{
				snumber = 1
				snumber_size = (int64)(4 + (fileSize & 0x3ffff))
			} else {
				snumber = 0
				snumber_size = 0
			}
			seblock -= 1
			// 每个enclave的共享内存的大小，调度器与enclave之间
			// shmsize = C.sizeof_MultiProcessTEEDispatchSHMBuffer + (eblock+snumber)*(4+256*1024)
			shmsize = C.sizeof_MultiProcessTEEDispatchSHMBuffer + eblock*(4+262144) + snumber_size;
			// 每一个enclave与dispath之间都有一个共享内存
			shm, err := dispath_longcreateShm(shmsize, i)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to create shared memory: %v\n", err)
				os.Exit(1)
			}
			reader.shmsm[i] = Shmsm{
				shmaddr: shm,
				shmsize: shmsize,
			}
			
			// 启动keystone之前先初始化内存空间
			C.dispath_InitSHM(unsafe.Pointer(&reader.shmsm[i].shmaddr[0]), C.longlong(eblock+snumber));
		}
	}

	// fmt.Println("ipfs-keystone testing SHM")

	// 获取当前ms_group的 engine_id
	dispathEngineSeq := GetDispathEngineSeq()

	for numflexible:=0;numflexible<flexible;numflexible++ {
		// var stdout, stderr bytes.Buffer

		// 启动第一个子进程，读取文件的前半部分
		cmd := exec.Command("./dispath_child_process", 
			fmt.Sprintf("%d", isAES), 
			fmt.Sprintf("%d", shmsize), 
			fmt.Sprintf("%d", numflexible), 
			fmt.Sprintf("%d", flexible),
			fmt.Sprintf("%d", dispathEngineSeq),
		)

		// cmd.Stdout = &stdout
		// cmd.Stderr = &stderr

		err := cmd.Start()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to start %d child process: %v\n", numflexible, err)
			os.Exit(1)
		}

		// fmt.Printf("ipfs-keystone Child %d process output:\n%s\n", numflexible, stdout.String())
		// fmt.Printf("ipfs-keystone Child %d process err   :\n%s\n", numflexible, stderr.String())
	}

	
	// fmt.Printf("ipfs-keystone testing numberflexible Sleep 1500 ms\n")
	// time.Sleep(1500 * time.Millisecond)

	// fmt.Println("ipfs-keystone testing cmd Start")
	
	for i := 0; i < flexible; i++ {
		for {
			if (C.dispathwaitKeystoneReady(unsafe.Pointer(&reader.shmsm[i].shmaddr[0])) == 1) {
				break
			}
		}
	}

	fmt.Println("ipfs-keystone testing ready")

	return reader, nil
}

func MultiProcess_Dispath_Ipfs_keystone_test(isAES int, flexible int) (MultiProcessTEEDispatch){

	// 打印
	fmt.Println("MultiProcess dispath Processing...")

	// 获取总大小
	var fileSize uint64
	C.dispathGetLength((*C.ulonglong)(unsafe.Pointer(&fileSize)))

	reader, _ := NewMultiProcessTEEDispatch(isAES, fileSize, flexible)


	return *reader
}

// Write 实现io.Write接口的方法，从p切片读取数据到缓冲区
func (MPDispath *MultiProcessTEEDispatch) Write(p []byte) (int, error) {
	MPDispath.mu.Lock()
	defer MPDispath.mu.Unlock()

	if MPDispath.closed {
		return 0, io.EOF
	}

	var readLen C.int = 0;
	
	// fmt.Println("ipfs testing dispath 1 block blockcount=%d", MPDispath.blockcount)
	// fmt.Println("ipfs testing dispath 1 block blockbytes=%d", MPDispath.blockbytes)

	var sbytes int64 = int64(262144 - (MPDispath.blockbytes + int64(len(p))))

	var bnumber int64;
	
	bnumber = MPDispath.blockcount % int64(MPDispath.flexible)

	if sbytes >= 0 {
		result := C.dispath_data_block_4096(unsafe.Pointer(&MPDispath.shmsm[bnumber].shmaddr[0]), C.longlong(MPDispath.shmsm[bnumber].shmsize), (*C.char)(unsafe.Pointer(&p[0])), C.int(len(p)), &readLen)
		MPDispath.blockbytes = MPDispath.blockbytes+int64(len(p))
		// fmt.Println("ipfs testing dispath block only bnumber=%d, len=%d", bnumber, len(p))
		if sbytes == 0 {
			MPDispath.blockbytes = 0
			MPDispath.blockcount++
		}
		if result == 0 {
			return int(readLen), io.EOF
		}
	} else {
		var syx int = int(262144 - MPDispath.blockbytes)
		result := C.dispath_data_block_4096(unsafe.Pointer(&MPDispath.shmsm[bnumber].shmaddr[0]), C.longlong(MPDispath.shmsm[bnumber].shmsize), (*C.char)(unsafe.Pointer(&p[0])), C.int(syx), &readLen)
		// fmt.Println("ipfs testing dispath block oonly bnumber=%d, len=%d", bnumber, syx)
		if result == 0 {
			return int(readLen), io.EOF
		}
		MPDispath.blockcount++

		bnumber = MPDispath.blockcount % int64(MPDispath.flexible)
		var readLen1 C.int = 0;
		result = C.dispath_data_block_4096(unsafe.Pointer(&MPDispath.shmsm[bnumber].shmaddr[0]), C.longlong(MPDispath.shmsm[bnumber].shmsize), (*C.char)(unsafe.Pointer(&p[syx])), C.int(len(p) - syx), &readLen1)
		// fmt.Println("ipfs testing dispath block oonly bnumber=%d, len=%d", bnumber, len(p) - syx)

		MPDispath.blockbytes = int64(len(p) - syx)
		readLen = readLen + readLen1
		if result == 0 {
			return int(readLen), io.EOF
		}
	}

	return int(readLen), nil

	// // var bnumber int64 = int64(C.dispathBNumber((*C.longlong)(unsafe.Pointer(&MPDispath.blockcount)), C.int(MPDispath.flexible)))
	// var bnumber int64 = int64(C.dispathBNumber_4096((*C.longlong)(unsafe.Pointer(&MPDispath.blockbytes)), (*C.longlong)(unsafe.Pointer(&MPDispath.blockcount)), C.int(MPDispath.flexible), C.int(len(p))))
	// fmt.Println("ipfs testing dispath 1 block data %d", bnumber)
	// result := C.dispath_data_block(unsafe.Pointer(&MPDispath.shmsm[bnumber].shmaddr[0]), C.longlong(MPDispath.shmsm[bnumber].shmsize), (*C.char)(unsafe.Pointer(&p[0])), C.int(len(p)), &readLen)
	// fmt.Println("ipfs testing dispath 1 block data done bnumber=%d", bnumber)
	// fmt.Println("ipfs testing dispath 1 block data done blockcount=%d", MPDispath.blockcount)
	// fmt.Println("ipfs testing dispath 1 block data done result=%d", result)
	// if result == 0 {
	// 	return int(readLen), io.EOF
	// }
	// fmt.Println("ipfs testing dispath 1 block data done %d", MPDispath.blockcount)
	// return int(readLen), nil
}

// Close 关闭TEEFileReader实例，释放相关资源
func (MPDispath *MultiProcessTEEDispatch) Close() error {
	MPDispath.mu.Lock()
	defer MPDispath.mu.Unlock()

	if !MPDispath.closed {
		MPDispath.closed = true
		close(MPDispath.readCh)  // 确保通道被关闭
		// defer detachShm(MPDispath.shmaddr)

		// 等待 Keystone done
		fmt.Println("ipfs testing wait keystone done")
		for i := 0; i < MPDispath.flexible; i++ {
			for {
				if (C.dispathwaitKeystoneReady(unsafe.Pointer(&MPDispath.shmsm[i].shmaddr[0])) == 2) {
					break
				}
			}
		}
		defer dispath_detachShm(MPDispath.shmsm, MPDispath.flexible)
		defer dispath_longremoveShm(MPDispath.shmsm, MPDispath.flexible)
	}
	fmt.Println("TEEWriterDispath Close")
	return nil
}


// ==================================================================================
//				Multi-process Keystone Decrypt secure dispatch
// ==================================================================================

type MultiProcessTEESecureDispatch struct {
	shmaddr     []byte				  	// 共享内存的地址
	shmsize     uint64				  	// 共享内存的长度
	blockNum 	uint64
	blockcount int64
	blockbytes int64
	flexible int
	readCh chan struct{}          		// 通道用于通知读取完成
	mu     sync.Mutex             		// 互斥锁，保护共享资源
	closed bool                   		// 标记是否已经关闭
}

// 创建一个新的共享内存段
func secure_dispatch_ulonglongcreateShm(shmsize uint64) ([]byte, error) {

	shmaddr := C.secure_dispatch_ulnoglong_create_shareMemory(C.ulonglong(shmsize))

	// 错误写法 (*[size]byte)中 size 必须为常量，只是类型转换，并没有分配空间
	// return (*[size]byte)(shmaddr)[:], nil
	// [low:high:max] 获取内存切片low-high 可以索引low-high  数组实际空间大小为max
	// 若不指定 max 则是前面类型的空间，即1 << 32 = 1GB
	return (*[1 << 32]byte)(shmaddr)[:shmsize:shmsize], nil
}

// NewMultiProcessTEESecureDispatch MultiProcessTEESecureDispatch
func NewMultiProcessTEESecureDispatch(isAES int, fileSize uint64, flexible int) (*MultiProcessTEESecureDispatch, error) {

	// MAXNUM <= 10
	C.fixFlexibleNum(unsafe.Pointer(&flexible))

	// 创建共享内存片段
	var blockNum uint64
	shmsize := uint64(C.MultiProcessTEESecureDispatchGetSHMSize(C.ulonglong(fileSize), unsafe.Pointer(&blockNum), C.int(flexible)))
	shmaddr, err := secure_dispatch_ulonglongcreateShm(shmsize)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create shared memory: %v\n", err)
		os.Exit(1)
	}
	
	reader := &MultiProcessTEESecureDispatch{
		shmaddr:    shmaddr,
		shmsize:	shmsize,
		blockNum: 	blockNum,
		blockcount: 0,
		blockbytes: 0,
		flexible: 	flexible,
		readCh: 	make(chan struct{}, 1),
		closed: 	false,
	}

	C.secure_dispacth_initSHM(unsafe.Pointer(&reader.shmaddr[0]), C.ulonglong(blockNum), C.int(flexible));

	// fmt.Println("ipfs-keystone testing SHM")

	// 获取当前ms_group的 engine_id  
	// dispatch
	dispatchEngineSeq := GetDispathEngineSeq()

	for numflexible:=0;numflexible<flexible;numflexible++ {
		// var stdout, stderr bytes.Buffer

		// 启动第一个子进程，读取文件的前半部分
		cmd := exec.Command("./secure_dispatch_child_process", 
			fmt.Sprintf("%d", isAES), 
			fmt.Sprintf("%d", shmsize), 
			fmt.Sprintf("%d", numflexible), 
			fmt.Sprintf("%d", flexible),
			fmt.Sprintf("%d", dispatchEngineSeq),
		)

		// cmd.Stdout = &stdout
		// cmd.Stderr = &stderr

		err := cmd.Start()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to start %d child process: %v\n", numflexible, err)
			os.Exit(1)
		}

		// fmt.Printf("ipfs-keystone Child %d process output:\n%s\n", numflexible, stdout.String())
		// fmt.Printf("ipfs-keystone Child %d process err   :\n%s\n", numflexible, stderr.String())
	}

	
	// fmt.Printf("ipfs-keystone testing numberflexible Sleep 1500 ms\n")
	// time.Sleep(1500 * time.Millisecond)

	// fmt.Println("ipfs-keystone testing cmd Start")
	
	C.secure_dispatch_waitKeystoneReady(unsafe.Pointer(&reader.shmaddr[0]), C.int(flexible))

	fmt.Println("ipfs-keystone testing ready")

	return reader, nil
}

func MultiProcess_Secure_Dispatch_Ipfs_keystone_test(isAES int, flexible int) (MultiProcessTEESecureDispatch){

	// 打印
	fmt.Println("MultiProcess secure dispatch Processing...")

	// 获取总大小
	var fileSize uint64
	C.dispathGetLength((*C.ulonglong)(unsafe.Pointer(&fileSize)))

	reader, _ := NewMultiProcessTEESecureDispatch(isAES, fileSize, flexible)


	return *reader
}

// Write 实现io.Write接口的方法，从p切片读取数据到缓冲区
func (MPSecureDispath *MultiProcessTEESecureDispatch) Write(p []byte) (int, error) {
	MPSecureDispath.mu.Lock()
	defer MPSecureDispath.mu.Unlock()

	if MPSecureDispath.closed {
		return 0, io.EOF
	}

	var readLen C.int = 0;
	
	// fmt.Println("ipfs testing dispath 1 block blockcount=%d", MPSecureDispath.blockcount)
	// fmt.Println("ipfs testing dispath 1 block blockbytes=%d", MPSecureDispath.blockbytes)

	result := C.secure_dispatch_write(unsafe.Pointer(&MPSecureDispath.shmaddr[0]), C.longlong(MPSecureDispath.shmsize), (*C.char)(unsafe.Pointer(&p[0])), C.int(len(p)), &readLen, C.int(MPSecureDispath.flexible))

	if result == 0 {
		return int(readLen), io.EOF
	}

	return int(readLen), nil
}

// Close 关闭TEEFileReader实例，释放相关资源
func (MPSecureDispath *MultiProcessTEESecureDispatch) Close() error {
	MPSecureDispath.mu.Lock()
	defer MPSecureDispath.mu.Unlock()

	if !MPSecureDispath.closed {
		MPSecureDispath.closed = true
		close(MPSecureDispath.readCh)  // 确保通道被关闭
		// defer detachShm(MPSecureDispath.shmaddr)

		// 等待 Keystone done
		fmt.Println("ipfs testing wait keystone done")
		C.secure_dispatch_waitKeystoneDone(unsafe.Pointer(&MPSecureDispath.shmaddr[0]), C.int(MPSecureDispath.flexible))

		// 断开连接共享内存
		defer C.secure_dispatch_detach_shareMemory(unsafe.Pointer(&MPSecureDispath.shmaddr[0]))
		// 删除共享内存段
		defer C.secure_dispatch_ulnoglong_remove_shareMemory(C.ulonglong(MPSecureDispath.shmsize))
	}
	fmt.Println("TEEWriterSeucreDispacth Close")
	return nil
}


// ==================================================================================
//				The new dir Multi-process Keystone Decrypt secure dispatch
// ==================================================================================

type TheNewDirMultiProcessTEESecureDispatch struct {
	shmaddr     []byte				  	// 共享内存的地址
	shmsize     uint64				  	// 共享内存的长度
	shmaddr_justcall     []byte				  	// 共享内存的地址
	shmsize_justcall     uint64				  	// 共享内存的长度
	fileCount	int64
	blockNum 	uint64
	blockcount int64
	blockbytes int64
	flexible int
	readCh chan struct{}          		// 通道用于通知读取完成
	mu     sync.Mutex             		// 互斥锁，保护共享资源
	closed bool                   		// 标记是否已经关闭
}

type TheNewDirMultiProcessTEESecureDispatchJustCall struct {
	shmaddr     []byte				  	// 共享内存的地址
	shmsize     uint64				  	// 共享内存的长度
	fileCount   int64
	flexible int
	transferfilereader *TheNewDirMultiProcessTEESecureDispatch
	readCh chan struct{}          		// 通道用于通知读取完成
	mu     sync.Mutex             		// 互斥锁，保护共享资源
	closed bool                   		// 标记是否已经关闭
}

// 创建一个新的共享内存段
func the_new_secure_dispatch_ulonglongcreateShm_just_call(shmsize uint64) ([]byte, error) {

	shmaddr := C.the_new_dir_secure_dispatch_ulnoglong_create_shareMemory_just_call(C.ulonglong(shmsize))

	// 错误写法 (*[size]byte)中 size 必须为常量，只是类型转换，并没有分配空间
	// return (*[size]byte)(shmaddr)[:], nil
	// [low:high:max] 获取内存切片low-high 可以索引low-high  数组实际空间大小为max
	// 若不指定 max 则是前面类型的空间，即1 << 32 = 1GB
	return (*[1 << 32]byte)(shmaddr)[:shmsize:shmsize], nil
}

// NewTheNewDirMultiProcessTEESecureDispatchJustCall TheNewDirMultiProcessTEESecureDispatchJustCall
// Just call keystone, it cant receive data dont know size
func NewTheNewDirMultiProcessTEESecureDispatchJustCall(isAES int, flexible int) (*TheNewDirMultiProcessTEESecureDispatchJustCall, error) {

	// MAXNUM <= 10
	C.fixFlexibleNum(unsafe.Pointer(&flexible))

	// 创建共享内存片段
	shmsize := uint64(C.TheNewDirMultiProcessTEESecureDispatchGetSHMSizeJustCall(C.int(flexible)))
	shmaddr, err := the_new_secure_dispatch_ulonglongcreateShm_just_call(shmsize)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create shared memory: %v\n", err)
		os.Exit(1)
	}
	
	reader := &TheNewDirMultiProcessTEESecureDispatchJustCall{
		shmaddr:    shmaddr,
		shmsize:	shmsize,
		flexible: 	flexible,
		transferfilereader: nil,
		fileCount:	0,
		readCh: 	make(chan struct{}, 1),
		closed: 	false,
	}

	C.the_new_secure_dispacth_initSHM_just_call(unsafe.Pointer(&reader.shmaddr[0]), C.int(flexible));

	// 获取当前ms_group的 engine_id  
	// dispatch
	dispatchEngineSeq := GetDispathEngineSeq()

	for numflexible:=0;numflexible<flexible;numflexible++ {

		// 启动第一个子进程，读取文件的前半部分
		cmd := exec.Command("./the_new_dir_secure_dispatch_child_process", 
			fmt.Sprintf("%d", isAES), 
			fmt.Sprintf("%d", shmsize), 
			fmt.Sprintf("%d", numflexible), 
			fmt.Sprintf("%d", flexible),
			fmt.Sprintf("%d", dispatchEngineSeq),
		)

		err := cmd.Start()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to start %d child process: %v\n", numflexible, err)
			os.Exit(1)
		}
	}
	
	C.the_new_secure_dispatch_waitKeystoneReady_just_call(unsafe.Pointer(&reader.shmaddr[0]), C.int(flexible))

	fmt.Println("ipfs-keystone testing ready just call")

	return reader, nil
}

func The_New_DIR_MultiProcess_Secure_Dispatch_Ipfs_keystone_test(isAES int, flexible int) (TheNewDirMultiProcessTEESecureDispatchJustCall){

	// 打印
	fmt.Println("The new dir multiProcess secure dispatch Processing...")

	reader, _ := NewTheNewDirMultiProcessTEESecureDispatchJustCall(isAES, flexible)

	return *reader
}

// set filesize
func TheNewDirSecureDispathSetLength(tee_just_call_reader *TheNewDirMultiProcessTEESecureDispatchJustCall, shmsize uint64){
	var blockNum uint64
	// fmt.Printf("shmsize: %d\n", shmsize)
	shmaddr := C.thenewdirsecuredispathSetLength(unsafe.Pointer(&tee_just_call_reader.shmaddr[0]), unsafe.Pointer(&blockNum), unsafe.Pointer(&shmsize), C.int(tee_just_call_reader.flexible))

	// fmt.Printf("shmsize: %d\n", shmsize)
	if shmsize == 0 {
		tee_just_call_reader.fileCount = 0
		return
	}
	tee_just_call_reader.fileCount++
	// 错误写法 (*[size]byte)中 size 必须为常量，只是类型转换，并没有分配空间
	// return (*[size]byte)(shmaddr)[:], nil
	// [low:high:max] 获取内存切片low-high 可以索引low-high  数组实际空间大小为max
	// 若不指定 max 则是前面类型的空间，即1 << 32 = 4GB
	reader := &TheNewDirMultiProcessTEESecureDispatch{
		shmaddr:    (*[1 << 32]byte)(shmaddr)[:shmsize:shmsize],
		shmsize:	shmsize,
		shmaddr_justcall:   tee_just_call_reader.shmaddr,
		shmsize_justcall:	tee_just_call_reader.shmsize,
		flexible: 	tee_just_call_reader.flexible,
		blockNum: 	blockNum,
		fileCount:	tee_just_call_reader.fileCount,
		blockcount: 0,
		blockbytes: 0,
		readCh: 	make(chan struct{}, 1),
		closed: 	false,
	}
	
	tee_just_call_reader.transferfilereader = reader
}

func TheNewDirSecureDispathWaitTransferKeystoneReady(tee_just_call_reader *TheNewDirMultiProcessTEESecureDispatchJustCall)(*TheNewDirMultiProcessTEESecureDispatch) {
	C.the_new_secure_dispatch_wait_transfer_keystone_ready(unsafe.Pointer(&tee_just_call_reader.shmaddr[0]), C.int(tee_just_call_reader.flexible))
	return tee_just_call_reader.transferfilereader
}


// Write 实现io.Write接口的方法，从p切片读取数据到缓冲区
func (theNDMPSecureDispath *TheNewDirMultiProcessTEESecureDispatch) Write(p []byte) (int, error) {
	theNDMPSecureDispath.mu.Lock()
	defer theNDMPSecureDispath.mu.Unlock()

	if theNDMPSecureDispath.closed {
		return 0, io.EOF
	}

	var readLen C.int = 0;
	
	// fmt.Println("ipfs testing dispath 1 block blockcount=%d", theNDMPSecureDispath.blockcount)
	// fmt.Println("ipfs testing dispath 1 block blockbytes=%d", theNDMPSecureDispath.blockbytes)

	result := C.the_new_secure_dispatch_write(unsafe.Pointer(&theNDMPSecureDispath.shmaddr[0]), C.longlong(theNDMPSecureDispath.shmsize), (*C.char)(unsafe.Pointer(&p[0])), C.int(len(p)), &readLen, C.int(theNDMPSecureDispath.flexible))

	if result == 0 {
		return int(readLen), io.EOF
	}

	return int(readLen), nil
}

// Close 关闭TEEFileReader实例，释放相关资源
func (theNDMPSecureDispath *TheNewDirMultiProcessTEESecureDispatch) Close() error {
	theNDMPSecureDispath.mu.Lock()
	defer theNDMPSecureDispath.mu.Unlock()

	if !theNDMPSecureDispath.closed {
		theNDMPSecureDispath.closed = true
		close(theNDMPSecureDispath.readCh)  // 确保通道被关闭
		// defer detachShm(theNDMPSecureDispath.shmaddr)

		// 等待 Keystone done
		fmt.Println("ipfs testing wait keystone done")
		C.the_new_secure_dispatch_wait_transfer_keystoneDone(unsafe.Pointer(&theNDMPSecureDispath.shmaddr_justcall[0]), C.int(theNDMPSecureDispath.flexible))

		// fmt.Printf("fileCount: %d\n", theNDMPSecureDispath.fileCount)
		// 断开连接共享内存
		defer C.the_new_secure_dispatch_detach_shareMemory(unsafe.Pointer(&theNDMPSecureDispath.shmaddr[0]))
		// 删除共享内存段
		defer C.the_new_secure_dispatch_ulnoglong_remove_shareMemory(C.ulonglong(theNDMPSecureDispath.shmsize), C.longlong(theNDMPSecureDispath.fileCount))
	}
	fmt.Println("the New TEEWriterSeucreDispacth Close")
	return nil
}


// ==================================================================================
//				The new dir Keystone Decrypt
// ==================================================================================

type TheNewDirTEEFileReaderJustCall struct {
	rb     *C.RingBuffer          // 指向C语言中的RingBuffer结构
	kjb    *C.KeystoneJustReady          
	readCh chan struct{}          // 通道用于通知读取完成
	wg     sync.WaitGroup         // 等待组用于等待后台goroutine完成
	mu     sync.Mutex             // 互斥锁，保护共享资源
	closed bool                   // 标记是否已经关闭
}

// TheNewDirTEEFileReader 结构体封装了环形缓冲区的相关操作
type TheNewDirTEEFileReader struct {
	rb     *C.RingBuffer          // 指向C语言中的RingBuffer结构
	kjb    *C.KeystoneJustReady
	readCh chan struct{}          // 通道用于通知读取完成
	wg     sync.WaitGroup         // 等待组用于等待后台goroutine完成
	mu     sync.Mutex             // 互斥锁，保护共享资源
	closed bool                   // 标记是否已经关闭
}

func NewTheNewDirTEEFileReaderJustCall(isAES int, FileName string) (*TheNewDirTEEFileReaderJustCall, error) {

	kjb := (*C.KeystoneJustReady)(C.malloc(C.sizeof_KeystoneJustReady))
	if kjb == nil { // 检查内存分配是否成功
		return nil, fmt.Errorf("failed to allocate memory for KeystoneJustReady")
	}

	rb := (*C.RingBuffer)(C.malloc(C.sizeof_RingBuffer))
	if rb == nil { // 检查内存分配是否成功
		return nil, fmt.Errorf("failed to allocate memory for RingBuffer")
	}

	// Convert Go int to C int
	cIsAES := C.int(isAES)

	C.init_keystone_just_ready(kjb)
	C.init_ring_buffer(rb)

	kjbreader := &TheNewDirTEEFileReaderJustCall{
		kjb:     kjb,
		rb:     rb,
		readCh: make(chan struct{}, 1),
		closed: false,
	}

	kjbreader.wg.Add(1)
	go func() {
		defer kjbreader.wg.Done() // 确保在goroutine结束时调用Done
		C.the_new_dir_ipfs_keystone_de(cIsAES, unsafe.Pointer(C.CString(FileName)), unsafe.Pointer(kjb), unsafe.Pointer(rb))
		fmt.Println("the new dir TEE read file done")
	}()

	C.the_new_dir_keystone_wait_ready(unsafe.Pointer(kjb))

	return kjbreader, nil
}

func The_New_DIR_Ipfs_keystone_test_de(isAES int, FileName string) (TheNewDirTEEFileReaderJustCall){

	// 打印FileName
	fmt.Println("The New Dir Get file:", FileName)

	kjbreader, _ := NewTheNewDirTEEFileReaderJustCall(isAES, FileName)

	return *kjbreader
}

// set filesize
func TheNewDirKeystoneDecryptSetLength(kjbreader *TheNewDirTEEFileReaderJustCall, shmsize uint64){
	// fmt.Printf("shmsize: %d\n", shmsize)
	C.thenewdirkeystonedecryptSetLength(unsafe.Pointer(kjbreader.kjb), C.ulonglong(shmsize))

	// fmt.Printf("shmsize: %d\n", shmsize)
	if shmsize == 0 {
		return
	}
	// return
}

func TheNewDirWaitKeystoneFileReady(kjbreader *TheNewDirTEEFileReaderJustCall) (TheNewDirTEEFileReader) {
	C.the_new_dir_wait_keystone_file_ready(unsafe.Pointer(kjbreader.kjb))

	rbreader := &TheNewDirTEEFileReader {
		kjb:    kjbreader.kjb,
		rb:     kjbreader.rb,
		readCh: make(chan struct{}, 1),
		closed: false,
	}

	return *rbreader
}


// Write 实现io.Write接口的方法，从p切片读取数据到缓冲区
func (r *TheNewDirTEEFileReader) Write(p []byte) (int, error) {
	// fmt.Printf("test 3\n");
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


// Close 关闭TheNewDirTEEFileReader实例，释放相关资源
func (r *TheNewDirTEEFileReader) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.closed {
		r.closed = true
		C.ring_buffer_stop((*C.RingBuffer)(r.rb));
		// C.free(unsafe.Pointer(r.rb))  // 由c语言程序释放内存
		close(r.readCh)  // 确保通道被关闭
		fmt.Println("wait TheNewDirTEEFileReader Close")
		C.the_new_dir_wait_keystone_file_end((*C.KeystoneJustReady)(r.kjb))
		// time.Sleep(1500 * time.Millisecond)  // 固定等待1.5s
		// r.wg.Wait()  // 等待后台goroutine完成
	}
	fmt.Println("TheNewDirTEEFileReader Close")
	return nil
}


// ==================================================================================
//				The New Dir Multi-process Cross-read Flexible Keystone Encrypt
// ==================================================================================


type TheNewDirMultiProcessCrossTEEFileFlexibleReaderJustCall struct {
	shmaddr     []byte				  	// 共享内存的地址
	shmsize     int64				  	// 共享内存的长度
	fileCount	int64
	flexible 	int
	readCh chan struct{}          		// 通道用于通知读取完成
	mu     sync.Mutex             		// 互斥锁，保护共享资源
	closed bool                   		// 标记是否已经关闭
}

type TheNewDirMultiProcessCrossTEEFileFlexibleReader struct {
	shmaddr     []byte				  	// 共享内存的地址
	shmsize     int64				  	// 共享内存的长度
	fileCount	int64
	flexible 	int
	shmaddr_justcall     []byte				  	// 共享内存的地址
	shmsize_justcall     int64				  	// 共享内存的长度
	readCh chan struct{}          		// 通道用于通知读取完成
	mu     sync.Mutex             		// 互斥锁，保护共享资源
	closed bool                   		// 标记是否已经关闭
}

// 创建一个新的共享内存段
func theNewDirlongcreateShm(size int64) ([]byte, error) {

	shmaddr := C.the_new_dir_long_create_shareMemory(C.longlong(size))

	// 错误写法 (*[size]byte)中 size 必须为常量，只是类型转换，并没有分配空间
	// return (*[size]byte)(shmaddr)[:], nil
	// [low:high:max] 获取内存切片low-high 可以索引low-high  数组实际空间大小为max
	// 若不指定 max 则是前面类型的空间，即1 << 32 = 1GB
	return (*[1 << 32]byte)(shmaddr)[:size:size], nil
}

// NewTheNewDirMultiProcessCrossTEEFileFlexibleReaderJustCall TheNewDirMultiProcessCrossTEEFileFlexibleReaderJustCall
func NewTheNewDirMultiProcessCrossTEEFileFlexibleReaderJustCall(isAES int, flexible int) (*TheNewDirMultiProcessCrossTEEFileFlexibleReaderJustCall, error) {
	
	// MAXNUM 10
	C.fixFlexibleNum(unsafe.Pointer(&flexible))

	// 创建共享内存片段
	shmsize := int64(C.sizeof_TheNewDirMultiProcessCrossFlexibleSHMBufferJustCall + (flexible * C.sizeof_int) + (flexible * C.sizeof_longlong))
	shm, err := theNewDirlongcreateShm(shmsize)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create shared memory: %v\n", err)
		os.Exit(1)
	}

	reader := &TheNewDirMultiProcessCrossTEEFileFlexibleReaderJustCall{
		shmaddr:    shm,
		shmsize:	shmsize,
		fileCount:	0,
		flexible: flexible,
		readCh: make(chan struct{}, 1),
		closed: false,
	}

	// 启动keystone之前先初始化内存空间
	C.theNewDirflexiblecrossInitSHMJustCall(unsafe.Pointer(&reader.shmaddr[0]), C.int(flexible))
	// fmt.Println("MultiProcess Processing file test")

	var numflexible int = 0
	for numflexible < flexible {

		// fmt.Printf("isAES:%d, shmsize:%d, numflexible:%d, flexible:%d\n", isAES, shmsize, numflexible, flexible)

		// 启动第一个子进程，读取文件的前半部分
		cmd := exec.Command("./the_new_dir_flexible_cross_child_process", 
			fmt.Sprintf("%d", isAES), 
			fmt.Sprintf("%d", shmsize), 
			fmt.Sprintf("%d", numflexible), 
			fmt.Sprintf("%d", flexible),
		)

		err = cmd.Start()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to start %d child process: %v\n", numflexible, err)
			os.Exit(1)
		}

		numflexible++

	}

	C.theNewDirflexiblecrosswaitKeystoneReady(unsafe.Pointer(&reader.shmaddr[0]), C.int(flexible))

	return reader, nil
}

func The_New_Dir_MultiProcess_Cross_Flexible_Ipfs_keystone_test_just_call(isAES int, flexible int) (TheNewDirMultiProcessCrossTEEFileFlexibleReaderJustCall){

	// 打印FileName
	fmt.Println("The New Dir MultiProcess flexible Processing file")

	reader, _ := NewTheNewDirMultiProcessCrossTEEFileFlexibleReaderJustCall(isAES, flexible)

	return *reader
}


// 创建一个新的共享内存段
func theNewDirlongcreateShmofFile(size int64, fileCount int64) ([]byte, error) {

	shmaddr := C.the_new_dir_long_create_shareMemory_of_file(C.longlong(size), C.longlong(fileCount))

	// 错误写法 (*[size]byte)中 size 必须为常量，只是类型转换，并没有分配空间
	// return (*[size]byte)(shmaddr)[:], nil
	// [low:high:max] 获取内存切片low-high 可以索引low-high  数组实际空间大小为max
	// 若不指定 max 则是前面类型的空间，即1 << 32 = 1GB
	return (*[1 << 32]byte)(shmaddr)[:size:size], nil
}

func (thenewdirReader *TheNewDirMultiProcessCrossTEEFileFlexibleReaderJustCall) The_New_Dir_MultiProcess_cross_Flexible_Set_fileAbsPath(fpath string, fileSize int64)(*TheNewDirMultiProcessCrossTEEFileFlexibleReader){
	// fmt.Printf("thenewdirReader fpath:%s, fileSize:%d, thenewdirReader.fileCount:%d\n", fpath, fileSize, thenewdirReader.fileCount)

	if fileSize == 0 || fpath == ""{
		C.theNewDirflexiblecrosswaitKeystoneTransferFilesReady(unsafe.Pointer(&thenewdirReader.shmaddr[0]), C.int(thenewdirReader.flexible), nil, 0, 0, nil)
		return nil
	}

	cFileSize := C.long_alignedFileSize(C.longlong(fileSize))
	cBlocksNums := C.long_alignedFileSize_blocksnums(cFileSize)

	thenewdirReader.fileCount++

	// 创建共享内存片段
	shmsize := C.sizeof_TheNewDirMultiProcessCrossFlexibleSHMBufferReader + int64(cBlocksNums * C.sizeof_int) + int64(cFileSize)
	shm, err := theNewDirlongcreateShmofFile(shmsize, thenewdirReader.fileCount)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create shared memory: %v\n", err)
		os.Exit(1)
	}

	reader := &TheNewDirMultiProcessCrossTEEFileFlexibleReader{
		shmaddr:    shm,
		shmsize:	shmsize,
		fileCount:	thenewdirReader.fileCount,
		flexible: 	thenewdirReader.flexible,
		shmaddr_justcall:    thenewdirReader.shmaddr,
		shmsize_justcall:	thenewdirReader.shmsize,
		readCh: 	make(chan struct{}, 1),
		closed: 	false,
	}

	C.theNewDirflexiblecrosswaitKeystoneTransferFilesReady(unsafe.Pointer(&thenewdirReader.shmaddr[0]), C.int(thenewdirReader.flexible), unsafe.Pointer(&reader.shmaddr[0]), cBlocksNums, cFileSize, unsafe.Pointer(C.CString(fpath)))
	// fmt.Println("MultiProcess Processing file test")

	return reader
	
}


func (r *TheNewDirMultiProcessCrossTEEFileFlexibleReader)Read(p []byte) (int, error)  {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return 0, io.EOF
	}

	// fmt.Println("The New Dir MultiProcess Processing read start")
	var readLen C.int = 0;
	// 交给c语言函数处理
	result := C.TheNewDirMultiProcessCrossReadFlexible(unsafe.Pointer(&r.shmaddr[0]), C.int(r.shmsize), unsafe.Pointer(&p[0]), C.int(len(p)), &readLen);
	if result == 0 {
		return int(readLen), io.EOF
	}

	// fmt.Println("The New Dir MultiProcess Processing read done")

	return int(readLen), nil;
}

// 删除共享内存段
func the_new_dir_flexbile_longremoveShm(shmsize int64, fileCount int64) error {
	C.the_new_dir_flexbile_long_removeShm(C.longlong(shmsize), C.longlong(fileCount))

	return nil
}

// Close 关闭 TheNewDirMultiProcessCrossTEEFileFlexibleReader 实例，释放相关资源
func (r *TheNewDirMultiProcessCrossTEEFileFlexibleReader) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.closed {
		r.closed = true
		close(r.readCh)  // 确保通道被关闭
		fmt.Println("The New Dir MultiProcess Cross Flexible wait TEEFileReader end")
		C.theNewDirflexiblecrosswaitKeystoneTransferFilesEnd(unsafe.Pointer(&r.shmaddr_justcall[0]), C.int(r.flexible))
		defer detachShm(r.shmaddr)
		defer the_new_dir_flexbile_longremoveShm(r.shmsize, r.fileCount)
	}
	fmt.Println("The New Dir MultiProcess Cross Flexible TEEFileReader Close")
	return nil
}


// ==================================================================================
//				The New Dir Keystone Encrypt
// ==================================================================================


type TheNewDirTEEFileReaderJustCallADD struct {
	rb     *C.RingBuffer          // 指向C语言中的RingBuffer结构
	kjb    *C.KeystoneJustReadyAdd          
	readCh chan struct{}          // 通道用于通知读取完成
	wg     sync.WaitGroup         // 等待组用于等待后台goroutine完成
	mu     sync.Mutex             // 互斥锁，保护共享资源
	closed bool                   // 标记是否已经关闭
}

// TheNewDirTEEFileReader 结构体封装了环形缓冲区的相关操作
type TheNewDirTEEFileReaderADD struct {
	rb     *C.RingBuffer          // 指向C语言中的RingBuffer结构
	kjb    *C.KeystoneJustReadyAdd
	readCh chan struct{}          // 通道用于通知读取完成
	wg     sync.WaitGroup         // 等待组用于等待后台goroutine完成
	mu     sync.Mutex             // 互斥锁，保护共享资源
	closed bool                   // 标记是否已经关闭
}

func NewTheNewDirTEEFileReaderJustCallADD(isAES int) (*TheNewDirTEEFileReaderJustCallADD, error) {

	kjb := (*C.KeystoneJustReadyAdd)(C.malloc(C.sizeof_KeystoneJustReadyAdd))
	if kjb == nil { // 检查内存分配是否成功
		return nil, fmt.Errorf("failed to allocate memory for KeystoneJustReadyAdd")
	}

	rb := (*C.RingBuffer)(C.malloc(C.sizeof_RingBuffer))
	if rb == nil { // 检查内存分配是否成功
		return nil, fmt.Errorf("failed to allocate memory for RingBuffer")
	}

	// Convert Go int to C int
	cIsAES := C.int(isAES)

	C.init_keystone_just_ready_add(kjb)
	C.init_ring_buffer(rb)

	kjbreader := &TheNewDirTEEFileReaderJustCallADD{
		kjb:     kjb,
		rb:     rb,
		readCh: make(chan struct{}, 1),
		closed: false,
	}

	kjbreader.wg.Add(1)
	go func() {
		defer kjbreader.wg.Done() // 确保在goroutine结束时调用Done
		C.the_new_dir_ipfs_keystone(cIsAES, unsafe.Pointer(kjb), unsafe.Pointer(rb))
		fmt.Println("the new dir TEE read file done")
	}()

	C.the_new_dir_keystone_wait_ready_add(unsafe.Pointer(kjb))

	return kjbreader, nil
}

func The_New_DIR_Ipfs_keystone_test(isAES int) (TheNewDirTEEFileReaderJustCallADD){

	// 打印FileName
	fmt.Println("The New Dir add file:")

	kjbreader, _ := NewTheNewDirTEEFileReaderJustCallADD(isAES)

	return *kjbreader
}

func (thenewdirReader *TheNewDirTEEFileReaderJustCallADD) The_New_Dir_Keystone_Set_fileAbsPath(fpath string, fileSize int64)(*TheNewDirTEEFileReaderADD){
	// fmt.Printf("thenewdirReader fpath:%s, fileSize:%d\n", fpath, fileSize)

	if fileSize == 0 || fpath == "" {
		C.theNewDirKeystoneTransferFilesReady(unsafe.Pointer(thenewdirReader.kjb), 0, nil)
		return nil
	}

	// C.init_ring_buffer(thenewdirReader.rb)

	reader := &TheNewDirTEEFileReaderADD{
		kjb:    thenewdirReader.kjb,
		rb:     thenewdirReader.rb,
		readCh: make(chan struct{}, 1),
		closed: false,
	}

	C.theNewDirKeystoneTransferFilesReady(unsafe.Pointer(thenewdirReader.kjb), C.longlong(fileSize), unsafe.Pointer(C.CString(fpath)))

	return reader
	
}

// Read 实现io.Reader接口的方法，从缓冲区读取数据到p切片
func (r *TheNewDirTEEFileReaderADD) Read(p []byte) (int, error) {
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


// Close 关闭TheNewDirTEEFileReaderADD实例，释放相关资源
func (r *TheNewDirTEEFileReaderADD) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.closed {
		r.closed = true
		// C.free(unsafe.Pointer(r.rb))  // 释放C语言分配的内存
		C.the_new_dir_wait_keystone_file_end_add((*C.KeystoneJustReadyAdd)(r.kjb), (*C.RingBuffer)(r.rb))
		close(r.readCh)  // 确保通道被关闭
		// r.wg.Wait()  // 等待后台goroutine完成
	}
	fmt.Println("TEEFileReader Close")
	return nil
}
