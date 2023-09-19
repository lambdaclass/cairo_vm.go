package hints

import (
	"github.com/ebfe/keccak"
	. "github.com/lambdaclass/cairo-vm.go/pkg/hints/hint_utils"
	. "github.com/lambdaclass/cairo-vm.go/pkg/lambdaworks"
	. "github.com/lambdaclass/cairo-vm.go/pkg/types"
	. "github.com/lambdaclass/cairo-vm.go/pkg/vm"
	. "github.com/lambdaclass/cairo-vm.go/pkg/vm/memory"
	"github.com/pkg/errors"
)

func unsafeKeccak(ids IdsManager, vm *VirtualMachine, scopes ExecutionScopes) error {
	// Fetch ids variable
	lengthFelt, err := ids.GetFelt("length", vm)
	if err != nil {
		return err
	}
	length, err := lengthFelt.ToU64()
	if err != nil {
		return err
	}
	data, err := ids.GetRelocatable("data", vm)
	if err != nil {
		return err
	}
	// Check __keccak_max_size if available
	keccakMaxSizeAny, err := scopes.Get("__keccak_max_size")
	if err == nil {
		keccakMaxSize, ok := keccakMaxSizeAny.(uint64)
		if ok {
			if length > keccakMaxSize {
				return errors.Errorf("unsafe_keccak() can only be used with length<=%d. Got: length=%d", keccakMaxSize, length)
			}
		}
	}
	keccakInput := make([]byte, 0)
	for byteIdx, wordIdx := 0, 0; byteIdx < int(length); byteIdx, wordIdx = byteIdx+16, wordIdx+1 {
		wordAddr := data.AddUint(uint(wordIdx))
		word, err := vm.Segments.Memory.GetFelt(wordAddr)
		if err != nil {
			return err
		}
		nBytes := int(length) - byteIdx
		if nBytes > 16 {
			nBytes = 16
		}

		if int(word.Bits()) > 8*nBytes {
			return errors.Errorf("Invalid word size: %s", word.ToHexString())
		}

		start := 32 - nBytes
		keccakInput = append(keccakInput, word.ToBeBytes()[start:]...)

	}

	hasher := keccak.New256()
	hasher.Write(keccakInput)
	resBytes := hasher.Sum(nil)

	highBytes := append([]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, resBytes[:16]...)
	lowBytes := append([]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, resBytes[16:32]...)

	high := FeltFromBeBytes((*[32]byte)(highBytes))
	low := FeltFromBeBytes((*[32]byte)(lowBytes))

	err = ids.Insert("high", NewMaybeRelocatableFelt(high), vm)
	if err != nil {
		return err
	}
	return ids.Insert("low", NewMaybeRelocatableFelt(low), vm)
}

func unsafeKeccakFinalize(ids IdsManager, vm *VirtualMachine, scopes ExecutionScopes) error {
	// Fetch ids variables
	startPtr, err := ids.GetStructFieldRelocatable("keccak_state", 0, vm)
	if err != nil {
		return err
	}
	endPtr, err := ids.GetStructFieldRelocatable("keccak_state", 1, vm)
	if err != nil {
		return err
	}
	n_elems, err := endPtr.Sub(startPtr)
	if err != nil {
		return err
	}
}
