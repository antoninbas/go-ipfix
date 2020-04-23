package ipfix

/*
#include <sys/types.h> // for ssize_t
#include <ipfix.h>
#include <ipfix_def_antrea.h>
#include <ipfix_fields_antrea.h>
#include <stdlib.h>
#include <string.h>

int registerAntrea() {
    return ipfix_add_vendor_information_elements(ipfix_ft_antrea);
}

void **makeFieldArray(int nFields) {
    return malloc(nFields * sizeof(void *));
}
uint16_t *makeLengthArray(int nFields) {
    return malloc(nFields * sizeof(uint16_t));
}
void freeFieldArray(void **fa, int s) {
    for (int i = 0; i < s; i++) {
        free(fa[i]);
    }
    free(fa);
}
void freeLengthArray(uint16_t *la) {
    free(la);
}
// TODO: check endianness, ensure whether this is handled correctly by libipfix
void addU8(void **fa, uint16_t *la, int idx, uint8_t f) {
    uint8_t *d = malloc(sizeof(f));
    *d = f;
    fa[idx] = (void *)d;
    la[idx] = 1;
}
void addU16(void **fa, uint16_t *la, int idx, uint16_t f) {
    uint16_t *d = malloc(sizeof(f));
    *d = f;
    fa[idx] = (void *)d;
    la[idx] = 2;
}
void addU32(void **fa, uint16_t *la, int idx, uint32_t f) {
    uint32_t *d = malloc(sizeof(f));
    *d = f;
    fa[idx] = (void *)d;
    la[idx] = 4;
}
void addU64(void **fa, uint16_t *la, int idx, uint64_t f) {
    uint64_t *d = malloc(sizeof(f));
    *d = f;
    fa[idx] = (void *)d;
    la[idx] = 8;
}
void addString(void **fa, uint16_t *la, int idx, char *f) {
    fa[idx] = (void *)f;
    la[idx] = strlen(f);
}
*/
import "C"

import (
	"fmt"
	"unsafe"
)

const (
	IPFIX_VERSION = 0x0A
)

func Init() error {
	rc := C.ipfix_init()
	if rc != 0 {
		return fmt.Errorf("ipfix_init returned error code %d", rc)
	}
	// shortcut for now, this libray has no reason to be antrea-specific
	rc = C.registerAntrea()
	if rc != 0 {
		return fmt.Errorf("error code %d when registering antrea elements", rc)
	}
	return nil
}

func Cleanup() error {
	C.ipfix_cleanup()
	return nil
}

type IPFix struct {
	ptr *C.ipfix_t
}

func NewIPFix(sourceId int) (*IPFix, error) {
	ipf := &IPFix{}
	rc := C.ipfix_open(&ipf.ptr, C.int(sourceId), C.int(IPFIX_VERSION))
	if rc != 0 {
		return nil, fmt.Errorf("ipfix_open returned error code %d", rc)
	}
	return ipf, nil
}

func (ipf *IPFix) Close() error {
	C.ipfix_close(ipf.ptr)
	return nil
}

type IPFixProto string

const (
	IPFixProtoTCP IPFixProto = "TCP"
	IPFixProtoUDP IPFixProto = "UDP"
)

func (ipf *IPFix) AddCollector(host string, port int, proto IPFixProto) error {
	// TODO: this should be freed at some point, but I don't know if
	// libipfix makes a copy.
	chost := C.CString(host)
	var cproto C.ipfix_proto_t
	switch proto {
	case IPFixProtoTCP:
		cproto = C.IPFIX_PROTO_TCP
	case IPFixProtoUDP:
		cproto = C.IPFIX_PROTO_UDP
	default:
		return fmt.Errorf("Unknown bearer protocol: %s", proto)
	}
	rc := C.ipfix_add_collector(ipf.ptr, chost, C.int(port), cproto)
	if rc != 0 {
		return fmt.Errorf("ipfix_add_collector returned error code %d", rc)
	}
	return nil
}

