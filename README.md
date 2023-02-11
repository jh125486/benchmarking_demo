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

---

## First round: Naive basic

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
go test -bench=. -count 10 -args Naive > naive.txt
go test -bench=. -count 10 -args Eratos > eratos.txt
```

4. [`benchstat`](https://pkg.go.dev/golang.org/x/perf/cmd/benchstat) can calculate differences with a confidence interval at level 0.95, so let's show generate that:
```shell
benchstat eratos.txt naive.txt
```
It should show an impressive speed up... so that's a great start.

> *Sidenote*: `benchstat` can only compare the benchmarks if the test names are the same, which is the reason we have been running the same benchmark function, and just passing a command line arg to switch between the different "backing" algorithms.

---

## Final round?: time to get concurrent

If you noticed during the last few benchmarks that the CPU really wasn't being used at all... that's because all modern CPUs have multiple cores and these algorithms are single-CPU bound.

So let's fix that.  There's a number of concurrent Prime-finding algorithms, but we'll use the Sieve of Atkin because, well, it's already written for us here [`primegen`](https://github.com/jbarham/primegen).

1. Run a benchmark on the concurrent Sieve of Atkin and save off the results:

```shell
go test -bench=. -count 10 -args Atkin > atkin.txt
```

2. Check the results against both the non-concurrent *Sieve of Eratosthenes*, and the *Naive* version:
```shell
benchstat atkin.txt eratos.txt
```

You should notice that on the smaller input values that Atkin is actually slower.  That's probably because the actual setup and coordination to handle the concurrent communication has some overhead, which overruns the performance gains on the low-end.  Once the values start to increase, we should see that it performs ~500% faster on the high inputs.


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