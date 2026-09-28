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

// ai is an odd-only, bit-packed Sieve of Eratosthenes: even numbers are
// never stored, and the composite flags are packed 64 to a word instead of
// one bool per candidate. Both cut the memory traffic that dominates cost
// at large max. Benchmarked against a cache-blocked/segmented variant of
// the same algorithm, which added bookkeeping overhead per block without
// enough locality win to pay for it on this input range — the flat version
// won at every size in inputs, so it's the one kept.
func ai(max int) []int {
	if max <= 2 {
		return []int{}
	}

	estimate := 8
	if max > 6 {
		// Prime-counting approximation, so append rarely reallocates.
		estimate = int(float64(max)/math.Log(float64(max))) + 8
	}

	primes := make([]int, 0, estimate)
	primes = append(primes, 2)

	if max <= 3 {
		return primes
	}

	// index i <-> odd value 2i+1, for i in [1, maxIndex].
	maxIndex := (max - 2) / 2
	bits := make([]uint64, maxIndex/64+1)

	for i := 1; i <= maxIndex; i++ {
		if bits[i>>6]&(1<<uint(i&63)) != 0 {
			continue
		}

		p := 2*i + 1
		primes = append(primes, p)

		for j := (p*p - 1) / 2; j <= maxIndex; j += p {
			bits[j>>6] |= 1 << uint(j&63)
		}
	}

	return primes
}
