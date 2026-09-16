package filesys

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

var (
	osStat        = os.Stat
	osMkdirAll    = os.MkdirAll
	osWriteFile   = os.WriteFile
	osOpen        = os.Open
	osCreate      = os.Create
	osRename      = os.Rename
	osRemove      = os.Remove
	osLstat       = os.Lstat
	osOpenFile    = os.OpenFile
	fileSync      = func(f *os.File) error { return f.Sync() }
	fileClose     = func(f *os.File) error { return f.Close() }
	ioCopy        = io.Copy
	zipOpenReader = zip.OpenReader
	zipFileInfo   = zip.FileInfoHeader
	filepathWalk  = filepath.Walk
	filepathMatch = filepath.Match
	filepathRel   = filepath.Rel
	filepathAbs   = filepath.Abs
	zipFileOpen   = func(f *zip.File) (io.ReadCloser, error) { return f.Open() }
	newZipWriter  = func(w io.Writer) zipWriter { return zip.NewWriter(w) }
)

type zipWriter interface {
	CreateHeader(*zip.FileHeader) (io.Writer, error)
	Close() error
}

func closeFile(file *os.File) {
	if err := fileClose(file); err != nil {
		return
	}
}

func closeCloser(closer io.Closer) {
	if err := closer.Close(); err != nil {
		return
	}
}

func removeFile(path string) {
	if err := osRemove(path); err != nil {
		return
	}
}

// WriteFile writes the given content to the specified path. It creates
// the necessary directories in the path if they do not already exist.
//
// The function accepts three parameters:
//   - path: A string representing the path where the file should be written.
//     If the directories in the path do not exist, they are created.
//   - content: A byte slice ([]byte) that contains the content to be written to the file.
//   - umask: An os.FileMode value representing the permissions to use when creating
//     the file and any necessary directories.
func WriteFile(path string, content []byte, umask os.FileMode) error {
	dir := filepath.Dir(path)

	if _, err := osStat(dir); os.IsNotExist(err) {
		err := osMkdirAll(dir, umask)
		if err != nil {
			return fmt.Errorf("error making directory %s", err)
		}
	}

	writeErr := osWriteFile(path, content, umask)
	if writeErr != nil {
		return fmt.Errorf("error writing pdf: %s", writeErr)
	}

	return nil
}

// AtomicCopyFileContent copies the contents from a source file to a destination file atomically.
// The operation ensures that the destination file is either fully updated or remains unchanged if an error occurs.
//
// Params:
//
//	src (string): Path to the source file.
//	dest (string): Path to the destination file.
//
// Returns:
//
//	error: Returns an error if the copying process fails at any step.
//
// The function performs the copy by creating a temporary file at the destination. This temporary file
// is used to prevent partial writes to the destination file in case of errors. After copying the content
// successfully, the temporary file is synchronized and then renamed to the destination file name, ensuring
// that the file appears to be updated atomically from the perspective of other processes.
func AtomicCopyFileContent(src string, dest string) error {
	srcFile, err := osOpen(src)
	if err != nil {
		return err
	}

	defer closeFile(srcFile)

	tmpFileName := dest + ".tmp"
	tmpFile, err := osCreate(tmpFileName)

	if err != nil {
		return err
	}
	tmpClosed := false
	defer func() {
		if !tmpClosed {
			closeFile(tmpFile)
		}
	}()

	if _, err := ioCopy(tmpFile, srcFile); err != nil {
		removeFile(tmpFileName)
		return err
	}

	if err := fileSync(tmpFile); err != nil {
		removeFile(tmpFileName)
		return err
	}

	if err := fileClose(tmpFile); err != nil {
		removeFile(tmpFileName)
		return err
	}
	tmpClosed = true

	if err := osRename(tmpFileName, dest); err != nil {
		removeFile(tmpFileName)
		return err
	}

	return nil
}

// IsDirectory checks if the specified path refers to a directory.
// It returns a boolean indicating whether the path is a directory, and any error encountered
// during the operation. If the path does not exist or cannot be accessed, the function returns
// an error, and the boolean is false.
//
// Parameters:
//   - path: The file system path to check.
//
// Returns:
//   - bool: True if the path is a directory, false otherwise.
//   - error: Non-nil error if there was an issue accessing the path.
func IsDirectory(path string) (bool, error) {
	fileInfo, err := osStat(path)
	if err != nil {
		return false, err
	}
	return fileInfo.IsDir(), err
}

