package lambdaworks

/*
#cgo LDFLAGS: pkg/lambdaworks/lib/liblambdaworks.a -ldl
#include "lib/lambdaworks.h"
*/
import "C"
import "errors"

// Go representation of a single limb (unsigned integer with 64 bits).
type Limb C.limb_t

// Go representation of a 256 bit prime field element (felt).
type Felt struct {
	limbs [4]Limb
}

// Converts a Go Felt to a C felt_t.
func (f Felt) toC() C.felt_t {
	var result C.felt_t
	for i, limb := range f.limbs {
		result[i] = C.limb_t(limb)
	}
	return result
}

// Converts a C felt_t to a Go Felt.
func fromC(result C.felt_t) Felt {
	var limbs [4]Limb
	for i, limb := range result {
		limbs[i] = Limb(limb)
	}
	return Felt{limbs: limbs}
}

// Gets a Felt representing the "value" number, in Montgomery format.
func From(value uint64) Felt {
	var result C.felt_t
	C.from(&result[0], C.uint64_t(value))
	return fromC(result)
}

// Gets a Felt representing 0.
func Zero() Felt {
	var result C.felt_t
	C.zero(&result[0])
	return fromC(result)
}

// Gets a Felt representing 1.
func One() Felt {
	var result C.felt_t
	C.one(&result[0])
	return fromC(result)
}

// Writes the result variable with the sum of a and b felts.
func Add(a, b Felt) Felt {
	var result C.felt_t
	var a_c C.felt_t = a.toC()
	var b_c C.felt_t = b.toC()
	C.add(&a_c[0], &b_c[0], &result[0])
	return fromC(result)
}

// Writes the result variable with a - b.
func Sub(a, b Felt) Felt {
	var result C.felt_t
	var a_c C.felt_t = a.toC()
	var b_c C.felt_t = b.toC()
	C.sub(&a_c[0], &b_c[0], &result[0])
	return fromC(result)
}

// Writes the result variable with a * b.
func Mul(a, b Felt) Felt {
	var result C.felt_t
	var a_c C.felt_t = a.toC()
	var b_c C.felt_t = b.toC()
	C.mul(&a_c[0], &b_c[0], &result[0])
	return fromC(result)
}

// Writes the result variable with a / b.
func Div(a, b Felt) Felt {
	var result C.felt_t
	var a_c C.felt_t = a.toC()
	var b_c C.felt_t = b.toC()
	C.div(&a_c[0], &b_c[0], &result[0])
	return fromC(result)
}

// Returns the result of the number function.
func Number() int {
	return int(C.number())
}

// turns a felt to usize
func (felt Felt) ToU64() (uint64, error) {
	if felt.limbs[0] == 0 && felt.limbs[1] == 0 && felt.limbs[2] == 0 {
		return uint64(felt.limbs[3]), nil
	} else {
		return 0, errors.New("Cannot convert felt to u64")
	}
}
