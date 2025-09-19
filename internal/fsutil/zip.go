package fsutil

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// AddToExistingZip adds a new file from the filesystem to an existing zip
// and replaces the original with a new one containing the added file.
func AddToExistingZip(zipPath, newFile, tempDir string) error {
	existingZip, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("failed to open existing zip: %w", err)
	}
	defer existingZip.Close()

	tempContent, err := os.MkdirTemp(tempDir, "content")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempContent)

	fileName := filepath.Base(zipPath)

	tempZip := filepath.Join(tempContent, fileName+".tmp")
	newZipFile, err := os.Create(tempZip)
	if err != nil {
		return fmt.Errorf("failed to create temp zip: %w", err)
	}
	defer newZipFile.Close()

	zipWriter := zip.NewWriter(newZipFile)
	defer zipWriter.Close()

	for _, file := range existingZip.File {
		err := copyFileToZip(zipWriter, file)
		if err != nil {
			return fmt.Errorf("failed to copy existing file %s: %w", file.Name, err)
		}
	}

	err = addNewFileToZip(zipWriter, newFile)
	if err != nil {
		return fmt.Errorf("failed to add new file: %w", err)
	}

	zipWriter.Close()
	newZipFile.Close()

	err = os.Rename(tempZip, zipPath)
	if err != nil {
		return fmt.Errorf("failed to replace original zip: %w", err)
	}

	return nil
}

// copyFileToZip copies an existing file from one zip to another
func copyFileToZip(zipWriter *zip.Writer, file *zip.File) error {
	if file.FileInfo().IsDir() {
		return nil
	}

	reader, err := file.Open()
	if err != nil {
		return err
	}
	defer reader.Close()

	base := filepath.Base(file.Name)

	file.Name = base

	writer, err := zipWriter.CreateHeader(&file.FileHeader)
	if err != nil {
		return err
	}

	_, err = io.Copy(writer, reader)
	return err
}

// addNewFileToZip adds a new file from the filesystem to the zip
func addNewFileToZip(zipWriter *zip.Writer, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}

	header.Name = filepath.Base(filePath)
	header.Method = zip.Deflate

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}

	_, err = io.Copy(writer, file)
	return err
}

// AddContentToExistingZip adds a new file containing the given content to an existing zip
// and replaces the original with a new one containing the added file.
func AddContentToExistingZip(zipPath, newFile, tempDir string, content []byte) error {
	existingZip, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("failed to open existing zip: %w", err)
	}
	defer existingZip.Close()

	tempContent, err := os.MkdirTemp(tempDir, "content")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempContent)

	fileName := filepath.Base(zipPath)

	tempZip := filepath.Join(tempContent, fileName+".tmp")
	newZipFile, err := os.Create(tempZip)
	if err != nil {
		return fmt.Errorf("failed to create temp zip: %w", err)
	}
	defer newZipFile.Close()

	zipWriter := zip.NewWriter(newZipFile)
	defer zipWriter.Close()

	for _, file := range existingZip.File {
		err := copyFileToZip(zipWriter, file)
		if err != nil {
			return fmt.Errorf("failed to copy existing file %s: %w", file.Name, err)
		}
	}

	writer, err := zipWriter.Create(newFile)
	if err != nil {
		return fmt.Errorf("failed to create new file in zip: %w", err)
	}

	_, err = writer.Write(content)
	if err != nil {
		return fmt.Errorf("failed to write content: %w", err)
	}

	zipWriter.Close()
	newZipFile.Close()

	err = os.Rename(tempZip, zipPath)
	if err != nil {
		return fmt.Errorf("failed to replace original zip: %w", err)
	}

	return nil
}
