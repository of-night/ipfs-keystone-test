package ipfsKeystone

// #cgo LDFLAGS: -L./ -lipfs_keystone 
// #cgo CFLAGS: -I./include -I./include/host -I./include/edge
// #include "ipfs_keystone.h"

import "C"

import (
	"fmt"
)

func ipfs_keystone_test(isAES int) {

    // Convert Go int to C int
    cIsAES := C.int(isAES)
    
    // Call the C function
    C.ipfs_keystone(cIsAES)

}