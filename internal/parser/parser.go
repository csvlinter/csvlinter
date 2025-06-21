package parser

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"unicode/utf8"
)

// Parser represents a streaming CSV parser
type Parser struct {
	file       *os.File
	reader     *csv.Reader
	lineNumber int
	headers    []string
}

// Row represents a single CSV row with metadata
type Row struct {
	LineNumber int
	Data       []string
	Headers    []string
}

// NewParser creates a new CSV parser
func NewParser(filePath, delimiter string) (*Parser, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	reader := csv.NewReader(file)
	reader.Comma = rune(delimiter[0])
	reader.FieldsPerRecord = -1 // Allow variable number of fields

	return &Parser{
		file:   file,
		reader: reader,
	}, nil
}

// Close closes the underlying file
func (p *Parser) Close() error {
	return p.file.Close()
}

// ReadHeaders reads and returns the header row
func (p *Parser) ReadHeaders() ([]string, error) {
	headers, err := p.reader.Read()
	if err != nil {
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

// ValidateUTF8 checks if the file is valid UTF-8
func (p *Parser) ValidateUTF8() error {
	// Reset file position
	if _, err := p.file.Seek(0, 0); err != nil {
		return fmt.Errorf("failed to reset file position: %w", err)
	}

	reader := bufio.NewReader(p.file)
	buffer := make([]byte, 4096)

	for {
		n, err := reader.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read file for UTF-8 validation: %w", err)
		}

		if !utf8.Valid(buffer[:n]) {
			return fmt.Errorf("file contains invalid UTF-8 encoding")
		}
	}

	// Reset file position again for normal parsing
	if _, err := p.file.Seek(0, 0); err != nil {
		return fmt.Errorf("failed to reset file position after UTF-8 validation: %w", err)
	}

	// Recreate reader after seeking
	p.reader = csv.NewReader(p.file)
	p.reader.Comma = p.reader.Comma
	p.reader.FieldsPerRecord = -1

	return nil
}

// GetLineNumber returns the current line number
func (p *Parser) GetLineNumber() int {
	return p.lineNumber
}
