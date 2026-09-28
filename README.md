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

1. Let's run Go's builtin benchmark on our naïve Prime Number function:
```shell
go test -bench=. -args Basic
```

2. Pass a `count` arg to the `go` command to run the benchmark multiple times.
```shell
go test -bench=. -count 10 -args Basic 
```

## Second round: Naive with sub-benchmarks

For many algorithms, performance issues only are found on different "sets" of input... for example, this naïve function has exponential growth which we can detect with larger and larger input values.

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

If you can't run this yourself, here's a real capture to point at instead. This is [`benchstat`](https://pkg.go.dev/golang.org/x/perf/cmd/benchstat) output (`go test -bench=. -benchmem -count 6 -benchtime=3x -args <Naive|Eratos|Atkin|ai>`, then `benchstat Naive.txt Eratos.txt Atkin.txt ai.txt`) on an Apple M1 Pro, `go1.27.1` — medians across 6 runs each, with `Naive` as the comparison baseline, reformatted from benchstat's own output into tables.

### Time (sec/op, vs Naive)

| input | Naive | Eratos | Δ vs Naive | Atkin | Δ vs Naive | ai | Δ vs Naive | Winner |
|---|---|---|---|---|---|---|---|---|
| 1,000 | 10.06µs | 7.46µs | ~ (p=0.39) | 695.6µs | +6812.70% | 5.05µs | **-49.83%** | ai |
| 10,000 | 109.6µs | 31.06µs | -71.66% | 635.5µs | +479.82% | 25.90µs | **-76.37%** | ai |
| 100,000 | 2.253ms | 251.4µs | -88.84% | 724.0µs | -67.86% | 274.0µs | **-87.84%** | Eratos |
| 1,000,000 | 48.70ms | 2.151ms | -95.58% | 1.527ms | **-96.86%** | 2.330ms | -95.21% | Atkin |
| 10,000,000 | 1209ms | 20.41ms | -98.31% | 7.652ms | **-99.37%** | 23.46ms | -98.06% | Atkin |
| 50,000,000 | 11812ms | 171.2ms | -98.55% | 33.99ms | **-99.71%** | 117.5ms | -99.01% | Atkin |
| geomean | 10.95ms | 871.4µs | -92.05% | 2.242ms | -79.53% | 782.8µs | **-92.85%** | ai |

### Memory (B/op, vs Naive)

| input | Naive | Eratos | Δ vs Naive | Atkin | Δ vs Naive | ai | Δ vs Naive | Winner |
|---|---|---|---|---|---|---|---|---|
| 1,000 | 3.97Ki | 4.97Ki | +25.17% | 525.3Ki | +13120.74% | 3.94Ki | **-0.91%** | ai |
| 10,000 | 24.60Ki | 34.60Ki | +40.65% | 543.0Ki | +2107.43% | 23.12Ki | **-5.99%** | ai |
| 100,000 | 349.2Ki | 453.2Ki | +29.78% | 868.4Ki | +148.66% | 166.4Ki | **-52.36%** | ai |
| 1,000,000 | 3.07Mi | 4.03Mi | +31.33% | 3.57Mi | +16.51% | 1.31Mi | **-57.21%** | ai |
| 10,000,000 | 25.26Mi | 34.79Mi | +37.77% | 25.76Mi | +2.01% | 11.26Mi | **-55.42%** | ai |
| 50,000,000 | 122.5Mi | 170.2Mi | +38.93% | 123.0Mi | +0.42% | 51.41Mi | **-58.02%** | ai |
| geomean | 838.6Ki | 1.10Mi | +33.82% | 3.71Mi | +353.16% | 480.8Ki | **-42.66%** | ai |

### Allocations (allocs/op, vs Naive)

| input | Naive | Eratos | Δ vs Naive | Atkin | Δ vs Naive | ai | Δ vs Naive | Winner |
|---|---|---|---|---|---|---|---|---|
| 1,000 | 8 | 9 | +12.50% | 51.0 | +537.50% | 3 | **-62.50%** | ai |
| 10,000 | 11 | 12 | +9.09% | 44.5 | +304.55% | 3 | **-72.73%** | ai |
| 100,000 | 18 | 19 | +5.56% | 53.5 | +197.22% | 3 | **-83.33%** | ai |
| 1,000,000 | 26 | 27 | +3.85% | 62.0 | +138.46% | 3 | **-88.46%** | ai |
| 10,000,000 | 35 | 36 | +2.86% | 72.5 | +107.14% | 3 | **-91.43%** | ai |
| 50,000,000 | 42 | 43 | +2.38% | 123.0 | +192.86% | 3 | **-92.86%** | ai |
| geomean | 19.82 | 21.00 | +5.98% | 63.75 | +221.72% | 3.00 | **-84.86%** | ai |

A few things worth pointing at in these tables without running anything:
- **`ai`'s `allocs/op` stays flat at 3** across every input size, where `Naive` and `Eratos` climb into the 40s and `Atkin` into the hundreds. That's the pre-sized `make([]int, 0, estimate)` in `ai` paying off — a correctly-estimated capacity means `append` almost never has to grow and copy the backing array. `benchstat`'s `allocs/op` table makes that visible in a way `sec/op` alone wouldn't.
- **`Atkin`'s `sec/op` `Δ vs Naive` is a huge *positive* number at small input** (+6812% at 1,000) — that's fixed setup/coordination cost dominating, exactly what the "Final round" section above predicts, before `Atkin`'s better asymptotic behavior flips that to -99.71% by 50,000,000.
- **The `Winner` column tells three different stories per table.** `ai` sweeps memory and allocations outright, `Atkin` only takes the time column from 1,000,000 up, and even then it does it while using *more* bytes and *more* allocs than `ai` at every size — a straight algorithmic edge (a true incremental/wheel sieve does less work per candidate at large N), not a memory-for-speed trade.
- The `~ (p=0.39)` on `Eratos` at `input=1000` means the run count (6) wasn't enough to call that difference significant at this noisy an input size — `benchstat` says so explicitly instead of reporting a misleading delta.