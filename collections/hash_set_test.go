package collections

import (
	"testing"
)

func TestHashSet_Add(t *testing.T) {
	set := NewHashSet[int]()
	err := set.Add(1)
	if err != nil {
		t.Errorf("Add() error = %v, wantErr %v", err, nil)
	}
	if !set.Contains(1) {
		t.Errorf("Add() did not add element to set")
	}
}

func TestHashSet_AddAll(t *testing.T) {
	set1 := NewHashSet[int]()
	set1.Add(1)
	set1.Add(2)

	set2 := NewHashSet[int]()
	set2.Add(3)
	set2.Add(4)

	err := set1.AddAll(set2)
	if err != nil {
		t.Errorf("AddAll() error = %v, wantErr %v", err, nil)
	}
	if !set1.Contains(3) || !set1.Contains(4) {
		t.Errorf("AddAll() did not add all elements to set")
	}
}

func TestHashSet_Clear(t *testing.T) {
	set := NewHashSet[int]()
	set.Add(1)
	set.Clear()
	if set.Size() != 0 {
		t.Errorf("Clear() did not remove all elements from set")
	}
}

func TestHashSet_Contains(t *testing.T) {
	set := NewHashSet[int]()
	set.Add(1)
	if !set.Contains(1) {
		t.Errorf("Contains() = false, want true")
	}
	if set.Contains(2) {
		t.Errorf("Contains() = true, want false")
	}
}

func TestHashSet_Remove(t *testing.T) {
	set := NewHashSet[int]()
	set.Add(1)
	removed := set.Remove(1)
	if !removed {
		t.Errorf("Remove() = false, want true")
	}
	if set.Contains(1) {
		t.Errorf("Remove() did not remove element from set")
	}
}

func TestHashSet_Size(t *testing.T) {
	set := NewHashSet[int]()
	set.Add(1)
	set.Add(2)
	if set.Size() != 2 {
		t.Errorf("Size() = %v, want %v", set.Size(), 2)
	}
}

func TestHashSet_ContainsAll(t *testing.T) {
	set1 := NewHashSet[int]()
	set1.Add(1)
	set1.Add(2)

	set2 := NewHashSet[int]()
	set2.Add(1)

	if !set1.ContainsAll(set2) {
		t.Errorf("ContainsAll() = false, want true")
	}
}

func TestHashSet_IsEmpty(t *testing.T) {
	set := NewHashSet[int]()
	if !set.IsEmpty() {
		t.Errorf("IsEmpty() = false, want true")
	}
	set.Add(1)
	if set.IsEmpty() {
		t.Errorf("IsEmpty() = true, want false")
	}
}

func TestHashSet_Iterator(t *testing.T) {
	set := NewHashSet[int]()
	set.Add(1)
	set.Add(2)
	it := set.Iterator()
	count := 0
	for it.HasNext() {
		it.Next()
		count++
	}
	if count != 2 {
		t.Errorf("Iterator() = %v, want %v", count, 2)
	}
}

func TestHashSet_String(t *testing.T) {
	set := NewHashSet[int]()
	set.Add(1)
	set.Add(2)
	str := set.String()
	expected := "{1, 2}"
	if len(str) != len(expected) {
		t.Errorf("String() = %v, want %v", str, expected)
	}
}

func TestHashSet_Union(t *testing.T) {
	set1 := NewHashSet[int]()
	set1.Add(1)
	set1.Add(2)

	set2 := NewHashSet[int]()
	set2.Add(3)
	set2.Add(4)

	union := set1.Union(set2)
	if !union.Contains(1) || !union.Contains(2) || !union.Contains(3) || !union.Contains(4) {
		t.Errorf("Union() did not contain all elements from both sets")
	}
}

func TestHashSet_Intersection(t *testing.T) {
	set1 := NewHashSet[int]()
	set1.Add(1)
	set1.Add(2)

	set2 := NewHashSet[int]()
	set2.Add(2)
	set2.Add(3)

	intersection := set1.Intersection(set2)
	if !intersection.Contains(2) || intersection.Contains(1) || intersection.Contains(3) {
		t.Errorf("Intersection() did not contain only common elements")
	}
}

func TestHashSet_Difference(t *testing.T) {
	set1 := NewHashSet[int]()
	set1.Add(1)
	set1.Add(2)

	set2 := NewHashSet[int]()
	set2.Add(2)
	set2.Add(3)

	difference := set1.Difference(set2)
	if !difference.Contains(1) || difference.Contains(2) || difference.Contains(3) {
		t.Errorf("Difference() did not contain only unique elements")
	}
}

func TestHashSet_IsSubset(t *testing.T) {
	set1 := NewHashSet[int]()
	set1.Add(1)
	set1.Add(2)

	set2 := NewHashSet[int]()
	set2.Add(1)

	if !set2.IsSubset(set1) {
		t.Errorf("IsSubset() = false, want true")
	}
	if set1.IsSubset(set2) {
		t.Errorf("IsSubset() = true, want false")
	}
}

