package main

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

var testType string

func TestMain(m *testing.M) {
	testType = os.Args[len(os.Args)-1]
	os.Exit(m.Run())
}

var inputs = []int{
	1e3,
	1e4,
	1e5,
	1e6,
	1e7,
	5e7,
}

func BenchmarkPrimeNumbers(b *testing.B) {
	switch strings.ToLower(testType) {
	case "basic":
		primeNumbersBasic(b)
	case "naive":
		primeNumbersNaive(b)
	case "eratos":
		primeNumbersSieveOfEratosthenes(b)
	case "atkin":
		primeNumbersSieveOfAtkin(b)
	default:
		b.Fatalf("No test type '%v' found", testType)
	}
}

func primeNumbersBasic(b *testing.B) {
	for i := 0; i < b.N; i++ {
		naive(inputs[0])
	}
}

func primeNumbersNaive(b *testing.B) {
	for _, v := range inputs {
		name := fmt.Sprintf("input=%d", v)
		b.Run(name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				naive(v)
			}
		})
	}
}

func primeNumbersSieveOfEratosthenes(b *testing.B) {
	for _, v := range inputs {
		name := fmt.Sprintf("input=%d", v)
		b.Run(name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				sieveOfEratosthenes(v)
			}
		})
	}
}

func primeNumbersSieveOfAtkin(b *testing.B) {
	for _, v := range inputs {
		name := fmt.Sprintf("input=%d", v)
		b.Run(name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				sieveOfAtkin(v)
			}
		})
	}
}
