package nickel

/*
#cgo linux,amd64 LDFLAGS: -L./lib/linux_amd64/ -lnickel_lang -lm

#cgo CFLAGS: -I./include

#include <nickel_lang.h>
#include <malloc.h>
*/
import "C"

import (
	"encoding/json"
	"runtime"
	"unsafe"
)

// Context is the main entry point.
//
// It allows you to customize various aspects of the Nickel interpreter, such
// as the path used to search for imported files.
type Context struct {
	ptr *C.nickel_context
}

// Expr is a Nickel expression.
//
// Since Nickel is lazy, it may not yet have been evaluated (TODO: link to the
// lazy eval once that's bound). If it has been evaluated, it could be null, a
// boolean, a number, a string, an enum, a record, or an array.
type Expr struct {
	ptr *C.nickel_expr
}

// Record is a Nickel map or dictionary.
//
// Every key is a string and every value is an Expr, which may not have
// been evaluated.
type Record struct {
	expr *Expr
	ptr *C.nickel_record
}

// Error is a Nickel error message.
type Error struct {
	ptr *C.nickel_error
}

func (e *Error) Error() string {
	s := C.nickel_string_alloc()
	defer C.nickel_string_free(s)

	result := C.nickel_error_format_as_string(e.ptr, s, C.NICKEL_ERROR_FORMAT_TEXT)
	if result == C.NICKEL_RESULT_ERR {
		return "error formatting error"
	} else {
		var len C.uintptr_t
		var bytes *C.char
		C.nickel_string_data(s, &bytes, &len)
		return C.GoStringN(bytes, C.int(len))
	}
}

func new_expr() *Expr {
	expr := &Expr {
		ptr: C.nickel_expr_alloc(),
	}

	runtime.SetFinalizer(expr, func(expr *Expr) {
		C.nickel_expr_free(expr.ptr)
	})

	return expr
}

func new_err() *Error {
	err := &Error {
		ptr: C.nickel_error_alloc(),
	}

	runtime.SetFinalizer(err, func(err *Error) {
		C.nickel_error_free(err.ptr)
	})

	return err
}


func NewContext() *Context {
	ctx := &Context {
		ptr: C.nickel_context_alloc(),
	}

	runtime.SetFinalizer(ctx, func(ctx *Context) {
		C.nickel_context_free(ctx.ptr)
	})

	return ctx
}

func (ctx *Context) EvalDeep(src string) (*Expr, error) {
	// This is a little silly, because eventually the Rust library converts
	// the null-terminated C string into a length-delimited Rust string.
	// We could avoid some extra copying by having the C API work with
	// length-delimited strings, but then it's a weird API for C users...
	csrc := C.CString(src)
	out_expr := new_expr()
	out_err := new_err()
	result := C.nickel_context_eval_deep(ctx.ptr, csrc, out_expr.ptr, out_err.ptr)
	C.free(unsafe.Pointer(csrc))

	if result == C.NICKEL_RESULT_OK {
		return out_expr, nil
	} else {
		return nil, out_err
	}
}

// ToRecord converts an Expr to a Record if it is one.
//
// Returns nil, false if expr is not a Nickel record.
func (expr *Expr) ToRecord() (*Record, bool) {
	if C.nickel_expr_is_record(expr.ptr) != 0 {
		ptr := C.nickel_expr_as_record(expr.ptr)
		return &Record {
			ptr: ptr,
			expr: expr,
		}, true
	} else {
		return nil, false
	}
}

func (expr *Expr) MarshalJSON() ([]byte, error) {
	out_err := new_err()
	out_string := C.nickel_string_alloc()
	defer C.nickel_string_free(out_string)
	
	result := C.nickel_expr_to_json(expr.ptr, out_string, out_err.ptr)
	if result == C.NICKEL_RESULT_ERR {
		return nil, out_err
	} else {
		var len C.uintptr_t
		var bytes *C.char
		C.nickel_string_data(out_string, &bytes, &len)

		// Copy the string data over, because it's owned on the Rust side
		// and will be freed with `out_string`.
		borrowedSlice := unsafe.Slice((*byte)(unsafe.Pointer(bytes)), int(len))
		slice := make([]byte, int(len))
		copy(slice, borrowedSlice)
		
		return slice, nil
	}
}

// ConvertTo converts an Expr to anything that can be unmarshaled from JSON.
func (expr *Expr) ConvertTo(target interface{}) error {
	data, err := expr.MarshalJSON()
	if err != nil {
		return err
	}

	return json.Unmarshal(data, target)
}

// Elements retrieves all of the record's elements, as a map.
//
// If the record was the result of lazy evaluation, it may have undefined
// fields. In that case, the returned map will have keys whose values are nil.
func (record *Record) Elements() map[string]*Expr {
	len := C.nickel_record_len(record.ptr)
	ret := make(map[string]*Expr)

	for i := C.uintptr_t(0); i < len; i++ {
		var key *C.char
		var key_len C.uintptr_t
		value := new_expr()

		has_value := C.nickel_record_key_value_by_index(record.ptr, C.uintptr_t(i), &key, &key_len, value.ptr)
		if has_value == 0 {
			value = nil
		}

		key_string := C.GoStringN(key, C.int(key_len))
		ret[key_string] = value
	}

	return ret
}
