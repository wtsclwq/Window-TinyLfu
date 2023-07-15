package cm_sketch

import (
	"hash"
	"math"

	"github.com/spaolacci/murmur3"
)

type CmSketch struct {
	rows        []cmRow
	hashHelpers []hash.Hash64
}

func NewCmSketch(counterNum uint64, rowNum int) *CmSketch {
	cms := new(CmSketch)
	cms.rows = make([]cmRow, 0, rowNum)
	cms.hashHelpers = make([]hash.Hash64, 0, rowNum)
	var initSeed uint32 = 64
	for i := 0; i < rowNum; i++ {
		cms.rows = append(cms.rows, newCmRow(counterNum))
		cms.hashHelpers = append(cms.hashHelpers, murmur3.New64WithSeed(initSeed))
	}
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

func rowIncrease(row cmRow, value string, helper hash.Hash64) {
	helper.Reset()
	if _, err := helper.Write([]byte(value)); err != nil {
		panic(err)
	}
	hashed := helper.Sum64() % uint64(len(row))
	row.increase(hashed)
}

func rowEstimate(row cmRow, value string, helper hash.Hash64) uint8 {
	helper.Reset()
	if _, err := helper.Write([]byte(value)); err != nil {
		panic(err)
	}
	hashed := helper.Sum64() % uint64(len(row))
	return row.get(hashed)
}
