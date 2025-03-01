package archive

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"io"
	"log"
	"os"
	"path/filepath"
)

var ErrMaxSizeExceeded = errors.New("file size exceeds the maximum limit of 1 GB")

// Gzip compresses the files in the srcDir directory and writes the compressed data to the destTar file.
func Gzip(srcDir string, destTar string) error {
	tarFile, err := os.Create(destTar)
	if err != nil {
		return err
	}
	defer tarFile.Close()

	gzipWriter, err := gzip.NewWriterLevel(tarFile, gzip.BestCompression)
	if err != nil {
		return err
	}
	defer gzipWriter.Close()

	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	baseDir := filepath.Clean(srcDir) + string(filepath.Separator)

	err = filepath.Walk(baseDir, func(file string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !fi.Mode().IsRegular() { // Skip non-regular files
			return nil
		}
		if fi.Name() == destTar { // Skip the file that we are creating file
			return nil
		}
		if fi.Name() == ".env" { // Skip the env file that is local
			return nil
		}

		relativePath, err := filepath.Rel(baseDir, file)
		if err != nil {
			return err
		}

		header, err := tar.FileInfoHeader(fi, relativePath)
		if err != nil {
			return err
		}

		header.Name = filepath.ToSlash(relativePath) // Use relative path
		header.Size = fi.Size()                      // Ensure header size is set correctly - avoid "write too long" err

		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}

		f, err := os.Open(file)
		if err != nil {
			return err
		}
		defer f.Close()

		if _, err := io.CopyN(tarWriter, f, fi.Size()); err != nil && errors.Is(err, io.EOF) {
			return err
		}

		return nil
	})

	return err
}

// UnGzip decompresses the .tar.gz file specified by srcTarGz and unpacks it to a temporary directory.
// It then sets the temporary directory as the current working directory.
func UnGzip(srcTarGz string) error { //revive:disable:function-length,cognitive-complexity,cyclomatic
	const maxArchiveSize = 1024 * 1024 * 1024

	// Open the gzip file
	gzipFile, err := os.Open(srcTarGz)
	if err != nil {
		return err
	}
	defer gzipFile.Close()

	// Create a gzip reader
	gzipReader, err := gzip.NewReader(gzipFile)
	if err != nil {
		return err
	}
	defer gzipReader.Close()

	// Create a temporary directory to extract to
	tmpDir, err := os.MkdirTemp("", "present_tar_")
	if err != nil {
		return err
	}
	log.Println("Created temporary directory:", tmpDir)

	// Change the current working directory to the temporary directory
	if err := os.Chdir(tmpDir); err != nil {
		return err
	}

	// Create a tar reader
	tarReader := tar.NewReader(gzipReader)

	// Extract the tar archive
	for {
		header, err := tarReader.Next()
		if err == io.EOF || header == nil {
			break // End of archive
		}
		fileSize := header.Size
		if fileSize > maxArchiveSize {
			return ErrMaxSizeExceeded
		}
		if err != nil {
			return err
		}

		// fmt.Println(header.Name, header.Typeflag)
		switch header.Typeflag {
		case tar.TypeDir: // Directory
			if err := os.MkdirAll(header.Name, 0o755); err != nil {
				return err
			}
		case tar.TypeReg: // Regular file
			if err := os.MkdirAll(filepath.Dir(header.Name), 0o755); err != nil {
				return err
			}
			outFile, err := os.Create(header.Name)
			if err != nil {
				return err
			}
			if _, err := io.Copy(outFile, tarReader); err != nil { //nolint:gosec
				_ = outFile.Close()
				return err
			}
			err = outFile.Close()
			if err != nil {
				return err
			}
		}
	}

	return nil
}
