package writer

import (
	"bufio"
	"os"
	"sync"
)

// File is a concurrent file based output writer.
type File struct {
	file   *os.File
	writer *bufio.Writer
	mutex  *sync.Mutex
}

var _ Writer = &File{}

// New creates a new mutex protected buffered writer for a file
func New(file string, JSON bool) (*File, error) {
	output, err := os.Create(file)
	if err != nil {
		return nil, err
	}
	return &File{file: output, writer: bufio.NewWriter(output), mutex: &sync.Mutex{}}, nil
}

// WriteString writes an output to the underlying file
func (w *File) Write(data []byte) error {
	if len(data) == 0 {
		return nil
	}
	w.mutex.Lock()
	defer w.mutex.Unlock()

	_, err := w.writer.Write(data)
	if err != nil {
		return err
	}
	if data[len(data)-1] != '\n' {
		_, err = w.writer.WriteRune('\n')
	}
	return err
}

// Close closes the underlying writer flushing everything to disk
func (w *File) Close() error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	w.writer.Flush()
	//nolint:errcheck // we don't care whether sync failed or succeeded.
	w.file.Sync()
	return w.file.Close()
}
