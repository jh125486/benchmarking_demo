package main

import (
	"math"

	"github.com/jbarham/primegen"
)

func main() {
	// noop
}

func naive(max int) []int {
	var primes []int

	for i := 2; i < max; i++ {
		isPrime := true

		for j := 2; j <= int(math.Sqrt(float64(i))); j++ {
			if i%j == 0 {
				isPrime = false
				break
			}
		}

		if isPrime {
			primes = append(primes, i)
		}
	}

	return primes
}

func sieveOfEratosthenes(max int) []int {
	b := make([]bool, max)

	var primes []int

	for i := 2; i < max; i++ {
		if b[i] {
			continue
		}

		primes = append(primes, i)

		for k := i * i; k < max; k += i {
			b[k] = true
		}
	}

	return primes
}

func sieveOfAtkin(max int) []int {
	sieve := primegen.New()
	// Start with 2
	sieve.SkipTo(2)
	primes := make([]int, 0)
	for i := sieve.Next(); int(i) <= max; i = sieve.Next() {
		primes = append(primes, int(i))
	}

	return primes
}
