package memory

import (
	"errors"
	"fmt"

	"github.com/lambdaclass/cairo-vm.go/pkg/lambdaworks"
)

// Relocatable in the Cairo VM represents an address
// in some memory segment. When the VM finishes running,
// these values are replaced by real memory addresses,
// represented by a field element.
type Relocatable struct {
	SegmentIndex int
	Offset       uint
}

// Creates a new Relocatable struct with the specified segment index
// and offset.
func NewRelocatable(segment_idx int, offset uint) Relocatable {
	return Relocatable{segment_idx, offset}
}

// Adds a Felt value to a Relocatable
// Fails if the new offset exceeds the size of a uint
func (r *Relocatable) AddFelt(other Int) (Relocatable, error) {
	felt_offset := lambdaworks.FeltFromUint64(uint64(r.Offset))
	new_offset := felt_offset.Add(other.Felt)
	res_offset, err := new_offset.ToU64()
	if err != nil {
		return Relocatable{}, err
	}
	return NewRelocatable(r.SegmentIndex, uint(res_offset)), nil

}

// Performs additions if other contains a Felt value, fails otherwise
func (r *Relocatable) AddMaybeRelocatable(other MaybeRelocatable) (Relocatable, error) {
	felt, ok := other.GetInt()
	if !ok {
		return Relocatable{}, errors.New("Can't add two relocatable values")
	}
	return r.AddFelt(felt)
}

func (r *Relocatable) RelocateAddress(relocationTable *[]uint) uint {
	return (*relocationTable)[r.SegmentIndex] + r.Offset
}

func (r *Relocatable) IsEqual(r1 *Relocatable) bool {
	return (r.SegmentIndex == r1.SegmentIndex && r.Offset == r1.Offset)
}

func (relocatable *Relocatable) SubUint(other uint) (Relocatable, error) {
	if relocatable.Offset < other {
		return NewRelocatable(0, 0), &SubReloctableError{Msg: "RelocatableSubUsizeNegOffset"}
	} else {
		new_offset := relocatable.Offset - other
		return NewRelocatable(relocatable.SegmentIndex, new_offset), nil
	}
}

func (relocatable *Relocatable) AddUint(other uint) (Relocatable, error) {
	new_offset := relocatable.Offset + other
	return NewRelocatable(relocatable.SegmentIndex, new_offset), nil
}

// Int in the Cairo VM represents a value in memory that
// is not an address.
type Int struct {
	Felt lambdaworks.Felt
}

// MaybeRelocatable is the type of the memory cells in the Cairo
// VM. For now, `inner` will hold any type but it should be
// instantiated only with `Relocatable` or `Int` types.
// We should analyze better alternatives to this.
type MaybeRelocatable struct {
	inner any
}

// Creates a new MaybeRelocatable with an Int inner value
func NewMaybeRelocatableInt(felt lambdaworks.Felt) *MaybeRelocatable {
	return &MaybeRelocatable{inner: Int{felt}}
}

// Creates a new MaybeRelocatable with a Relocatable inner value
func NewMaybeRelocatableRelocatable(relocatable Relocatable) *MaybeRelocatable {
	return &MaybeRelocatable{inner: relocatable}
}

// If m is Int, returns the inner value + true, if not, returns zero + false
func (m *MaybeRelocatable) GetInt() (Int, bool) {
	int, is_type := m.inner.(Int)
	return int, is_type
}

// If m is Relocatable, returns the inner value + true, if not, returns zero + false
func (m *MaybeRelocatable) GetRelocatable() (Relocatable, bool) {
	rel, is_type := m.inner.(Relocatable)
	return rel, is_type
}

func (m *MaybeRelocatable) IsZero() bool {
	felt, is_int := m.GetInt()
	return is_int && felt.Felt == lambdaworks.FeltFromUint64(0)
}

// Turns a MaybeRelocatable into a Felt252 value.
// If the inner value is an Int, it will extract the Felt252 value from it.
// If the inner value is a Relocatable, it will relocate it according to the relocation_table
func (m *MaybeRelocatable) RelocateValue(relocationTable *[]uint) (lambdaworks.Felt, error) {
	inner_int, ok := m.GetInt()
	if ok {
		return inner_int.Felt, nil
	}

	inner_relocatable, ok := m.GetRelocatable()
	if ok {
		return lambdaworks.FeltFromUint64(uint64(inner_relocatable.RelocateAddress(relocationTable))), nil
	}

	return lambdaworks.FeltFromUint64(0), errors.New(fmt.Sprintf("Unexpected type %T", m.inner))
}

func (m *MaybeRelocatable) IsEqual(m1 *MaybeRelocatable) bool {
	a, a_type := m.GetInt()
	b, b_type := m1.GetInt()
	if a_type == b_type {
		if a_type {
			return a == b
		} else {
			a, _ := m.GetRelocatable()
			b, _ := m1.GetRelocatable()
			return a.IsEqual(&b)
		}
	} else {
		return false
	}
}

func (m MaybeRelocatable) AddMaybeRelocatable(other MaybeRelocatable) (MaybeRelocatable, error) {
	// check if they are felt
	m_int, m_is_int := m.GetInt()
	other_int, other_is_int := other.GetInt()

	if m_is_int && other_is_int {
		result := NewMaybeRelocatableInt(m_int.Felt.Add(other_int.Felt))
		return *result, nil
	}

	// check if one is relocatable and the other int
	m_rel, is_rel_m := m.GetRelocatable()
	other_rel, is_rel_other := other.GetRelocatable()

	if is_rel_m && !is_rel_other {
		other_felt, _ := other.GetInt()
		felt, err := m_rel.AddFelt(other_felt)
		if err != nil {
			return MaybeRelocatable{}, nil
		}
		return *NewMaybeRelocatableRelocatable(felt), nil

	} else if !is_rel_m && is_rel_other {

		m_felt, _ := m.GetInt()
		felt, err := other_rel.AddFelt(m_felt)
		if err != nil {
			return MaybeRelocatable{}, err
		}
		return *NewMaybeRelocatableRelocatable(felt), nil
	} else {
		return MaybeRelocatable{}, errors.New("RelocatableAdd")
	}
}
