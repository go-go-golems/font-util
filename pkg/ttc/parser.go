package ttc

import (
	"encoding/binary"
	"fmt"
	"os"
)

// TTCTag is the magic bytes at the start of a TTC file.
const TTCTag = "ttcf"

// TTCHeader represents the top-level header of a TrueType Collection file.
type TTCHeader struct {
	Tag          string
	MajorVersion uint16
	MinorVersion uint16
	NumFonts     uint32
	FontOffsets  []uint32
}

// FontHeader represents the Offset Table (SFNT header) of a single font.
type FontHeader struct {
	SFNTVersion   uint32
	NumTables     uint16
	SearchRange   uint16
	EntrySelector uint16
	RangeShift    uint16
	TableRecords  []TableRecord
}

// TableRecord represents a single entry in the font's table directory.
type TableRecord struct {
	TagBytes [4]byte
	CheckSum uint32
	Offset   uint32
	Length   uint32
}

// Tag returns the 4-character tag as a string.
func (tr TableRecord) Tag() string {
	return string(tr.TagBytes[:])
}

// FontEntry represents a single font within the TTC, with its header
// and extracted name.
type FontEntry struct {
	Header FontHeader
	Index  int
	Name   string
}

// TTCFile represents a parsed TTC file containing all member fonts.
type TTCFile struct {
	Header TTCHeader
	Fonts  []FontEntry
	Data   []byte // Raw file data for table extraction
}

// ParseFile reads and parses a TTC file from disk.
func ParseFile(path string) (*TTCFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading TTC file %s: %w", path, err)
	}
	return Parse(data)
}

// Parse reads and parses a TTC file from a byte slice.
func Parse(data []byte) (*TTCFile, error) {
	if len(data) < 12 {
		return nil, fmt.Errorf("file too small to be a TTC (got %d bytes)", len(data))
	}

	tag := string(data[0:4])
	if tag != TTCTag {
		return nil, fmt.Errorf("not a TTC file: expected tag %q, got %q", TTCTag, tag)
	}

	header := TTCHeader{
		Tag:          tag,
		MajorVersion: binary.BigEndian.Uint16(data[4:6]),
		MinorVersion: binary.BigEndian.Uint16(data[6:8]),
		NumFonts:     binary.BigEndian.Uint32(data[8:12]),
	}

	if header.NumFonts == 0 {
		return nil, fmt.Errorf("TTC file contains no fonts")
	}

	// Read font offsets
	header.FontOffsets = make([]uint32, header.NumFonts)
	for i := uint32(0); i < header.NumFonts; i++ {
		offset := 12 + i*4
		if offset+4 > uint32(len(data)) {
			return nil, fmt.Errorf("font offset %d out of bounds (need %d, have %d)", i, offset+4, len(data))
		}
		header.FontOffsets[i] = binary.BigEndian.Uint32(data[offset : offset+4])
	}

	ttcFile := &TTCFile{
		Header: header,
		Data:   data,
	}

	// Parse each font
	for i, fontOffset := range header.FontOffsets {
		fontEntry, err := parseFontEntry(data, fontOffset, i)
		if err != nil {
			return nil, fmt.Errorf("parsing font %d: %w", i, err)
		}
		ttcFile.Fonts = append(ttcFile.Fonts, *fontEntry)
	}

	return ttcFile, nil
}

func parseFontEntry(data []byte, offset uint32, index int) (*FontEntry, error) {
	if offset+12 > uint32(len(data)) {
		return nil, fmt.Errorf("font offset table at %d out of bounds (need %d, have %d)", offset, offset+12, len(data))
	}

	fontHeader := FontHeader{
		SFNTVersion:   binary.BigEndian.Uint32(data[offset : offset+4]),
		NumTables:     binary.BigEndian.Uint16(data[offset+4 : offset+6]),
		SearchRange:   binary.BigEndian.Uint16(data[offset+6 : offset+8]),
		EntrySelector: binary.BigEndian.Uint16(data[offset+8 : offset+10]),
		RangeShift:    binary.BigEndian.Uint16(data[offset+10 : offset+12]),
	}

	if fontHeader.NumTables == 0 {
		return nil, fmt.Errorf("font %d has no tables", index)
	}

	// Parse table records
	fontHeader.TableRecords = make([]TableRecord, fontHeader.NumTables)
	for i := uint16(0); i < fontHeader.NumTables; i++ {
		recOffset := offset + 12 + uint32(i)*16
		if recOffset+16 > uint32(len(data)) {
			return nil, fmt.Errorf("table record %d for font %d out of bounds (need %d, have %d)", i, index, recOffset+16, len(data))
		}
		copy(fontHeader.TableRecords[i].TagBytes[:], data[recOffset:recOffset+4])
		fontHeader.TableRecords[i].CheckSum = binary.BigEndian.Uint32(data[recOffset+4 : recOffset+8])
		fontHeader.TableRecords[i].Offset = binary.BigEndian.Uint32(data[recOffset+8 : recOffset+12])
		fontHeader.TableRecords[i].Length = binary.BigEndian.Uint32(data[recOffset+12 : recOffset+16])
	}

	// Validate table data bounds
	for _, rec := range fontHeader.TableRecords {
		end := rec.Offset + rec.Length
		if end > uint32(len(data)) {
			return nil, fmt.Errorf("table %s data out of bounds for font %d (offset=%d, length=%d, need %d, have %d)",
				rec.Tag(), index, rec.Offset, rec.Length, end, len(data))
		}
	}

	// Extract font name from the 'name' table
	name := extractFontName(data, fontHeader.TableRecords, index)

	return &FontEntry{
		Header: fontHeader,
		Index:  index,
		Name:   name,
	}, nil
}