func TestHashSet_IsSuperset(t *testing.T) {
	set1 := NewHashSet[int]()
	set1.Add(1)
	set1.Add(2)

	set2 := NewHashSet[int]()
	set2.Add(1)

	if !set1.IsSuperset(set2) {
		t.Errorf("IsSuperset() = false, want true")
	}
	if set2.IsSuperset(set1) {
		t.Errorf("IsSuperset() = true, want false")
	}
}

func TestHashSet_IsProperSubset(t *testing.T) {
	set1 := NewHashSet[int]()
	set1.Add(1)
	set1.Add(2)

	set2 := NewHashSet[int]()
	set2.Add(1)

	if !set2.IsProperSubset(set1) {
		t.Errorf("IsProperSubset() = false, want true")
	}
	if set1.IsProperSubset(set2) {
		t.Errorf("IsProperSubset() = true, want false")
	}
}

func TestHashSet_IsProperSuperset(t *testing.T) {
	set1 := NewHashSet[int]()
	set1.Add(1)
	set1.Add(2)

	set2 := NewHashSet[int]()
	set2.Add(1)

	if !set1.IsProperSuperset(set2) {
		t.Errorf("IsProperSuperset() = false, want true")
	}
	if set2.IsProperSuperset(set1) {
		t.Errorf("IsProperSuperset() = true, want false")
	}
}

func TestHashSet_IsDisjoint(t *testing.T) {
	set1 := NewHashSet[int]()
	set1.Add(1)
	set1.Add(2)

	set2 := NewHashSet[int]()
	set2.Add(3)
	set2.Add(4)

	if !set1.IsDisjoint(set2) {
		t.Errorf("IsDisjoint() = false, want true")
	}

	set2.Add(2)
	if set1.IsDisjoint(set2) {
		t.Errorf("IsDisjoint() = true, want false")
	}
}

func TestHashSet_SymmetricDifference(t *testing.T) {
	set1 := NewHashSet[int]()
	set1.Add(1)
	set1.Add(2)

	set2 := NewHashSet[int]()
	set2.Add(2)
	set2.Add(3)

	symmetricDifference := set1.SymmetricDifference(set2)
	if !symmetricDifference.Contains(1) || !symmetricDifference.Contains(3) || symmetricDifference.Contains(2) {
		t.Errorf("SymmetricDifference() did not contain only unique elements")
	}
}
func TestSyncSet_Add(t *testing.T) {
	set := NewSyncSet[int]()
	err := set.Add(1)
	if err != nil {
		t.Errorf("Add() error = %v, wantErr %v", err, nil)
	}
	if !set.Contains(1) {
		t.Errorf("Add() did not add element to set")
	}
}

func TestSyncSet_AddAll(t *testing.T) {
	set1 := NewSyncSet[int]()
	set1.Add(1)
	set1.Add(2)

	set2 := NewSyncSet[int]()
	set2.Add(3)
	set2.Add(4)

	err := set1.AddAll(set2)
	if err != nil {
		t.Errorf("AddAll() error = %v, wantErr %v", err, nil)
	}
	if !set1.Contains(3) || !set1.Contains(4) {
		t.Errorf("AddAll() did not add all elements to set")
	}
}

func TestSyncSet_Clear(t *testing.T) {
	set := NewSyncSet[int]()
	set.Add(1)
	set.Clear()
	if set.Size() != 0 {
		t.Errorf("Clear() did not remove all elements from set")
	}
}

func TestSyncSet_Contains(t *testing.T) {
	set := NewSyncSet[int]()
	set.Add(1)
	if !set.Contains(1) {
		t.Errorf("Contains() = false, want true")
	}
	if set.Contains(2) {
		t.Errorf("Contains() = true, want false")
	}
}

func TestSyncSet_Remove(t *testing.T) {
	set := NewSyncSet[int]()
	set.Add(1)
	removed := set.Remove(1)
	if !removed {
		t.Errorf("Remove() = false, want true")
	}
	if set.Contains(1) {
		t.Errorf("Remove() did not remove element from set")
	}
}

func TestSyncSet_Size(t *testing.T) {
	set := NewSyncSet[int]()
	set.Add(1)
	set.Add(2)
	if set.Size() != 2 {
		t.Errorf("Size() = %v, want %v", set.Size(), 2)
	}
}

func TestSyncSet_IsEmpty(t *testing.T) {
	set := NewSyncSet[int]()
	if !set.IsEmpty() {
		t.Errorf("IsEmpty() = false, want true")
	}
	set.Add(1)
	if set.IsEmpty() {
		t.Errorf("IsEmpty() = true, want false")
	}
}

