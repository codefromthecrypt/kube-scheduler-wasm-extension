/*
   Copyright 2023 The Kubernetes Authors.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package imports

import "unsafe"

// BufLimit is the possibly zero maximum length of a result value to write in
// bytes. If the actual value is larger than this, nothing is written to
// memory.
type bufLimit = uint32

var (
	// readBuf is sharable because there is no parallelism in wasm.
	readBuf = make([]byte, readBufLimit)
	// ReadBufPtr is used to avoid duplicate host function calls.
	readBufPtr = uintptr(unsafe.Pointer(&readBuf[0]))
	// ReadBufLimit is constant memory overhead for reading fields.
	readBufLimit = uint32(2048)
)

// stringToPtr returns a pointer and size pair for the given string in a way
// compatible with WebAssembly numeric types.
// The returned pointer aliases the string hence the string must be kept alive
// until ptr is no longer needed.
func stringToPtr(s string) (uint32, uint32) {
	ptr := unsafe.Pointer(unsafe.StringData(s))
	return uint32(uintptr(ptr)), uint32(len(s))
}

func getBytes(fn func(ptr uint32, limit bufLimit) (len uint32)) []byte {
	size := fn(uint32(readBufPtr), readBufLimit)
	if size == 0 {
		return nil
	}

	// Ensure the result isn't a shared buffer.
	buf := make([]byte, size)

	// If the function result fit in our read buffer, copy it out.
	if size <= readBufLimit {
		copy(buf, readBuf)
		return buf
	}

	// If the size returned from the function was larger than our read buffer,
	// we need to execute it again. buf is exactly the right size now.
	ptr := uintptr(unsafe.Pointer(&buf[0]))
	_ = fn(uint32(ptr), size)
	return buf
}

// conditionalBytesReader reads bytes into memory up to the given limit, if the
// ID is out-of-date.
//
// The etagLen result includes the real etag and the length of its data, as two
// uint32 values packed into a single uint64. When the length is larger than
// bufLimit, the function must be re-tried with a larger one.
type conditionalBytesReader func(id, ptr uint32, limit bufLimit) (idLen uint64)

// unmarshal invokes `conditionalBytesReader` with the given ID. If the
// resulting ID is the same, it returns nil. Otherwise, it runs the updater
// and returns any error that may have occurred.
func conditionallyUpdate(
	id uint32,
	conditionalBytesReader conditionalBytesReader,
	updater func(newID uint32, bytes []byte) error,
) error {
	// Call the function that reads bytes if the ID is different.
	idLen := conditionalBytesReader(id, uint32(readBufPtr), readBufLimit)
	currentID := uint32(idLen >> 32)

	// If the current ID is the same as input, we shouldn't decode.
	if id == currentID {
		return nil
	}

	// If the size in bytes to update is larger than our read buffer, retry
	// using the out-of-date ID, so that the host updates the memory.
	var buf []byte
	if size := uint32(idLen); size > readBufLimit {
		buf = make([]byte, size)
		ptr := uintptr(unsafe.Pointer(&buf[0]))
		_ = conditionalBytesReader(id, uint32(ptr), size)
	} else { // re-slice the buffer to size.
		buf = readBuf[:size]
	}

	// Pass the current ID to the updater along with the new raw bytes.
	return updater(currentID, buf)
}

func getString(fn func(ptr uint32, limit bufLimit) (len uint32)) string {
	size := fn(uint32(readBufPtr), readBufLimit)
	if size == 0 {
		return ""
	}

	// If the function result fit in our read buffer, copy it out.
	if size <= readBufLimit {
		return string(readBuf[:size])
	}

	// If the size returned from the function was larger than our read buffer,
	// we need to execute it again. Make a buffer of exactly the right size.
	buf := make([]byte, size)
	ptr := unsafe.Pointer(&buf[0])
	_ = fn(uint32(uintptr(ptr)), size)

	// TODO: track https://github.com/tinygo-org/tinygo/issues/3789
	return unsafe.String((*byte)(ptr), size /* unsafe.IntegerType */)
}
