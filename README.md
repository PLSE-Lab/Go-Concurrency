## Running the Analyzer

The analyzer can be run on either individual files or on an entire
directory. To run it on an individual file, you should use the 
`--filePath` command line argument. For instance:

```
go run ast-search.go --filePath sample/goroutines/mem-benchmark.go
```

To run the analyzer on a directory containing .go files, you should
use the `--dirPath` command line argument. For instance:

```
go run ast-search.go --dirPath sample
```