func TestSyncSet_Iterator(t *testing.T) {
	set := NewSyncSet[int]()
	set.Add(1)
	set.Add(2)
	it := set.Iterator()
	count := 0
	for it.HasNext() {
		it.Next()
		count++
	}
	if count != 2 {
		t.Errorf("Iterator() = %v, want %v", count, 2)
	}
}

func TestSyncSet_String(t *testing.T) {
	set := NewSyncSet[int]()
	set.Add(1)
	set.Add(2)
	str := set.String()
	expected := "{1, 2}"
	if len(str) != len(expected) {
		t.Errorf("String() = %v, want %v", str, expected)
	}
}

func TestSyncSet_Union(t *testing.T) {
	set1 := NewSyncSet[int]()
	set1.Add(1)
	set1.Add(2)

	set2 := NewSyncSet[int]()
	set2.Add(3)
	set2.Add(4)

	union := set1.Union(set2)
	if !union.Contains(1) || !union.Contains(2) || !union.Contains(3) || !union.Contains(4) {
		t.Errorf("Union() did not contain all elements from both sets")
	}
}

func TestSyncSet_Intersection(t *testing.T) {
	set1 := NewSyncSet[int]()
	set1.Add(1)
	set1.Add(2)

	set2 := NewSyncSet[int]()
	set2.Add(2)
	set2.Add(3)

	intersection := set1.Intersection(set2)
	if !intersection.Contains(2) || intersection.Contains(1) || intersection.Contains(3) {
		t.Errorf("Intersection() did not contain only common elements")
	}
}

func TestSyncSet_Difference(t *testing.T) {
	set1 := NewSyncSet[int]()
	set1.Add(1)
	set1.Add(2)

	set2 := NewSyncSet[int]()
	set2.Add(2)
	set2.Add(3)

	difference := set1.Difference(set2)
	if !difference.Contains(1) || difference.Contains(2) || difference.Contains(3) {
		t.Errorf("Difference() did not contain only unique elements")
	}
}

func TestSyncSet_IsSubset(t *testing.T) {
	set1 := NewSyncSet[int]()
	set1.Add(1)
	set1.Add(2)

	set2 := NewSyncSet[int]()
	set2.Add(1)

	if !set2.IsSubset(set1) {
		t.Errorf("IsSubset() = false, want true")
	}
	if set1.IsSubset(set2) {
		t.Errorf("IsSubset() = true, want false")
	}
}

func TestSyncSet_IsSuperset(t *testing.T) {
	set1 := NewSyncSet[int]()
	set1.Add(1)
	set1.Add(2)

	set2 := NewSyncSet[int]()
	set2.Add(1)

	if !set1.IsSuperset(set2) {
		t.Errorf("IsSuperset() = false, want true")
	}
	if set2.IsSuperset(set1) {
		t.Errorf("IsSuperset() = true, want false")
	}
}

func TestSyncSet_IsProperSubset(t *testing.T) {
	set1 := NewSyncSet[int]()
	set1.Add(1)
	set1.Add(2)

	set2 := NewSyncSet[int]()
	set2.Add(1)

	if !set2.IsProperSubset(set1) {
		t.Errorf("IsProperSubset() = false, want true")
	}
	if set1.IsProperSubset(set2) {
		t.Errorf("IsProperSubset() = true, want false")
	}
}

func TestSyncSet_IsProperSuperset(t *testing.T) {
	set1 := NewSyncSet[int]()
	set1.Add(1)
	set1.Add(2)

	set2 := NewSyncSet[int]()
	set2.Add(1)

	if !set1.IsProperSuperset(set2) {
		t.Errorf("IsProperSuperset() = false, want true")
	}
	if set2.IsProperSuperset(set1) {
		t.Errorf("IsProperSuperset() = true, want false")
	}
}

func TestSyncSet_IsDisjoint(t *testing.T) {
	set1 := NewSyncSet[int]()
	set1.Add(1)
	set1.Add(2)

	set2 := NewSyncSet[int]()
	set2.Add(3)
	set2.Add(4)

	if !set1.IsDisjoint(set2) {
		t.Errorf("IsDisjoint() = false, want true")
	}

	set2.Add(2)
	if set1.IsDisjoint(set2) {
		t.Errorf("IsDisjoint() = true, want false")
	}
}

func TestSyncSet_SymmetricDifference(t *testing.T) {
	set1 := NewSyncSet[int]()
	set1.Add(1)
	set1.Add(2)

	set2 := NewSyncSet[int]()
	set2.Add(2)
	set2.Add(3)

	symmetricDifference := set1.SymmetricDifference(set2)
	if !symmetricDifference.Contains(1) || !symmetricDifference.Contains(3) || symmetricDifference.Contains(2) {
		t.Errorf("SymmetricDifference() did not contain only unique elements")
	}
}
