package nickel

/*
#cgo CFLAGS: -I./include

#include <nickel_lang.h>
*/
import "C"

import "runtime"

// VirtualMachine can be used to further evaluate lazy expressions.
//
// See EvalShallow for more.
type VirtualMachine struct {
	ptr *C.nickel_virtual_machine
}

func new_vm() *VirtualMachine {
	vm := &VirtualMachine{
		ptr: C.nickel_virtual_machine_alloc(),
	}

	runtime.SetFinalizer(vm, func(vm *VirtualMachine) {
		C.nickel_virtual_machine_free(vm.ptr)
	})

	return vm
}

// EvalShallow evaluates an expression shallowly.
//
// This has no effect if the expression is already evaluated.
//
// The result of this evaluation is a null, bool, number, string,
// enum, record, or array. In case it's a record, array, or enum
// variant, the payload (record values, array elements, or enum
// payloads) will be left unevaluated.
func (vm *VirtualMachine) EvalShallow(expr *Expr) (*Expr, error) {
	out_expr := new_expr()
	out_err := new_err()

	result := C.nickel_virtual_machine_eval_shallow(vm.ptr, expr.ptr, out_expr.ptr, out_err.ptr)

	if result == C.NICKEL_RESULT_OK {
		return out_expr, nil
	} else {
		return nil, out_err
	}
}
