package parser

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"math"
	"unicode/utf8"
)

// ErrInvalidUTF8 is returned when a row or header contains invalid UTF-8.
var ErrInvalidUTF8 = errors.New("invalid UTF-8 encoding")

// EncodingError wraps ErrInvalidUTF8 with line context for reporting.
type EncodingError struct {
	LineNumber int
	Err        error
}

func (e *EncodingError) Error() string {
	return fmt.Sprintf("line %d: %v", e.LineNumber, e.Err)
}

func (e *EncodingError) Unwrap() error { return e.Err }

// Parser represents a streaming CSV parser that reads from the input without buffering the entire file.
type Parser struct {
	reader     *csv.Reader
	lineNumber int
	headers    []string
	delimiter  rune
}

// Row represents a single CSV row with metadata
type Row struct {
	LineNumber int
	Data       []string
	Headers    []string
}

// IsEmpty checks if all fields in the row are empty
func (r *Row) IsEmpty() bool {
	if len(r.Data) == 0 {
		return true
	}
	if len(r.Data) == 1 && r.Data[0] == "" {
		return true
	}
	for _, field := range r.Data {
		if field != "" {
			return false
		}
	}
	return true
}

func validUTF8Strings(ss []string) bool {
	for _, s := range ss {
		if !utf8.ValidString(s) {
			return false
		}
	}
	return true
}

// NewParser creates a new streaming CSV parser that reads directly from input without loading the entire file into memory.
func NewParser(input io.Reader, delimiter string) (*Parser, error) {
	if delimiter == "" {
		return nil, fmt.Errorf("delimiter cannot be empty")
	}
	reader := csv.NewReader(input)
	reader.Comma = rune(delimiter[0])
	reader.FieldsPerRecord = -1

	return &Parser{
		reader:    reader,
		delimiter: rune(delimiter[0]),
	}, nil
}

// Close is a no-op since we don't own the reader
func (p *Parser) Close() error {
	return nil
}

// ReadHeaders reads and returns the header row, validating UTF-8.
func (p *Parser) ReadHeaders() ([]string, error) {
	headers, err := p.reader.Read()
	if err != nil {
		if err == io.EOF {
			return nil, fmt.Errorf("empty input: no headers found")
		}
		return nil, fmt.Errorf("failed to read headers: %w", err)
	}
	if !validUTF8Strings(headers) {
		return nil, &EncodingError{LineNumber: 1, Err: ErrInvalidUTF8}
	}
	p.lineNumber++
	p.headers = headers
	return headers, nil
}

// ReadRow reads the next row from the CSV file and validates UTF-8 per record.
func (p *Parser) ReadRow() (*Row, error) {
	record, err := p.reader.Read()
	if err == io.EOF {
		return nil, io.EOF
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read row %d: %w", p.lineNumber+1, err)
	}
	if !validUTF8Strings(record) {
		return nil, &EncodingError{LineNumber: p.lineNumber + 1, Err: ErrInvalidUTF8}
	}
	p.lineNumber++
	return &Row{
		LineNumber: p.lineNumber,
		Data:       record,
		Headers:    p.headers,
	}, nil
}

// GetLineNumber returns the current line number
func (p *Parser) GetLineNumber() int {
	return p.lineNumber
}

const (
	// DefaultInferSamplePct is the fraction of total rows used for schema
	// inference. Combined with the maxRows hard cap this yields
	// k = min(N, min(maxRows, max(DefaultInferSampleFloor, ⌈N×pct⌉))).
	DefaultInferSamplePct = 0.25

	// DefaultInferSampleFloor is the minimum number of rows included in the
	// inference sample regardless of file size, provided maxRows allows it.
	DefaultInferSampleFloor = 5
)

// computeSampleK returns the number of head rows to use as the inference
// sample given the total number of non-empty data rows and the hard cap.
func computeSampleK(totalRows, maxRows int) int {
	if totalRows == 0 {
		return 0
	}
	pct := int(math.Ceil(float64(totalRows) * DefaultInferSamplePct))
	k := max(pct, DefaultInferSampleFloor)
	k = min(k, maxRows)
	k = min(k, totalRows)
	return k
}

// ReadSampleFromReader reads a head sample from r for schema inference.
func ReadSampleFromReader(r io.Reader, delimiter string, maxRows int) (headers []string, sample [][]string, replay io.Reader, err error) {
	if delimiter == "" {
		return nil, nil, nil, fmt.Errorf("delimiter cannot be empty")
	}

	if rs, ok := r.(io.ReadSeeker); ok {
		return readSampleSeekable(rs, delimiter, maxRows)
	}
	return readSampleStream(r, delimiter, maxRows)
}

