// Package bitflag provides a BitFlag type for managing sets of boolean flags
// packed into a single unsigned integer. Individual flags are defined as
// BitFlag constants using powers of two (e.g. 1 << iota).
package bitflag

// BitFlag represents a collection of bit flags stored as an unsigned integer.
// Each bit position can represent a different boolean flag.
type BitFlag uint

// Add sets the specified flag bits in the BitFlag.
// It performs a bitwise OR operation to enable the given flags.
func (bitflag *BitFlag) Add(flag BitFlag) {
	*bitflag |= flag
}

// Delete clears the specified flag bits from the BitFlag.
// It performs a bitwise AND NOT operation to disable the given flags.
func (bitflag *BitFlag) Delete(flag BitFlag) {
	*bitflag &^= flag
}

// Has checks if all the specified flag bits are set in the BitFlag.
// It returns true if all bits in the flag parameter are set, false otherwise.
func (bitflag *BitFlag) Has(flag BitFlag) bool {
	return *bitflag&flag != 0
}
