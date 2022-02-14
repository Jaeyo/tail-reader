package tail_reader

import (
	"github.com/pkg/errors"
	"io"
	"os"
	"strings"
)

type TailReader struct {
	file           *os.File
	fileSize       int64
	bufferSize     int64
	reverseCursor  int64 // always minus
	lines          []string
	incompleteLine string
}

func New(filePath string, bufferSize int64) (*TailReader, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to open file (filePath: %s)", filePath)
	}

	stat, err := file.Stat()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to stat file (filePath: %s)", filePath)
	}

	return &TailReader{
		file:           file,
		fileSize:       stat.Size(),
		bufferSize:     bufferSize,
		reverseCursor:  0,
		lines:          []string{},
		incompleteLine: "",
	}, nil
}

func (r *TailReader) Read() (string, error) {
	// max line length to pop: buffer * 10
	for i := 0; i < 10; i++ {
		line := r.popLine()
		if line != nil {
			return *line, nil
		}

		if err := r.read(); err != nil {
			return "", err
		}
	}

	line := r.incompleteLine
	r.incompleteLine = ""
	return line, nil
}

func (r *TailReader) HasNext() bool {
	return !r.isReadComplete() || len(r.lines) > 0
}

func (r *TailReader) popLine() *string {
	if len(r.lines) == 0 {
		return nil
	}

	size := len(r.lines)
	line := r.lines[size-1]
	r.lines = r.lines[0 : size-1]
	return &line
}

func (r *TailReader) read() error {
	bufferSize := r.getNextBufferSize()
	isAlreadyDone := bufferSize == 0
	if isAlreadyDone {
		return nil
	}

	r.reverseCursor -= bufferSize
	if _, err := r.file.Seek(r.reverseCursor, io.SeekEnd); err != nil {
		return errors.Wrap(err, "failed to seek file")
	}

	buf := make([]byte, bufferSize)
	if _, err := r.file.Read(buf); err != nil {
		return errors.Wrap(err, "failed to read file into buffer")
	}

	r.lines, r.incompleteLine = r.parse(string(buf) + r.incompleteLine)

	return nil
}

func (r *TailReader) parse(lines string) ([]string, string) {
	splited := strings.Split(lines, "\n")

	if len(splited) == 1 {
		if r.isReadComplete() {
			return []string{lines}, ""
		}

		return []string{}, lines
	}

	if r.isReadComplete() {
		return splited, ""
	}

	return splited[1:], splited[0]
}

func (r *TailReader) isReadComplete() bool {
	return r.reverseCursor == -r.fileSize
}

func (r *TailReader) getNextBufferSize() int64 {
	if r.fileSize+r.reverseCursor < r.bufferSize {
		return r.fileSize + r.reverseCursor
	}

	return r.bufferSize
}

func (r *TailReader) Close() error {
	return r.file.Close()
}
