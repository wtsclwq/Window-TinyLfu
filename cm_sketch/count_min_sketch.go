package cm_sketch

import (
	"hash"
	"math"
)

type CmSketch struct {
	rows        []cmRow
	hashHelpers []hash.Hash32
}

func NewCmSketch(counterNum int, hashHelpers []hash.Hash32) *CmSketch {
	cms := new(CmSketch)
	cms.rows = make([]cmRow, 0, len(hashHelpers))
	for i := 0; i < len(hashHelpers); i++ {
		cms.rows = append(cms.rows, newCmRow(uint32(counterNum)))
	}
	cms.hashHelpers = hashHelpers
	return cms
}

func (s *CmSketch) Increase(value string) {
	for i, r := range s.rows {
		rowIncrease(r, value, s.hashHelpers[i])
	}
}

func (s *CmSketch) Estimate(value string) uint8 {
	var min uint8 = math.MaxUint8
	for i, r := range s.rows {
		val := rowEstimate(r, value, s.hashHelpers[i])
		if val < min {
			min = val
		}
	}
	return min
}

func (s *CmSketch) Reset() {
	for _, r := range s.rows {
		r.reset()
	}
}

func (s *CmSketch) Clear() {
	for _, r := range s.rows {
		r.clear()
	}
}

func rowIncrease(row cmRow, value string, helper hash.Hash32) {
	helper.Reset()
	if _, err := helper.Write([]byte(value)); err != nil {
		panic(err)
	}
	hashed := helper.Sum32() % uint32(len(row))
	row.increase(hashed)
}

func rowEstimate(row cmRow, value string, helper hash.Hash32) uint8 {
	helper.Reset()
	if _, err := helper.Write([]byte(value)); err != nil {
		panic(err)
	}
	hashed := helper.Sum32() % uint32(len(row))
	return row.get(hashed)
}
