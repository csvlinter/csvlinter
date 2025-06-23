package parser

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"unicode/utf8"
)

// Parser represents a streaming CSV parser
type Parser struct {
	reader     *csv.Reader
	lineNumber int
	headers    []string
	buffer     *bytes.Buffer
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

// NewParser creates a new CSV parser
func NewParser(input io.Reader, delimiter string) (*Parser, error) {
	// Read all input into a buffer for UTF-8 validation and rewinding
	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, input); err != nil {
		return nil, fmt.Errorf("failed to read input: %w", err)
	}

	reader := csv.NewReader(bytes.NewReader(buf.Bytes()))
	reader.Comma = rune(delimiter[0])
	reader.FieldsPerRecord = -1 // Allow variable number of fields

	return &Parser{
		reader:    reader,
		buffer:    buf,
		delimiter: rune(delimiter[0]),
	}, nil
}

// Close is a no-op since we don't own the reader
func (p *Parser) Close() error {
	return nil
}

// ReadHeaders reads and returns the header row
func (p *Parser) ReadHeaders() ([]string, error) {
	headers, err := p.reader.Read()
	if err != nil {
		if err == io.EOF {
			return nil, fmt.Errorf("empty input: no headers found")
		}
		return nil, fmt.Errorf("failed to read headers: %w", err)
	}

	p.lineNumber++
	p.headers = headers
	return headers, nil
}

// ReadRow reads the next row from the CSV file
func (p *Parser) ReadRow() (*Row, error) {
	record, err := p.reader.Read()
	if err == io.EOF {
		return nil, io.EOF
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read row %d: %w", p.lineNumber+1, err)
	}

	p.lineNumber++
	return &Row{
		LineNumber: p.lineNumber,
		Data:       record,
		Headers:    p.headers,
	}, nil
}

// ValidateUTF8 checks if the input is valid UTF-8
func (p *Parser) ValidateUTF8() error {
	data := p.buffer.Bytes()
	if !utf8.Valid(data) {
		return fmt.Errorf("input contains invalid UTF-8 encoding")
	}

	// Create a new reader from the buffer for parsing
	p.reader = csv.NewReader(bytes.NewReader(data))
	p.reader.Comma = p.delimiter
	p.reader.FieldsPerRecord = -1

	return nil
}

// GetLineNumber returns the current line number
func (p *Parser) GetLineNumber() int {
	return p.lineNumber
}