// readSampleSeekable handles the seekable (file) case for ReadSampleFromReader.
func readSampleSeekable(rs io.ReadSeeker, delimiter string, maxRows int) (headers []string, sample [][]string, replay io.Reader, err error) {
	startPos, err := rs.Seek(0, io.SeekCurrent)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("seek: %w", err)
	}

	// Pass 1: count non-empty data rows.
	rdCount := csv.NewReader(rs)
	rdCount.Comma = rune(delimiter[0])
	rdCount.FieldsPerRecord = -1
	if _, err = rdCount.Read(); err != nil { // skip header
		if err == io.EOF {
			return nil, nil, nil, fmt.Errorf("empty input: no headers found")
		}
		return nil, nil, nil, fmt.Errorf("failed to read headers: %w", err)
	}
	totalRows := 0
	for {
		record, readErr := rdCount.Read()
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return nil, nil, nil, readErr
		}
		row := &Row{Data: record}
		if !row.IsEmpty() {
			totalRows++
		}
	}

	// Seek back to start for pass 2.
	if _, err = rs.Seek(startPos, io.SeekStart); err != nil {
		return nil, nil, nil, fmt.Errorf("seek: %w", err)
	}

	// Pass 2: read exactly k rows.
	k := computeSampleK(totalRows, maxRows)
	rdSample := csv.NewReader(rs)
	rdSample.Comma = rune(delimiter[0])
	rdSample.FieldsPerRecord = -1
	headers, err = rdSample.Read()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to read headers: %w", err)
	}
	if !validUTF8Strings(headers) {
		return nil, nil, nil, &EncodingError{LineNumber: 1, Err: ErrInvalidUTF8}
	}
	lineNum := 1
	for len(sample) < k {
		record, readErr := rdSample.Read()
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return nil, nil, nil, readErr
		}
		lineNum++
		if !validUTF8Strings(record) {
			return nil, nil, nil, &EncodingError{LineNumber: lineNum, Err: ErrInvalidUTF8}
		}
		row := &Row{Data: record}
		if !row.IsEmpty() {
			sample = append(sample, record)
		}
	}

	// Seek back to start so the caller can replay the full file from rs.
	if _, err = rs.Seek(startPos, io.SeekStart); err != nil {
		return nil, nil, nil, fmt.Errorf("seek: %w", err)
	}
	return headers, sample, rs, nil
}

// readSampleStream handles non-seekable readers (STDIN, pipes) for
// ReadSampleFromReader. N is unknown so it falls back to a head sample of up
// to maxRows rows, buffering only those rows via TeeReader.
func readSampleStream(r io.Reader, delimiter string, maxRows int) (headers []string, sample [][]string, replay io.Reader, err error) {
	var captureBuf bytes.Buffer
	tee := io.TeeReader(r, &captureBuf)
	rd := csv.NewReader(tee)
	rd.Comma = rune(delimiter[0])
	rd.FieldsPerRecord = -1
	headers, err = rd.Read()
	if err != nil {
		if err == io.EOF {
			return nil, nil, nil, fmt.Errorf("empty input: no headers found")
		}
		return nil, nil, nil, fmt.Errorf("failed to read headers: %w", err)
	}
	if !validUTF8Strings(headers) {
		return nil, nil, nil, &EncodingError{LineNumber: 1, Err: ErrInvalidUTF8}
	}
	lineNum := 1
	for len(sample) < maxRows {
		record, readErr := rd.Read()
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return nil, nil, nil, readErr
		}
		lineNum++
		if !validUTF8Strings(record) {
			return nil, nil, nil, &EncodingError{LineNumber: lineNum, Err: ErrInvalidUTF8}
		}
		row := &Row{Data: record}
		if !row.IsEmpty() {
			sample = append(sample, record)
		}
	}
	// replay = captured head bytes + unconsumed remainder of the stream.
	replay = io.MultiReader(bytes.NewReader(captureBuf.Bytes()), r)
	return headers, sample, replay, nil
}

// ReadSampleFromBytes parses CSV from b and returns headers and up to k
// non-empty data rows where k = computeSampleK(N, maxRows).
func ReadSampleFromBytes(b []byte, delimiter string, maxRows int) (headers []string, sample [][]string, err error) {
	if delimiter == "" {
		return nil, nil, fmt.Errorf("delimiter cannot be empty")
	}
	rd := csv.NewReader(bytes.NewReader(b))
	rd.Comma = rune(delimiter[0])
	rd.FieldsPerRecord = -1
	headers, err = rd.Read()
	if err != nil {
		if err == io.EOF {
			return nil, nil, fmt.Errorf("empty input: no headers found")
		}
		return nil, nil, fmt.Errorf("failed to read headers: %w", err)
	}
	if !validUTF8Strings(headers) {
		return nil, nil, &EncodingError{LineNumber: 1, Err: ErrInvalidUTF8}
	}
	var allRows [][]string
	lineNum := 1
	for {
		record, readErr := rd.Read()
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return nil, nil, readErr
		}
		lineNum++
		if !validUTF8Strings(record) {
			return nil, nil, &EncodingError{LineNumber: lineNum, Err: ErrInvalidUTF8}
		}
		row := &Row{Data: record}
		if !row.IsEmpty() {
			allRows = append(allRows, record)
		}
	}
	k := computeSampleK(len(allRows), maxRows)
	return headers, allRows[:k], nil
}