// GlobWithDepth finds files within a specified root directory that match a given pattern.
// This function walks the directory tree starting from 'root' and adds paths to the returned slice
// if they match the 'pattern'. The pattern matching respects the syntax of filepath.Match.
//
// Only non-directory file paths that match the pattern are included in the results. Directories
// are traversed but not included in the result set. Any errors encountered during directory traversal
// or pattern matching are returned.
//
// Parameters:
//   - root: The root directory from which the file tree traversal begins.
//   - pattern: The pattern used to match file names, not file paths.
//
// Returns:
//   - []string: A slice of matched file paths.
//   - error: Non-nil error if an error occurred during file tree traversal or pattern matching.
func GlobWithDepth(root string, pattern string) ([]string, error) {
	matches := []string{}
	err := filepathWalk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			if match, _ := filepathMatch(pattern, filepath.Base(path)); match {
				matches = append(matches, path)
			}
		}
		return nil
	})

	if err != nil {
		return []string{}, err
	}

	return matches, nil
}

// CreateZipFile creates a zip archive from all files and directories within a specified source directory.
// The resulting zip file is stored at the 'target' path. Each entry in the zip file corresponds to a file
// or directory from the source, where directory entries end with a path separator.
//
// The function walks through the source directory, compresses files using the Deflate algorithm, and
// adds them to the zip archive. Directories are added to the archive with appropriate markers but no content.
// Errors during file handling (e.g., file cannot be opened, read, or written) or archive creation are returned.
//
// Parameters:
//   - source: The directory whose contents are to be zipped.
//   - target: The file path where the zip archive will be saved.
//
// Returns:
//   - error: Non-nil error if there was a problem creating the zip file or processing its contents.
func CreateZipFile(source, target string) error {
	source += "/"

	f, err := osCreate(target)
	if err != nil {
		return err
	}
	defer closeFile(f)

	writer := newZipWriter(f)
	defer closeCloser(writer)

	return filepathWalk(source, func(path string, info os.FileInfo, err error) error {
		if source != path {
			if err != nil {
				return err
			}

			header, err := zipFileInfo(info)
			if err != nil {
				return err
			}

			header.Method = zip.Deflate

			header.Name, err = filepathRel(filepath.Dir(source), path)
			if err != nil {
				return err
			}

			if info.IsDir() {
				header.Name += string(os.PathSeparator)
			}

			headerWriter, err := writer.CreateHeader(header)
			if err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			f, err := osOpen(path)
			if err != nil {
				return err
			}
			defer closeFile(f)

			_, err = ioCopy(headerWriter, f)
			return err
		}

		return nil
	})
}

// UnzipFile extracts the contents of a zip archive to a specified destination directory.
// It creates directories and files as needed to match the structure of the archive.
//
// Parameters:
//   - source: The path to the zip archive to be extracted.
//   - target: The directory where the contents of the archive will be extracted.
//
// Returns:
//   - error: Non-nil error if there was a problem reading the archive or writing its contents.
func UnzipFile(source, target string) error {
	reader, err := zipOpenReader(source)
	if err != nil {
		return fmt.Errorf("error opening zip file: %w", err)
	}
	defer closeCloser(reader)

	targetRoot, err := filepathAbs(target)
	if err != nil {
		return fmt.Errorf("error resolving target path: %w", err)
	}

	for _, file := range reader.File {
		path, err := filepathAbs(filepath.Join(targetRoot, file.Name)) // #nosec G305 -- zip entry path is validated against targetRoot before use.
		if err != nil {
			return fmt.Errorf("error resolving zip entry path: %w", err)
		}

		if path != targetRoot && !strings.HasPrefix(path, targetRoot+string(os.PathSeparator)) {
			return fmt.Errorf("illegal target path: %s", path)
		}

		if file.FileInfo().IsDir() {
			err := osMkdirAll(path, 0750)
			if err != nil {
				return fmt.Errorf("error making directory: %w", err)
			}
			continue
		}

		err = osMkdirAll(filepath.Dir(path), 0750)
		if err != nil {
			return fmt.Errorf("error making directory: %w", err)
		}

		if err := func() error {
			destFile, err := osOpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
			if err != nil {
				return fmt.Errorf("error opening file for writing: %w", err)
			}
			defer closeFile(destFile)

			srcFile, err := zipFileOpen(file)
			if err != nil {
				return fmt.Errorf("error opening file in zip: %w", err)
			}
			defer closeCloser(srcFile)

			_, err = ioCopy(destFile, srcFile)
			if err != nil {
				return fmt.Errorf("error copying file content: %w", err)
			}

			return nil
		}(); err != nil {
			return err
		}
	}

	return nil
}

