package types

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func testRange(rangeStr string, expectedMin SequenceNumber, expectedMax SequenceNumber, expectedErr error) {
	testName := fmt.Sprintf("should interpret %s", rangeStr)
	if expectedErr != nil {
		testName = fmt.Sprintf("should throw an error attempting to interpret %s", rangeStr)
	}

	It(testName, func() {
		rng, err := InterpretMessageRange(rangeStr)
		Expect(rng.Min).To(Equal(expectedMin))
		Expect(rng.Max).To(Equal(expectedMax))

		if expectedErr == nil {
			Expect(err).ShouldNot(HaveOccurred())
		} else {
			Expect(err).To(Equal(expectedErr))
		}
	})
}

func testSet(setStr string, expectedSet SequenceSet, expectedErr error) {
	testName := fmt.Sprintf("should interpret %s", setStr)
	if expectedErr != nil {
		testName = fmt.Sprintf("should throw an error attempting to interpret %s", setStr)
	}

	It(testName, func() {
		set, err := InterpretSequenceSet(setStr)
		if expectedErr == nil {
			Expect(err).ShouldNot(HaveOccurred())
		} else {
			Expect(err).To(Equal(expectedErr))
		}

		Expect(set).To(HaveLen(len(expectedSet)))
		Expect(set).To(Equal(expectedSet))
	})
}

func testNumber(numberStr string, expectedLast, expectedNil, expectedErr bool, expectedVal uint32) {
	number := SequenceNumber(numberStr)

	testStr := fmt.Sprintf("SequenceNumber(\"%s\").Last() should return %v", numberStr, expectedLast)
	It(testStr, func() {
		Expect(number.Last()).To(Equal(expectedLast))
	})

	testStr = fmt.Sprintf("SequenceNumber(\"%s\").Nil() should return %v", numberStr, expectedNil)
	It(testStr, func() {
		Expect(number.Nil()).To(Equal(expectedNil))
	})

	testStr = fmt.Sprintf("SequenceNumber(\"%s\").Value() should return %d", numberStr, expectedVal)
	if expectedErr {
		testStr = fmt.Sprintf("SequenceNumber(\"%s\").Value() should throw an error", numberStr)
	}
	It(testStr, func() {
		val, err := number.Value()
		Expect(val).To(Equal(expectedVal))
		if expectedErr {
			Expect(err).To(HaveOccurred())
		}
	})
}

var _ = Describe("Sequence Numbers", func() {
	Context("SequenceRange", func() {
		testRange("15:95", "15", "95", nil)
		testRange("95:15", "15", "95", nil)
		testRange("*:16", "16", "*", nil)
		testRange("*:*", "*", "", nil)
		testRange("12:12", "12", "12", nil)
		testRange("53:*", "53", "*", nil)
		testRange("35", "35", "", nil)
		testRange("*", "*", "", nil)
		testRange("5*", "", "", errInvalidRangeString("5*"))
		testRange("*5*", "", "", errInvalidRangeString("*5*"))
		testRange("hello", "", "", errInvalidRangeString("hello"))
	})

	Context("SequenceSet", func() {
		testSet("118:*", SequenceSet{
			SequenceRange{Min: "118", Max: "*"},
		}, nil)
		testSet("1,3,4:14", SequenceSet{
			SequenceRange{Min: "1", Max: ""},
			SequenceRange{Min: "3", Max: ""},
			SequenceRange{Min: "4", Max: "14"},
		}, nil)
		testSet("1,3,8:14,18:*", SequenceSet{
			SequenceRange{Min: "1", Max: ""},
			SequenceRange{Min: "3", Max: ""},
			SequenceRange{Min: "8", Max: "14"},
			SequenceRange{Min: "18", Max: "*"},
		}, nil)
		testSet("1,3,:8:14,18:*", nil, errInvalidSequenceSetString("1,3,:8:14,18:*"))
	})

	Context("SequenceNumber", func() {
		const (
			IsNil    = true
			NotNil   = false
			IsLast   = true
			NotLast  = false
			HasValue = false
			ErrValue = true
		)

		testNumber("*", IsLast, NotNil, ErrValue, 0)
		testNumber("56", NotLast, NotNil, HasValue, 56)
		testNumber("6", NotLast, NotNil, HasValue, 6)
		testNumber("", NotLast, IsNil, ErrValue, 0)
	})
})
