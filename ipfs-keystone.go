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
	// 若不指定 max 则是前面类型的空间，即1 << 30 = 1GB
	return (*[1 << 30]byte)(shmaddr)[:size:size], nil
}

// 连接到现有的共享内存段
func attachShm(size int) ([]byte, error) {

	shmaddr := C.attach_shareMemory(C.int(size))

	// 错误写法 (*[size]byte)中 size 必须为常量，只是类型转换，并没有分配空间
	// return (*[size]byte)(shmaddr)[:], nil
	// [low:high:max] 获取内存切片low-high 可以索引low-high  数组实际空间大小为max
	// 若不指定 max 则是前面类型的空间，即1 << 30 = 1GB
	return (*[1 << 30]byte)(shmaddr)[:size:size], nil
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
	// 若不指定 max 则是前面类型的空间，即1 << 30 = 1GB
	return (*[1 << 30]byte)(shmaddr)[:size:size], nil
}


// NewMultiProcessCrossTEEFileReader MultiProcessCrossTEEFileReader
func NewMultiProcessCrossTEEFileReader(isAES int, FileName string, fileSize int64) (*MultiProcessCrossTEEFileReader, error) {

	// Convert Go int to C int
	cFileSize := C.longlong(fileSize)

	cFileSize = C.long_alignedFileSize(cFileSize)
	cBlocksNums := C.long_alignedFileSize_blocksnums(cFileSize)
	
	// 创建共享内存片段
	shmsize := C.sizeof_MultiProcessCrossSHMBuffer + int64(cBlocksNums) + int64(cFileSize)
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

	// 启动第一个子进程，读取文件的前半部分
	cmd1 := exec.Command("./cross_child_process", fmt.Sprintf("%d", isAES), fmt.Sprintf("%d", shmsize), FileName, fmt.Sprintf("%d", 0))

	err = cmd1.Start()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start first child process: %v\n", err)
		os.Exit(1)
	}

	// 启动第二个子进程，读取文件的后半部分
	cmd2 := exec.Command("./cross_child_process", fmt.Sprintf("%d", isAES), fmt.Sprintf("%d", shmsize), FileName, fmt.Sprintf("%d", 1))

	err = cmd2.Start()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start second child process: %v\n", err)
		os.Exit(1)
	}

	C.crosswaitKeystoneReady(unsafe.Pointer(&reader.shmaddr[0]))


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

	var readLen C.int = 0;
	// 交给c语言函数处理
	result := C.MultiProcessCrossRead(unsafe.Pointer(&mpcr.shmaddr[0]), C.int(mpcr.shmsize), unsafe.Pointer(&p[0]), C.int(len(p)), &readLen);
	if result == 0 {
		return int(readLen), io.EOF
	}

	return int(readLen), nil;
}