type Template struct {
	ptr *C.ipfix_template_t
}

func (ipf *IPFix) NewDataTemplate(nFields int) (*Template, error) {
	tpl := &Template{}
	rc := C.ipfix_new_data_template(ipf.ptr, &tpl.ptr, C.int(nFields))
	if rc != 0 {
		return nil, fmt.Errorf("ipfix_new_data_template returned error code %d", rc)
	}
	return tpl, nil
}

func (ipf *IPFix) NewOptionTemplate(nFields int) (*Template, error) {
	tpl := &Template{}
	rc := C.ipfix_new_option_template(ipf.ptr, &tpl.ptr, C.int(nFields))
	if rc != 0 {
		return nil, fmt.Errorf("ipfix_new_option_template returned error code %d", rc)
	}
	return tpl, nil
}

func (ipf *IPFix) AddField(tpl *Template, eno uint32, fType uint16, fLen uint16) error {
	rc := C.ipfix_add_field(ipf.ptr, tpl.ptr, C.uint32_t(eno), C.uint16_t(fType), C.uint16_t(fLen))
	if rc != 0 {
		return fmt.Errorf("ipfix_add_field returned error code %d", rc)
	}
	return nil
}

func (ipf *IPFix) AddScopeField(tpl *Template, eno uint32, fType uint16, fLen uint16) error {
	rc := C.ipfix_add_scope_field(ipf.ptr, tpl.ptr, C.uint32_t(eno), C.uint16_t(fType), C.uint16_t(fLen))
	if rc != 0 {
		return fmt.Errorf("ipfix_add_scope_field returned error code %d", rc)
	}
	return nil
}

func (ipf *IPFix) DeleteTemplate(tpl *Template) error {
	C.ipfix_delete_template(ipf.ptr, tpl.ptr)
	return nil
}

type FieldArray struct {
	tpl     *Template
	nFields int
	idx     int
	array   *unsafe.Pointer
	lengths *C.uint16_t
}

func NewFieldArray(tpl *Template, nFields int) *FieldArray {
	fa := C.makeFieldArray(C.int(nFields))
	la := C.makeLengthArray(C.int(nFields))
	return &FieldArray{tpl, nFields, 0, fa, la}
}

func (fArray *FieldArray) Delete() {
	C.freeFieldArray(fArray.array, C.int(fArray.idx))
	C.freeLengthArray(fArray.lengths)
}

func (fArray *FieldArray) AddU8(f uint8) error {
	C.addU8(fArray.array, fArray.lengths, C.int(fArray.idx), C.uint8_t(f))
	fArray.idx++
	return nil
}

func (fArray *FieldArray) AddU16(f uint16) error {
	C.addU16(fArray.array, fArray.lengths, C.int(fArray.idx), C.uint16_t(f))
	fArray.idx++
	return nil
}

func (fArray *FieldArray) AddU32(f uint32) error {
	C.addU32(fArray.array, fArray.lengths, C.int(fArray.idx), C.uint32_t(f))
	fArray.idx++
	return nil
}

func (fArray *FieldArray) AddU64(f uint64) error {
	C.addU64(fArray.array, fArray.lengths, C.int(fArray.idx), C.uint64_t(f))
	fArray.idx++
	return nil
}

func (fArray *FieldArray) AddString(f string) error {
	C.addString(fArray.array, fArray.lengths, C.int(fArray.idx), C.CString(f))
	fArray.idx++
	return nil
}

func (ipf *IPFix) Export(fArray *FieldArray) error {
	rc := C.ipfix_export_array(ipf.ptr, fArray.tpl.ptr, C.int(fArray.idx), fArray.array, fArray.lengths)
	if rc != 0 {
		return fmt.Errorf("ipfix_export_array returned error code %d", rc)
	}
	rc = C.ipfix_export_flush(ipf.ptr)
	if rc != 0 {
		return fmt.Errorf("ipfix_export_flush returned error code %d", rc)
	}
	return nil
}
