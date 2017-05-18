package argo

/*
#include <CoreFoundation/CoreFoundation.h>
#cgo LDFLAGS: -framework CoreFoundation
*/
import "C"

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"unsafe"
)

const (
	INPUT = iota
	OUTPUT
)

const (
	LOW = iota
	HIGH
)

func PinMode(pin, mode int) {

}

func DigitalWrite(pin, level int) {
	m, err := service.send(map[string]interface{}{"command": "digitalWrite", "param1": pin, "param2": level})
	if err != nil {
		panic(err)
	}
	if m["status"].(string) != "OK" {
		panic(fmt.Errorf("error received: %s", m))
	}
}

var service *portService

func init() {
	s, err := newService()
	if err != nil {
		panic(err)
	}
	service = s
}

type portService struct {
	port C.CFMessagePortRef
}

func newService() (*portService, error) {
	n := C.CString("ph.mac.ArgoSim.service")
	defer C.free(unsafe.Pointer(n))
	name := C.CFStringCreateWithCString(nil, n, C.kCFStringEncodingUTF8)
	defer C.CFRelease(C.CFTypeRef(name))
	port := C.CFMessagePortCreateRemote(nil, name)
	if port == nil {
		return nil, errors.New("couldn't find remote port")
	}
	return &portService{port}, nil
}

func (service *portService) send(m map[string]interface{}) (map[string]interface{}, error) {
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	err := encoder.Encode(m)
	if err != nil {
		return nil, errors.New("couldn't encode JSON")
	}
	b := buf.Bytes()
	d := C.CBytes(b)
	data := C.CFDataCreateWithBytesNoCopy(nil, (*C.UInt8)(d), C.CFIndex(len(b)), C.kCFAllocatorMalloc)
	defer C.CFRelease(C.CFTypeRef(data))
	var resp C.CFDataRef
	if C.CFMessagePortSendRequest(service.port, 0, data, 5, 5, C.kCFRunLoopDefaultMode, &resp) == C.kCFMessagePortSuccess {
		defer C.CFRelease(C.CFTypeRef(resp))
		resp := C.GoBytes(unsafe.Pointer(C.CFDataGetBytePtr(resp)), (C.int)(C.CFDataGetLength(resp)))
		decoder := json.NewDecoder(bytes.NewReader(resp))
		var m map[string]interface{}
		err := decoder.Decode(&m)
		if err != nil {
			return nil, errors.New("couldn't decode JSON")
		}
		return m, nil
	} else {
		panic("couldn't send request")
	}
}
