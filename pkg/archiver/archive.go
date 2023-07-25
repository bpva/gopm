package archiver

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func Archive(sourceDir string) ([]byte, error) {
	// Create a new buffer to hold the ZIP archive
	buf := new(bytes.Buffer)

	// Create a new ZIP writer
	archive := zip.NewWriter(buf)

	// Walk through the source directory and add all files to the ZIP archive
	err := filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("failed to access file or directory: %v", err)
		}

		// Get the relative path of the file/directory within the source directory
		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %v", err)
		}

		// Create a new ZIP file header using the relative path
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return fmt.Errorf("failed to create ZIP file header: %v", err)
		}
		header.Name = relPath

		// Check if the file is a directory
		if info.IsDir() {
			header.Name += "/"
		}

		// Write the header to the ZIP archive
		writer, err := archive.CreateHeader(header)
		if err != nil {
			return fmt.Errorf("failed to create ZIP archive entry: %v", err)
		}

		// If the file is not a directory, open it and copy its contents to the ZIP archive
		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return fmt.Errorf("failed to open file: %v", err)
			}
			defer file.Close()

			_, err = io.Copy(writer, file)
			if err != nil {
				return fmt.Errorf("failed to write file contents to ZIP archive: %v", err)
			}
		}

		return nil
	})

	if err != nil {
		archive.Close()
		return nil, fmt.Errorf("failed to walk through source directory: %v", err)
	}

	// Close the ZIP writer to finalize the archive
	err = archive.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to close ZIP archive: %v", err)
	}

	// Get the bytes of the ZIP archive from the buffer
	zipBytes := buf.Bytes()

	return zipBytes, nil
}
