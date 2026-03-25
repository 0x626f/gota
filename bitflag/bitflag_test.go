package bitflag

import "testing"

const (
	FlagRead    BitFlag = 1 << 0 // 0001
	FlagWrite   BitFlag = 1 << 1 // 0010
	FlagExecute BitFlag = 1 << 2 // 0100
	FlagDelete  BitFlag = 1 << 3 // 1000
)

func TestBitFlag_Add(t *testing.T) {
	tests := []struct {
		name     string
		initial  BitFlag
		toAdd    BitFlag
		expected BitFlag
	}{
		{
			name:     "add single flag to empty",
			initial:  0,
			toAdd:    FlagRead,
			expected: FlagRead,
		},
		{
			name:     "add single flag to existing",
			initial:  FlagRead,
			toAdd:    FlagWrite,
			expected: FlagRead | FlagWrite,
		},
		{
			name:     "add multiple flags",
			initial:  FlagRead,
			toAdd:    FlagWrite | FlagExecute,
			expected: FlagRead | FlagWrite | FlagExecute,
		},
		{
			name:     "add already existing flag",
			initial:  FlagRead | FlagWrite,
			toAdd:    FlagRead,
			expected: FlagRead | FlagWrite,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bf := tt.initial
			bf.Add(tt.toAdd)
			if bf != tt.expected {
				t.Errorf("Add() = %b, want %b", bf, tt.expected)
			}
		})
	}
}

func TestBitFlag_Delete(t *testing.T) {
	tests := []struct {
		name     string
		initial  BitFlag
		toDelete BitFlag
		expected BitFlag
	}{
		{
			name:     "delete from empty",
			initial:  0,
			toDelete: FlagRead,
			expected: 0,
		},
		{
			name:     "delete single flag",
			initial:  FlagRead | FlagWrite,
			toDelete: FlagRead,
			expected: FlagWrite,
		},
		{
			name:     "delete multiple flags",
			initial:  FlagRead | FlagWrite | FlagExecute,
			toDelete: FlagRead | FlagWrite,
			expected: FlagExecute,
		},
		{
			name:     "delete non-existing flag",
			initial:  FlagRead,
			toDelete: FlagWrite,
			expected: FlagRead,
		},
		{
			name:     "delete all flags",
			initial:  FlagRead | FlagWrite | FlagExecute,
			toDelete: FlagRead | FlagWrite | FlagExecute,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bf := tt.initial
			bf.Delete(tt.toDelete)
			if bf != tt.expected {
				t.Errorf("Delete() = %b, want %b", bf, tt.expected)
			}
		})
	}
}

func TestBitFlag_Has(t *testing.T) {
	tests := []struct {
		name     string
		bitflag  BitFlag
		check    BitFlag
		expected bool
	}{
		{
			name:     "has single flag - true",
			bitflag:  FlagRead,
			check:    FlagRead,
			expected: true,
		},
		{
			name:     "has single flag - false",
			bitflag:  FlagRead,
			check:    FlagWrite,
			expected: false,
		},
		{
			name:     "has multiple flags - all present",
			bitflag:  FlagRead | FlagWrite | FlagExecute,
			check:    FlagRead | FlagWrite,
			expected: true,
		},
		{
			name:     "has multiple flags - some missing",
			bitflag:  FlagRead | FlagWrite,
			check:    FlagRead | FlagExecute,
			expected: true,
		},
		{
			name:     "empty bitflag",
			bitflag:  0,
			check:    FlagRead,
			expected: false,
		},
		{
			name:     "check zero flag",
			bitflag:  FlagRead,
			check:    FlagRead,
			expected: true,
		},
		{
			name:     "both empty",
			bitflag:  0,
			check:    0,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.bitflag.Has(tt.check)
			if result != tt.expected {
				t.Errorf("Has(%b) = %v, want %v (bitflag: %b)", tt.check, result, tt.expected, tt.bitflag)
			}
		})
	}
}

func TestBitFlag_Combined(t *testing.T) {
	var bf BitFlag

	// Start with empty
	if bf.Has(FlagRead) {
		t.Error("Empty BitFlag should not have FlagRead")
	}

	// Add read permission
	bf.Add(FlagRead)
	if !bf.Has(FlagRead) {
		t.Error("BitFlag should have FlagRead after Add")
	}

	// Add write permission
	bf.Add(FlagWrite)
	if !bf.Has(FlagRead) || !bf.Has(FlagWrite) {
		t.Error("BitFlag should have both FlagRead and FlagWrite")
	}

	// Add execute permission
	bf.Add(FlagExecute)
	if !bf.Has(FlagRead | FlagWrite | FlagExecute) {
		t.Error("BitFlag should have FlagRead, FlagWrite, and FlagExecute")
	}

	// Delete write permission
	bf.Delete(FlagWrite)
	if !bf.Has(FlagRead) || bf.Has(FlagWrite) || !bf.Has(FlagExecute) {
		t.Error("BitFlag should have FlagRead and FlagExecute, but not FlagWrite")
	}

	// Delete all permissions
	bf.Delete(FlagRead | FlagExecute)
	if bf != 0 {
		t.Error("BitFlag should be empty after deleting all flags")
	}
}

func BenchmarkBitFlag_Add(b *testing.B) {
	var bf BitFlag
	for i := 0; i < b.N; i++ {
		bf.Add(FlagRead)
	}
}

func BenchmarkBitFlag_Delete(b *testing.B) {
	var bf BitFlag = FlagRead | FlagWrite | FlagExecute
	for i := 0; i < b.N; i++ {
		bf.Delete(FlagRead)
		bf.Add(FlagRead)
	}
}

func BenchmarkBitFlag_Has(b *testing.B) {
	bf := FlagRead | FlagWrite | FlagExecute
	for i := 0; i < b.N; i++ {
		_ = bf.Has(FlagWrite)
	}
}
