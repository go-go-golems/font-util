package ttc

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
)

// ExtractFont extracts a single font from TTC data and writes a standalone TTF file.
// It reassembles the offset table with recalculated search fields and updated offsets.
func ExtractFont(ttcData []byte, font FontEntry, outputPath string) error {
	numTables := font.Header.NumTables

	// Size of offset table + table directory
	headerSize := uint32(12 + numTables*16)

	type tableEntry struct {
		record    TableRecord
		data      []byte
		newOffset uint32
	}

	entries := make([]tableEntry, numTables)
	currentOffset := headerSize

	for i, rec := range font.Header.TableRecords {
		start := rec.Offset
		end := start + rec.Length
		if end > uint32(len(ttcData)) {
			return fmt.Errorf("table %s data out of bounds (offset=%d, length=%d, file size=%d)",
				rec.Tag(), rec.Offset, rec.Length, len(ttcData))
		}

		tableData := make([]byte, rec.Length)
		copy(tableData, ttcData[start:end])

		// Pad to 4-byte boundary
		paddedLen := rec.Length
		if paddedLen%4 != 0 {
			paddedLen += 4 - paddedLen%4
		}

		entries[i] = tableEntry{
			record:    rec,
			data:      tableData,
			newOffset: currentOffset,
		}
		currentOffset += paddedLen
	}

	// Build output buffer
	totalSize := currentOffset
	output := make([]byte, totalSize)

	// Write offset table header
	binary.BigEndian.PutUint32(output[0:4], font.Header.SFNTVersion)
	binary.BigEndian.PutUint16(output[4:6], numTables)

	// Recalculate binary search fields
	searchRange, entrySelector, rangeShift := CalcSearchFields(numTables)
	binary.BigEndian.PutUint16(output[6:8], searchRange)
	binary.BigEndian.PutUint16(output[8:10], entrySelector)
	binary.BigEndian.PutUint16(output[10:12], rangeShift)

	// Write table directory with updated offsets
	for i, entry := range entries {
		off := uint32(12 + i*16)
		copy(output[off:off+4], entry.record.TagBytes[:])
		binary.BigEndian.PutUint32(output[off+4:off+8], entry.record.CheckSum)
		binary.BigEndian.PutUint32(output[off+8:off+12], entry.newOffset)
		binary.BigEndian.PutUint32(output[off+12:off+16], entry.record.Length)
	}

	// Write table data
	for _, entry := range entries {
		copy(output[entry.newOffset:], entry.data)
	}

	// Ensure output directory exists
	dir := filepath.Dir(outputPath)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("creating output directory %s: %w", dir, err)
		}
	}

	return os.WriteFile(outputPath, output, 0644)
}

// ExtractAllFonts extracts all fonts from a TTC file and writes
// them to the output directory. Returns a slice of output file paths
// and a slice of font names.
func ExtractAllFonts(ttcPath string, outputDir string, force bool) ([]string, []string, error) {
	data, err := os.ReadFile(ttcPath)
	if err != nil {
		return nil, nil, fmt.Errorf("reading TTC file: %w", err)
	}

	ttc, err := Parse(data)
	if err != nil {
		return nil, nil, fmt.Errorf("parsing TTC file: %w", err)
	}

	var outputPaths []string
	var fontNames []string

	for _, font := range ttc.Fonts {
		filename := fmt.Sprintf("%s.ttf", font.Name)
		outputPath := filepath.Join(outputDir, filename)

		if !force {
			if _, err := os.Stat(outputPath); err == nil {
				return nil, nil, fmt.Errorf("file already exists: %s (use --force to overwrite)", outputPath)
			}
		}

		if err := ExtractFont(data, font, outputPath); err != nil {
			return nil, nil, fmt.Errorf("extracting font %d (%s): %w", font.Index, font.Name, err)
		}

		outputPaths = append(outputPaths, outputPath)
		fontNames = append(fontNames, font.Name)
	}

	return outputPaths, fontNames, nil
}