// extractFontName reads the 'name' table from the raw data
// and returns the PostScript name (Name ID 6) from the
// Windows platform (Platform 3, Encoding 1).
// Falls back to "font-{index}" if not found.
func extractFontName(data []byte, records []TableRecord, index int) string {
	for _, rec := range records {
		if rec.Tag() != "name" {
			continue
		}

		if rec.Offset+6 > uint32(len(data)) {
			return fallbackName(index)
		}

		nameTableOffset := rec.Offset
		// name table header
		// Format:       uint16 (0 or 1)
		// Count:        uint16
		// StringOffset: uint16 (offset from start of name table to string storage)
		_ = binary.BigEndian.Uint16(data[nameTableOffset : nameTableOffset+2]) // format
		count := binary.BigEndian.Uint16(data[nameTableOffset+2 : nameTableOffset+4])
		stringOffset := uint32(binary.BigEndian.Uint16(data[nameTableOffset+4 : nameTableOffset+6]))

		// Search for Name ID 6 (PostScript name) from Windows platform
		for i := uint16(0); i < count; i++ {
			recOff := nameTableOffset + 6 + uint32(i)*12
			if recOff+12 > uint32(len(data)) {
				break
			}

			platformID := binary.BigEndian.Uint16(data[recOff : recOff+2])
			encodingID := binary.BigEndian.Uint16(data[recOff+2 : recOff+4])
			// languageID := binary.BigEndian.Uint16(data[recOff+4 : recOff+6])
			nameID := binary.BigEndian.Uint16(data[recOff+6 : recOff+8])
			length := binary.BigEndian.Uint16(data[recOff+8 : recOff+10])
			strOffset := uint32(binary.BigEndian.Uint16(data[recOff+10 : recOff+12]))

			// Prefer Windows platform (3), Unicode BMP encoding (1), English US (0x0409)
			// Also accept Mac Roman (platform 1, encoding 0) as fallback
			if nameID != 6 {
				continue
			}

			isWindows := platformID == 3 && encodingID == 1
			isMac := platformID == 1 && encodingID == 0

			if !isWindows && !isMac {
				continue
			}

			strStart := nameTableOffset + stringOffset + strOffset
			strEnd := strStart + uint32(length)
			if strEnd > uint32(len(data)) {
				continue
			}

			raw := data[strStart:strEnd]
			var result string

			if isWindows {
				// UTF-16BE to UTF-8
				result = decodeUTF16BE(raw)
			} else {
				// Mac Roman — treat as ASCII/Latin-1
				result = string(raw)
			}

			// Sanitize for filename
			result = sanitizeFilename(result)
			if result != "" {
				return result
			}
		}
	}

	return fallbackName(index)
}

// decodeUTF16BE decodes a UTF-16BE byte sequence to a Go string.
func decodeUTF16BE(data []byte) string {
	if len(data)%2 != 0 {
		// Odd length — trim last byte
		data = data[:len(data)-1]
	}
	runes := make([]rune, 0, len(data)/2)
	for i := 0; i < len(data); i += 2 {
		r := rune(binary.BigEndian.Uint16(data[i : i+2]))
		runes = append(runes, r)
	}
	return string(runes)
}

// sanitizeFilename removes characters that are unsafe in filenames.
func sanitizeFilename(s string) string {
	result := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c >= 'a' && c <= 'z':
			result = append(result, c)
		case c >= 'A' && c <= 'Z':
			result = append(result, c)
		case c >= '0' && c <= '9':
			result = append(result, c)
		case c == '-' || c == '_' || c == '.':
			result = append(result, c)
		case c == ' ':
			result = append(result, '-')
			// Skip all other characters
		}
	}
	return string(result)
}

func fallbackName(index int) string {
	return fmt.Sprintf("font-%d", index)
}

// CalcSearchFields computes the binary search optimization fields
// for a table directory with numTables entries.
func CalcSearchFields(numTables uint16) (uint16, uint16, uint16) {
	power := uint16(1)
	entrySelector := uint16(0)
	for power*2 <= numTables {
		power *= 2
		entrySelector++
	}
	searchRange := power * 16
	rangeShift := numTables*16 - searchRange
	return searchRange, entrySelector, rangeShift
}
