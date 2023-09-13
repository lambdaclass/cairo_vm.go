package dict_manager

import (
	. "github.com/lambdaclass/cairo-vm.go/pkg/vm"
	. "github.com/lambdaclass/cairo-vm.go/pkg/vm/memory"
	"github.com/pkg/errors"
)

// Manages dictionaries in a Cairo program.
// Uses the segment index to associate the corresponding go dict with the Cairo dict.
type DictManager struct {
	trackers map[int]DictTracker
}

func NewDictManager() DictManager {
	return DictManager{
		trackers: make(map[int]DictTracker),
	}
}

func (d *DictManager) NewDictionary(dict *map[MaybeRelocatable]MaybeRelocatable, vm *VirtualMachine) Relocatable {
	base := vm.Segments.AddSegment()
	d.trackers[base.SegmentIndex] = NewDictTrackerForDictionary(base, dict)
	return base
}

func (d *DictManager) NewDefaultDictionary(defaultValue *MaybeRelocatable, dict *map[MaybeRelocatable]MaybeRelocatable, vm *VirtualMachine) Relocatable {
	base := vm.Segments.AddSegment()
	d.trackers[base.SegmentIndex] = NewDictTrackerForDefaultDictionary(base, defaultValue, dict)
	return base
}

func (d *DictManager) GetTracker(dict_ptr Relocatable) (*DictTracker, error) {
	tracker, ok := d.trackers[dict_ptr.SegmentIndex]
	if !ok {
		return nil, errors.Errorf("Dict Error: No dict tracker found for segment %d", dict_ptr.SegmentIndex)
	}
	if tracker.currentPtr != dict_ptr {
		return nil, errors.Errorf("Dict Error: Wrong dict pointer supplied. Got %v, expected %v", dict_ptr, tracker.currentPtr)
	}
	return &tracker, nil
}

// Tracks the go dict associated with a Cairo dict.
type DictTracker struct {
	data Dictionary
	// Pointer to the first unused position in the dict segment.
	currentPtr Relocatable
}

func NewDictTrackerForDictionary(base Relocatable, dict *map[MaybeRelocatable]MaybeRelocatable) DictTracker {
	return DictTracker{
		data:       NewDictionary(dict),
		currentPtr: base,
	}
}

func NewDictTrackerForDefaultDictionary(base Relocatable, defaultValue *MaybeRelocatable, dict *map[MaybeRelocatable]MaybeRelocatable) DictTracker {
	return DictTracker{
		data:       NewDefaultDictionary(defaultValue, dict),
		currentPtr: base,
	}
}

func (d *DictTracker) CopyDictionary() map[MaybeRelocatable]MaybeRelocatable {
	return d.data.dict
}

func (d *DictTracker) GetValue(key *MaybeRelocatable) *MaybeRelocatable {
	return d.data.Get(key)
}

func (d *DictTracker) InsertValue(key *MaybeRelocatable, val *MaybeRelocatable) {
	d.data.Insert(key, val)
}

type Dictionary struct {
	dict         map[MaybeRelocatable]MaybeRelocatable
	defaultValue *MaybeRelocatable
}

func NewDefaultDictionary(defaultValue *MaybeRelocatable, dict *map[MaybeRelocatable]MaybeRelocatable) Dictionary {
	return Dictionary{
		dict:         *dict,
		defaultValue: defaultValue,
	}
}

func NewDictionary(dict *map[MaybeRelocatable]MaybeRelocatable) Dictionary {
	return Dictionary{
		dict:         *dict,
		defaultValue: nil,
	}
}

func (d *Dictionary) Get(key *MaybeRelocatable) *MaybeRelocatable {
	val, ok := d.dict[*key]
	if ok {
		return &val
	}
	if d.defaultValue != nil {
		d.dict[*key] = *d.defaultValue
	}
	return d.defaultValue
}

func (d *Dictionary) Insert(key *MaybeRelocatable, val *MaybeRelocatable) {
	d.dict[*key] = *val
}
