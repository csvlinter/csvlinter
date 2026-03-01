package parser

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
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
