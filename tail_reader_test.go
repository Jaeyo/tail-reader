package tailreader

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
	"testing"
)

func TestRead(t *testing.T) {
	tempFile, err := ioutil.TempFile(os.TempDir(), "")
	require.NoError(t, err)
	defer os.Remove(tempFile.Name())

	_, err = tempFile.WriteString("aaaa\nbbbb\ncccc")
	require.NoError(t, err)

	tests := []struct {
		bufferSize int64
	}{
		{1},
		{2},
		{3},
		{4},
		{5},
		{1024},
	}

	for i, tc := range tests {
		t.Run(fmt.Sprintf("%d", i+1), func(t *testing.T) {
			reader, err := New(tempFile.Name(), tc.bufferSize)
			require.NoError(t, err)

			require.True(t, reader.HasNext())
			line, err := reader.Read()
			require.NoError(t, err)
			require.Equal(t, "cccc", line)

			require.True(t, reader.HasNext())
			line, err = reader.Read()
			require.NoError(t, err)
			require.Equal(t, "bbbb", line)

			require.True(t, reader.HasNext())
			line, err = reader.Read()
			require.NoError(t, err)
			require.Equal(t, "aaaa", line)

			require.False(t, reader.HasNext())
		})
	}
}

func TestGetNextBufferSize(t *testing.T) {
	bufferSize := int64(1024)

	tests := []struct {
		fileSize           int64
		reverseCursor      int64
		expectedBufferSize int64
	}{
		{0, 0, 0},
		{bufferSize, 0, bufferSize},
		{bufferSize + 1, 0, bufferSize},
		{bufferSize - 1, 0, bufferSize - 1},
		{bufferSize * 2, -300, bufferSize},
		{bufferSize, -300, bufferSize - 300},
		{bufferSize, -bufferSize, 0},
	}

	for i, tc := range tests {
		t.Run(fmt.Sprintf("%d", i+1), func(t *testing.T) {
			reader := &TailReader{
				fileSize:      tc.fileSize,
				reverseCursor: tc.reverseCursor,
				bufferSize:    bufferSize,
			}

			require.Equal(t, tc.expectedBufferSize, reader.getNextBufferSize())
		})
	}
}

func TestParse(t *testing.T) {
	tests := []struct {
		lines                  string
		isReadComplete         bool
		expectedLines          []string
		expectedIncompleteLine string
	}{
		{
			lines:                  "asdf\nasdf",
			isReadComplete:         true,
			expectedLines:          []string{"asdf", "asdf"},
			expectedIncompleteLine: "",
		},
		{
			lines:                  "asdf\nasdf",
			isReadComplete:         false,
			expectedLines:          []string{"asdf"},
			expectedIncompleteLine: "asdf",
		},
		{
			lines:                  "asdfasdf",
			isReadComplete:         false,
			expectedLines:          []string{},
			expectedIncompleteLine: "asdfasdf",
		},
		{
			lines:                  "asdfasdf",
			isReadComplete:         true,
			expectedLines:          []string{"asdfasdf"},
			expectedIncompleteLine: "",
		},
		{
			lines:                  "asdf\n\nasdf",
			isReadComplete:         true,
			expectedLines:          []string{"asdf", "", "asdf"},
			expectedIncompleteLine: "",
		},
		{
			lines:                  "",
			isReadComplete:         true,
			expectedLines:          []string{""},
			expectedIncompleteLine: "",
		},
		{
			lines:                  "",
			isReadComplete:         false,
			expectedLines:          []string{},
			expectedIncompleteLine: "",
		},
		{
			lines:                  "\nasdf\n",
			isReadComplete:         false,
			expectedLines:          []string{"asdf", ""},
			expectedIncompleteLine: "",
		},
	}

	for i, tc := range tests {
		t.Run(fmt.Sprintf("%d", i+1), func(t *testing.T) {
			reader := &TailReader{
				fileSize:      1024,
				reverseCursor: 0,
			}
			if tc.isReadComplete {
				reader.reverseCursor = -reader.fileSize
			}

			lines, incompleteLine := reader.parse(tc.lines)
			require.Equal(t, tc.expectedLines, lines)
			require.Equal(t, tc.expectedIncompleteLine, incompleteLine)
		})
	}
}

func TestParseContinuously(t *testing.T) {
	tests := []struct {
		lines                  []string
		isReadComplete         bool
		expectedLines          []string
		expectedIncompleteLine string
	}{
		{
			lines:                  []string{"hello, ", "world"},
			isReadComplete:         true,
			expectedLines:          []string{"hello, world"},
			expectedIncompleteLine: "",
		},
		{
			lines:                  []string{"aaa\nbb", "b\nccc"},
			isReadComplete:         true,
			expectedLines:          []string{"aaa", "bbb", "ccc"},
			expectedIncompleteLine: "",
		},
		{
			lines:                  []string{"aaa\nbb", "b\nccc"},
			isReadComplete:         false,
			expectedLines:          []string{"bbb", "ccc"},
			expectedIncompleteLine: "aaa",
		},
	}

	for i, tc := range tests {
		t.Run(fmt.Sprintf("%d", i+1), func(t *testing.T) {
			reader := &TailReader{
				fileSize:      1024,
				reverseCursor: 0,
			}

			parsed1, incompleteLine := reader.parse(tc.lines[1])

			if tc.isReadComplete {
				reader.reverseCursor = -reader.fileSize
			}

			parsed2, incompleteLine := reader.parse(tc.lines[0] + incompleteLine)
			lines := append(parsed2, parsed1...)
			require.Equal(t, tc.expectedLines, lines)
			require.Equal(t, tc.expectedIncompleteLine, incompleteLine)
		})
	}
}
