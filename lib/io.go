package lib

import (
	"encoding/binary"
	"io"
	"log"
)

// WriteFramed writes a length-prefixed frame to the writer, where the first 4 bytes contain the frame length
// in big-endian format followed by the frame data.
func WriteFramed(w io.Writer, in []byte) (int, error) {
	length := uint32(len(in))
	out := make([]byte, 4+length)
	binary.BigEndian.PutUint32(out[0:4], length)
	copy(out[4:], in)

	n, err := w.Write(out)
	if err != nil {
		return n, err
	}
	return n, nil
}

// WriteFixed writes a fixed-size byte slice to the writer and returns any error encountered during the write.
func WriteFixed(w io.Writer, in []byte) (int, error) {
	n, err := w.Write(in)
	if err != nil {
		return n, err
	}
	return n, nil
}

// ReadFramed reads a length-prefixed message from the reader, returning the message data without the prefix.
// The function first reads a 4-byte big-endian uint32 indicating the message length, then reads that many bytes.
// Returns an error if reading the length prefix or message data fails.
func ReadFramed(conn io.Reader) ([]byte, error) {
	log.Printf("Reading framed message")
	var lenPrefixSize uint32
	if err := binary.Read(conn, binary.BigEndian, &lenPrefixSize); err != nil {
		return nil, err
	}

	log.Printf("lenPrefixSize: %d", lenPrefixSize)

	data := make([]byte, lenPrefixSize)
	if _, err := io.ReadFull(conn, data); err != nil {
		return nil, err
	}

	return data, nil
}

// ReadFixed reads exactly len(buf) bytes from conn into buf, returning the buffer or an error if the read fails.
func ReadFixed(conn io.Reader, buf []byte) ([]byte, error) {
	if _, err := io.ReadFull(conn, buf); err != nil {
		return nil, err
	}
	return buf, nil
}
