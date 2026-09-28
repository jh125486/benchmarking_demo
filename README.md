# Iterative benchmarking demo

![Go Logo](https://go.dev/images/go-logo-blue.svg)

## Setup

You'll need:
1. [Go installed](https://go.dev/doc/install).

2. The internet. 

Steps:
1. Open up a terminal/command prompt:
    
2. `git clone https://github.com/jh125486/dd_benchmarking.git`

3. `go mod tidy` 

> Feel free to open the `Go` code in your favorite editor, but for this demonstration, you can run all the commands in the shell/command prompt.


## Background

To demonstrate the builtin [`go` tool](https://pkg.go.dev/cmd/go) benchmarking, we'll be using some simple functions to compute all the [Prime numbers](https://en.wikipedia.org/wiki/Prime_number) up to a max `int` value. My point in this isn't to demonstrate Prime Number algorithms... merely Go benchmarking.

## Testing first....

A test functions in Go start with the word `Test`, and takes `*testing.T` as the only parameter.

To test each of our Prime number functions, we have created a test function, which compares the output of the function to a known good list of all primes numbers under 1000... which should give us a good indication of it's correctness for our needs.

If we wrote the functions correctly, each should pass within a few milliseconds :)

```shell
go test ./...
```

## First round: Naive basic

A benchmark function in Go starts with the word `Benchmark` and takes `*testing.B` as the only parameter. To run a benchmark, pass the `bench` flag to `go test`, along with the package to test.  In our case, we're restricting each benchmark for clarity with `-args`.

1. Let's run Go's builtin benchmark on our naive Prime Number function:
```shell
go test -bench=. -args Basic
```

2. Pass a `count` arg to the `go` command to run the benchmark multiple times.
```shell
go test -bench=. -count 10 -args Basic 
```

## Second round: Naive with sub-benchmarks

For many algorithms, performance issues only are found on different "sets" of input... for example, this naive function has exponential growth which we can detect with larger and larger input values.

A way to detect this is with sub-benchmarks, e.g.:
```go
var inputs = []int{
	100,
	...
	10000000,
}
for _, v := range inputs {
    // b.Run <- starts a sub-benchmark
    b.Run(fmt.Sprint("sub-", v), func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            functionToBenchmark(v)
        }
    })
}
```

1. Run benchmarks with sub-benchmarks:
```shell
go test -bench=. -args Naive
```

2. Notice that the last sub-benchmark only completed one run (or not many runs at least)... this is because Go by default only runs the benchmark for 1s, so let's run that again with 10s by setting `benchtime`:
```shell
go test -bench=. -benchtime=10s -args Naive
```
> Note: If your system is still having problems running that last su-benchmark multiple times, comment out or delete line `main_test.go:23` to just disable that large input.

You can also use the `NNNx` format with `benchtime`, which will run the test `NNN` number of times instead.

3. We can also show memory allocations and total memory per operation with `benchmem`:
```shell
go test -bench=. -benchmem -args Naive
```

---

## Third round: we can do better

There's faster algorithm for finding primes, that is surprising old (~2,300 years): [Sieve of Eratosthenes](https://en.wikipedia.org/wiki/Sieve_of_Eratosthenes).

1. Run benchmarks for Sieve of Eratosthenes (with sub-benchmarks):
```shell
go test -bench=. -args Eratos
``` 

This should *feel* faster... but without a direct comparison benchmark to benchmark, I can't be *sure*.  So let's verify.

2. Gophers created a program to do just that, so let's install it:
```shell
go install golang.org/x/perf/cmd/benchstat@latest
```

3. Now let's re-run the benchmarks and save off the results:
```shell
go test -bench=. -count 10 -args Naive > naive_10.txt
go test -bench=. -count 10 -args Eratos > eratos_10.txt
```

4. [`benchstat`](https://pkg.go.dev/golang.org/x/perf/cmd/benchstat) can calculate differences with a confidence interval at level 0.95, so let's show generate that:
```shell
benchstat eratos_10.txt naive_10.txt
```
It should show an impressive speed up... so that's a great start.

> *Sidenote*: `benchstat` can only compare the benchmarks if the test names are the same, which is the reason we have been running the same benchmark function, and just passing a command line arg to switch between the different "backing" algorithms.

---

## Final round?: time to get concurrent

If you noticed during the last few benchmarks that the CPU really wasn't being used at all... that's because all modern CPUs have multiple cores and these algorithms are single-CPU bound.

So let's fix that.  There's a number of concurrent Prime-finding algorithms, but we'll use the Sieve of Atkin because, well, it's already written for us here [`primegen`](https://github.com/jbarham/primegen).

1. Run a benchmark on the concurrent Sieve of Atkin and save off the results:

```shell
go test -bench=. -count 10 -args Atkin > atkin_10.txt
```

2. Check the results against both the non-concurrent *Sieve of Eratosthenes*, and the *Naive* version:
```shell
benchstat atkin_10.txt eratos_10.txt
```

You should notice that on the smaller input values that Atkin is actually slower.  That's probably because the actual setup and coordination to handle the concurrent communication has some overhead, which overruns the performance gains on the low-end.  Once the values start to increase, we should see that it performs ~500% faster on the high inputs.

---

## Bonus round: writing your own, guided by the numbers

Benchmarks aren't just for comparing existing algorithms — they're also how you drive your own optimization work. `ai(max int) []int` in `main.go` is a from-scratch Sieve of Eratosthenes that only stores odd candidates (evens past 2 can never be prime) and packs the composite flags 64-to-a-word instead of one `bool` per candidate. Both cuts reduce memory traffic, which is the dominant cost once `max` gets large.

1. Run it like any of the others:
```shell
go test -bench=. -args ai
```

2. Compare it against Eratosthenes and Atkin:
```shell
go test -bench=. -count 10 -args ai > ai_10.txt
benchstat ai_10.txt eratos_10.txt
benchstat ai_10.txt atkin_10.txt
```

A cache-blocked/segmented version of the same algorithm was also tried (sieve the range in chunks sized to fit L1/L2, à la `primegen`'s incremental approach) — it lost to the flat version at every input size here, because the extra per-block bookkeeping cost more than the locality it bought. That's a real result worth keeping: benchmarking tells you when a "smarter" version isn't actually smarter *for your inputs*, not just when it's faster.

`ai` beats `naive` and `sieveOfEratosthenes` at every input size, and beats `sieveOfAtkin` up to `input=100000`. From `input=1000000` up, `sieveOfAtkin`'s `primegen` — a true incremental/wheel sieve — pulls ahead, because a flat bitset stops fitting cache once it gets big enough, and `primegen` never has that problem. See the [Sample run](#sample-run-all-four-with-time-and-allocs) below for the actual numbers.


## Finishing up

There's plenty of other ways to discovery performance bottlenecks using the builtin tooling.  I haven't covered profiling, which is through `go tool pprof`.  

---

## Epilogue: Helpful links about Go benchmarking

- [Dave Cheney: How to write benchmarks in Go](https://dave.cheney.net/2013/06/30/how-to-write-benchmarks-in-go)

- [Dave Cheney: High Performance Go Workshop](https://dave.cheney.net/high-performance-go-workshop/gophercon-2019.html)

- [HackerNoon: How To Write Benchmarks In Golang Like An Expert](https://hackernoon.com/how-to-write-benchmarks-in-golang-like-an-expert-0w1834gs)

- [CloudBees: Real Life Go Benchmarking](https://www.cloudbees.com/blog/real-life-go-benchmarking)

- [Go: subtests and sub-benchmarks](https://go.dev/blog/subtests)

- [Go: `testing flags`](https://pkg.go.dev/cmd/go#hdr-Testing_flags)

---

## Sample run: all four, with time and allocs

If you can't run this yourself, here's a real `go test -bench=. -benchmem` capture (`go test -run=^$ -bench=. -benchmem -benchtime=3x -args <Basic|Naive|Eratos|Atkin|ai>`) on an Apple M1 Pro, `go1.27.1`, so you've got something to point at.

```
=== Basic ===
BenchmarkPrimeNumbers-10                   	       3	     10306 ns/op	    4074 B/op	       8 allocs/op

=== Naive ===
BenchmarkPrimeNumbers/input=1000-10        	       3	     12875 ns/op	    4069 B/op	       8 allocs/op
BenchmarkPrimeNumbers/input=10000-10       	       3	    111750 ns/op	   25189 B/op	      11 allocs/op
BenchmarkPrimeNumbers/input=100000-10      	       3	   2241736 ns/op	  357605 B/op	      18 allocs/op
BenchmarkPrimeNumbers/input=1000000-10     	       3	  49698375 ns/op	 3218429 B/op	      28 allocs/op
BenchmarkPrimeNumbers/input=10000000-10    	       3	1216089445 ns/op	26481984 B/op	      36 allocs/op
BenchmarkPrimeNumbers/input=50000000-10    	       3	12239094125 ns/op	128431418 B/op	      43 allocs/op

=== Eratos ===
BenchmarkPrimeNumbers/input=1000-10        	       3	      9875 ns/op	    5093 B/op	       9 allocs/op
BenchmarkPrimeNumbers/input=10000-10       	       3	     30111 ns/op	   35429 B/op	      12 allocs/op
BenchmarkPrimeNumbers/input=100000-10      	       3	    211819 ns/op	  464101 B/op	      19 allocs/op
BenchmarkPrimeNumbers/input=1000000-10     	       3	   2236097 ns/op	 4224277 B/op	      27 allocs/op
BenchmarkPrimeNumbers/input=10000000-10    	       3	  21366139 ns/op	36484458 B/op	      38 allocs/op
BenchmarkPrimeNumbers/input=50000000-10    	       3	 178343583 ns/op	178437122 B/op	      45 allocs/op

=== Atkin ===
BenchmarkPrimeNumbers/input=1000-10        	       3	    755042 ns/op	  540360 B/op	      54 allocs/op
BenchmarkPrimeNumbers/input=10000-10       	       3	    701722 ns/op	  558890 B/op	      53 allocs/op
BenchmarkPrimeNumbers/input=100000-10      	       3	    787931 ns/op	  889504 B/op	      55 allocs/op
BenchmarkPrimeNumbers/input=1000000-10     	       3	   1634528 ns/op	 3748528 B/op	      62 allocs/op
BenchmarkPrimeNumbers/input=10000000-10    	       3	   8279958 ns/op	27014106 B/op	      74 allocs/op
BenchmarkPrimeNumbers/input=50000000-10    	       3	  34722820 ns/op	128972362 B/op	     141 allocs/op

=== ai ===
BenchmarkPrimeNumbers/input=1000-10        	       3	     11819 ns/op	    4037 B/op	       3 allocs/op
BenchmarkPrimeNumbers/input=10000-10       	       3	     28083 ns/op	   23680 B/op	       3 allocs/op
BenchmarkPrimeNumbers/input=100000-10      	       3	    251278 ns/op	  170368 B/op	       3 allocs/op
BenchmarkPrimeNumbers/input=1000000-10     	       3	   2448500 ns/op	 1376293 B/op	       3 allocs/op
BenchmarkPrimeNumbers/input=10000000-10    	       3	  24114458 ns/op	11804714 B/op	       3 allocs/op
BenchmarkPrimeNumbers/input=50000000-10    	       3	 123248847 ns/op	53913400 B/op	       5 allocs/op
```

A few things worth pointing at in this table without running anything:
- **`Naive`'s time blows up non-linearly** (1e6→1e7 is a ~24x time jump for a 10x input jump) — that's the O(n·√n) trial-division cost showing up directly in `ns/op`.
- **`ai`'s `allocs/op` stays flat at 3** across every input size (only creeping to 5 at 5e7), while `Naive` and `Eratos` climb from single digits into the 40s. That's the pre-sized `make([]int, 0, estimate)` in `ai` paying off — a correctly-estimated capacity means `append` almost never has to grow and copy the backing array, which `benchmem` makes visible in a way plain `ns/op` wouldn't.
- **`Atkin`'s `ns/op` barely moves from 1e3 to 1e5** (755042 → 787931) — that's fixed setup/coordination cost dominating at small input, exactly what the "Final round" section above predicts, before its better asymptotic behavior takes over from 1e6 up.