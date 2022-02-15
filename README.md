# tail-reader

[![Latest Release](https://img.shields.io/github/release/Jaeyo/tail-reader.svg?style=for-the-badge)](https://github.com/Jaeyo/tail-reader/releases)
[![Go Doc](https://img.shields.io/badge/godoc-reference-blue.svg?style=for-the-badge)](https://pkg.go.dev/github.com/Jaeyo/tail-reader)
[![Software License](https://img.shields.io/badge/license-MIT-blue.svg?style=for-the-badge)](/LICENSE)
[![Go ReportCard](https://goreportcard.com/badge/github.com/Jaeyo/tail-reader?style=for-the-badge)](https://goreportcard.com/report/Jaeyo/tail-reader)

The reader that reads from the tail of the file line by line.

## When to use

ex) periodically access the log file and read only newly added

## Usage

```go
bufferSize := 1024
reader, err := tailreader.New(file, bufferSize)
if err != nil {
    // handle error
    ...
}
defer reader.Close()

for reader.HasNext() {
	line, _ := reader.Read()
	fmt.Println(line)
}

// if file content is below:
//  aaa
//  bbb
//  ccc
//
// then this code prints below:
//  ccc
//  bbb
//  aaa
```
