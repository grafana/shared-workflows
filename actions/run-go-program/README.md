# go-build-cache-run

This is a composite GitHub Action used to build, cache, and run Go programs. The
idea is to build the program only one time, and then cache the binary for future
runs, to save runtime.

This action handles building your Go program, caching the built binary, and
running it with specified arguments. It uses the Go version specified in your
`go.mod` file and efficiently caches dependencies.

```yaml
name: Build and Run Go Program
jobs:
  build-and-run:
    name: Build and Run Go Program
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Build Go program
        id: build
        uses: grafana/shared-workflows/actions/run-go-program@main
        with:
          source-dir: myapp/
          working-dir: ./workdir

    - name: Run Go program
      run: |
        ${{ steps.build.outputs.binary }} -flag1 -flag2
```

## Inputs

| Name          | Type   | Description                                        | Required | Default |
| ------------- | ------ | -------------------------------------------------- | -------- | ------- |
| `source-dir`  | String | The directory containing the Go source files       | Yes      | N/A     |
| `packages`    | String | The packages to build (first argument to go build) | No       | "."     |
| `run`         | String | Whether to run the binary after building           | No       | "true"  |
| `args`        | String | Arguments to pass to the binary                    | No       | ""      |
| `working-dir` | String | The working directory to run the binary in         | Yes      | N/A     |

Think of `packages` as the arguments you would pass to `go build`, and `args` as
the arguments you would pass to the binary being build.

## Notes

The cache is based on the contents of the source directory, including all files
and subdirectories.
