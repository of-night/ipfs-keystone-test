#ifndef __SM_ERR_H__
#define __SM_ERR_H__

#define SBI_ERR_SM_ENCLAVE_SUCCESS                     0
#define SBI_ERR_SM_ENCLAVE_UNKNOWN_ERROR               100000
#define SBI_ERR_SM_ENCLAVE_INVALID_ID                  100001
#define SBI_ERR_SM_ENCLAVE_INTERRUPTED                 100002
#define SBI_ERR_SM_ENCLAVE_PMP_FAILURE                 100003
#define SBI_ERR_SM_ENCLAVE_NOT_RUNNABLE                100004
#define SBI_ERR_SM_ENCLAVE_NOT_DESTROYABLE             100005
#define SBI_ERR_SM_ENCLAVE_REGION_OVERLAPS             100006
#define SBI_ERR_SM_ENCLAVE_NOT_ACCESSIBLE              100007
#define SBI_ERR_SM_ENCLAVE_ILLEGAL_ARGUMENT            100008
#define SBI_ERR_SM_ENCLAVE_NOT_RUNNING                 100009
#define SBI_ERR_SM_ENCLAVE_NOT_RESUMABLE               100010
#define SBI_ERR_SM_ENCLAVE_EDGE_CALL_HOST              100011
#define SBI_ERR_SM_ENCLAVE_NOT_INITIALIZED             100012
#define SBI_ERR_SM_ENCLAVE_NO_FREE_RESOURCE            100013
#define SBI_ERR_SM_ENCLAVE_SBI_PROHIBITED              100014
#define SBI_ERR_SM_ENCLAVE_ILLEGAL_PTE                 100015
#define SBI_ERR_SM_ENCLAVE_NOT_FRESH                   100016
#define SBI_ERR_SM_DEPRECATED                          100099
#define SBI_ERR_SM_NOT_IMPLEMENTED                     100100

#define SBI_ERR_SM_PMP_SUCCESS                         0
#define SBI_ERR_SM_PMP_REGION_SIZE_INVALID             100020
#define SBI_ERR_SM_PMP_REGION_NOT_PAGE_GRANULARITY     100021
#define SBI_ERR_SM_PMP_REGION_NOT_ALIGNED              100022
#define SBI_ERR_SM_PMP_REGION_MAX_REACHED              100023
#define SBI_ERR_SM_PMP_REGION_INVALID                  100024
#define SBI_ERR_SM_PMP_REGION_OVERLAP                  100025
#define SBI_ERR_SM_PMP_REGION_IMPOSSIBLE_TOR           100026

#endif  // __SM_ERR_H__
