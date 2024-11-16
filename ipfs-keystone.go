package ipfsKeystoneTest

// #cgo LDFLAGS: -L/usr/local/ipfs-keystone -lipfs_keystone -lstdc++
// #cgo CFLAGS: -I/usr/local/ipfs-keystone/include -I/usr/local/ipfs-keystone/include/host -I/usr/local/ipfs-keystone/include/edge
// #include "ipfs_keystone.h"
import "C"

func Ipfs_keystone_test(isAES int) {

    // Convert Go int to C int
    cIsAES := C.int(isAES)

    // Call the C function
    C.ipfs_keystone(cIsAES)

}

