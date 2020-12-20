package writer

// Writer type writes data to an output type.
type Writer interface {
	// Close closes the output writer flushing it.
	Close() error
	// Write writes the data to an output writer.
	Write(data []byte) error
}