// IsReadableFile checks if file exists, and is readable.
// Params:
//   - path: path of the file to be checked.
//
// Returns:
//   - bool: true if file is readable, false if not file is not readable
//   - error: Non-nil error if the file is not readable due to various reasons
func IsReadableFile(path string) (bool, error) {
	fileInfo, err := osStat(path)

	if err != nil {
		if os.IsNotExist(err) {
			return false, errors.New("file does not exist at provided path")
		} else {
			return false, fmt.Errorf("error getting fileInfo: %v", err.Error())
		}
	} else {
		mode := fileInfo.Mode()
		if mode.IsDir() {
			return false, errors.New("directory found instead of file at provided path")
		}

		if mode.IsRegular() && (mode.Perm()&400) != 0 {
			return true, nil
		}

		return false, errors.New("file exists, but not readable")
	}
}

// GetFileChecksum computes and returns the SHA-256 checksum of the file at the given path.
//
// The function reads the entire content of the file and calculates its SHA-256 hash,
// returning the result as a hexadecimal-encoded string.
//
// Parameters:
//   - path: The full path to the file to be hashed.
//
// Returns:
//   - string: The SHA-256 hash of the file as a hex string.
//   - error:  An error if the file cannot be opened or read.
func GetFileChecksum(path string) (string, error) {
	file, err := osOpen(path)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer closeFile(file)

	hasher := sha256.New()
	if _, err := ioCopy(hasher, file); err != nil {
		return "", fmt.Errorf("failed to read file for hashing: %w", err)
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// IsSymlink checks whether the given path is a symbolic link.
//
// The function uses os.Lstat to inspect the file metadata without following symlinks.
// If the file mode includes os.ModeSymlink, it returns true.
//
// Parameters:
//   - path: The file or directory path to check.
//
// Returns:
//   - bool:  True if the path is a symbolic link.
//   - error: An error if the path does not exist or cannot be inspected.
func IsSymlink(path string) (bool, error) {
	fi, err := osLstat(path)
	if err != nil {
		return false, err
	}
	return fi.Mode()&os.ModeSymlink != 0, nil
}

// CreateDirIfNotExists ensures that a directory exists at the specified path.
// If the directory does not exist, it creates it with the specified permissions.
// If the path exists but is not a directory, an error is returned.
//
// Parameters:
//   - path: The file system path where the directory should exist.
//   - perm: The file mode permissions to use when creating the directory.
//
// Returns:
//   - error: An error if the path exists but is not a directory, if there is an
//     issue creating the directory, or any other file system-related error.
func CreateDirIfNotExists(path string, perm os.FileMode) error {
	info, err := osStat(path)
	if err == nil {
		if !info.IsDir() {
			return fmt.Errorf("path exists and is not a directory: %s", path)
		}
		return nil
	}
	if os.IsNotExist(err) {
		return osMkdirAll(path, perm)
	}
	return err
}

// SanitizeFilename keeps UTF-8 characters but removes/neutralizes header-unsafe ones,
// returning a safe filename and a percent-encoded variant suitable for `filename*`.
// - Replaces: " \ \r \n  → '_'
// - Strips: control chars (<0x20) and DEL (0x7F)
// - Falls back to defaultName if nothing remains
func SanitizeFilename(name, defaultName string) (safe string, encoded string) {
	var b strings.Builder
	for _, r := range name {
		switch {
		case r == '"' || r == '\\' || r == '\r' || r == '\n':
			b.WriteRune('_')
		case r < 0x20 || r == 0x7F: // control chars
			// skip
		default:
			b.WriteRune(r) // keep valid UTF-8 as-is
		}
	}
	if b.Len() == 0 {
		return defaultName, url.PathEscape(defaultName)
	}
	s := b.String()
	return s, url.PathEscape(s)
}
