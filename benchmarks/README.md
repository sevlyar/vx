# benchmarks

The benchmarks behind the [Performance](../README.md#performance) section of the main README, comparing `vx` against `go-playground/validator`, `invopop/validation`, and `nobl9/govy`.

This is a separate Go module (`replace github.com/sevlyar/vx => ../` points it at the local checkout) specifically so these three dependencies never appear in `vx`'s own `go.mod` — `vx` itself stays at zero dependencies.

Run it:

```sh
go test -bench=. -benchmem ./...
```

Each library is called the way its own benchmarks/docs call it, not forced into one shared calling convention — e.g. `vx` is called as `schema.Check(&data)` (a pointer avoids boxing the struct into the `any` parameter), while `go-playground/validator` is called by value, since pointer input measured slower for it specifically. See `bench_test.go` for exactly what each scenario validates.
