package graph

import "github.com/0x626f/gota/bitflag"

// IVertex is implemented by any type that can act as a graph vertex.
// Key returns the comparable identity used to address the vertex in a topology.
type IVertex[Key comparable] interface {
	Key() Key
}

// Weight is the numeric type used for edge weights and heuristic values.
type Weight = float64

// IWeightedEdge is implemented by edge types that carry a numeric weight,
// enabling automatic cost extraction during weighted traversal.
type IWeightedEdge interface {
	Weight() Weight
}

type bitRow = []byte

// Feature is a flag that configures topology behaviour.
type Feature bitflag.BitFlag

const (
	None     Feature = iota // default — undirected, cycles allowed
	Directed                // edges are one-way
	Acyclic                 // cycle-creating edges are rejected
)

// Features is a convenience constructor that returns a slice of Feature values.
func Features(features ...Feature) []Feature {
	return features
}

// FeatureStorage is a compact bit-flag store for topology features.
type FeatureStorage struct {
	features bitflag.BitFlag
}

// SetFeature enables the given feature.
func (storage *FeatureStorage) SetFeature(feature Feature) {
	storage.features.Add(bitflag.BitFlag(feature))
}

// HasFeature reports whether the given feature is enabled.
func (storage *FeatureStorage) HasFeature(feature Feature) bool {
	return storage.features.Has(bitflag.BitFlag(feature))
}
