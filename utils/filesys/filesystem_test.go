package filesys

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fsHooks struct {
	osStat        func(string) (os.FileInfo, error)
	osMkdirAll    func(string, os.FileMode) error
	osWriteFile   func(string, []byte, os.FileMode) error
	osOpen        func(string) (*os.File, error)
	osCreate      func(string) (*os.File, error)
	osRename      func(string, string) error
	osRemove      func(string) error
	osLstat       func(string) (os.FileInfo, error)
	osOpenFile    func(string, int, os.FileMode) (*os.File, error)
	fileSync      func(*os.File) error
	fileClose     func(*os.File) error
	ioCopy        func(io.Writer, io.Reader) (int64, error)
	zipOpenReader func(string) (*zip.ReadCloser, error)
	zipFileInfo   func(os.FileInfo) (*zip.FileHeader, error)
	filepathWalk  func(string, filepath.WalkFunc) error
	filepathMatch func(string, string) (bool, error)
	filepathRel   func(string, string) (string, error)
	filepathAbs   func(string) (string, error)
	zipFileOpen   func(*zip.File) (io.ReadCloser, error)
	newZipWriter  func(io.Writer) zipWriter
}

type errCloser struct{}

func (errCloser) Close() error {
	return errors.New("close failed")
}

func snapshotFSHooks() fsHooks {
	return fsHooks{
		osStat:        osStat,
		osMkdirAll:    osMkdirAll,
		osWriteFile:   osWriteFile,
		osOpen:        osOpen,
		osCreate:      osCreate,
		osRename:      osRename,
		osRemove:      osRemove,
		osLstat:       osLstat,
		osOpenFile:    osOpenFile,
		fileSync:      fileSync,
		fileClose:     fileClose,
		ioCopy:        ioCopy,
		zipOpenReader: zipOpenReader,
		zipFileInfo:   zipFileInfo,
		filepathWalk:  filepathWalk,
		filepathMatch: filepathMatch,
		filepathRel:   filepathRel,
		filepathAbs:   filepathAbs,
		zipFileOpen:   zipFileOpen,
		newZipWriter:  newZipWriter,
	}
}

func (h fsHooks) restore() {
	osStat = h.osStat
	osMkdirAll = h.osMkdirAll
	osWriteFile = h.osWriteFile
	osOpen = h.osOpen
	osCreate = h.osCreate
	osRename = h.osRename
	osRemove = h.osRemove
	osLstat = h.osLstat
	osOpenFile = h.osOpenFile
	fileSync = h.fileSync
	fileClose = h.fileClose
	ioCopy = h.ioCopy
	zipOpenReader = h.zipOpenReader
	zipFileInfo = h.zipFileInfo
	filepathWalk = h.filepathWalk
	filepathMatch = h.filepathMatch
	filepathRel = h.filepathRel
	filepathAbs = h.filepathAbs
	zipFileOpen = h.zipFileOpen
	newZipWriter = h.newZipWriter
}

func TestWriteFile(t *testing.T) {
	t.Run("SuccessfulWriteOnDirExists", func(t *testing.T) {
		tmpDir := t.TempDir()
		file := filepath.Join(tmpDir, "testfile.txt")
		err := WriteFile(file, []byte("hello"), 0644)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("failed to read back file: %v", err)
		}
		if string(data) != "hello" {
			t.Errorf("expected 'hello', got '%s'", data)
		}
	})

	t.Run("CreateMissingDir", func(t *testing.T) {
		tmpDir := t.TempDir()
		nestedDir := filepath.Join(tmpDir, "a", "b", "c")
		file := filepath.Join(nestedDir, "nested.txt")
		err := WriteFile(file, []byte("nested"), 0755)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
	})

	t.Run("FailToCreatetDir", func(t *testing.T) {
		tmpDir := t.TempDir()
		blocker := filepath.Join(tmpDir, "dir-as-file")
		_ = os.WriteFile(blocker, []byte("not a dir"), 0644)

		file := filepath.Join(blocker, "will-fail.txt")
		err := WriteFile(file, []byte("x"), 0644)
		if err == nil || !errors.Is(err, os.ErrInvalid) && !os.IsNotExist(err) {
			t.Logf("Expected dir creation error, got: %v", err)
		}
	})

	t.Run("FailsToWriteFile", func(t *testing.T) {
		tmpDir := t.TempDir()

		readOnlyDir := filepath.Join(tmpDir, "readonly")
		_ = os.MkdirAll(readOnlyDir, 0500)

		file := filepath.Join(readOnlyDir, "no-write.txt")
		err := WriteFile(file, []byte("fail"), 0644)

		if err == nil {
			t.Fatal("expected error while writing to read-only dir, got nil")
		}
	})

	t.Run("MkdirAllError", func(t *testing.T) {
		hooks := snapshotFSHooks()
		t.Cleanup(hooks.restore)

		osStat = func(string) (os.FileInfo, error) { return nil, os.ErrNotExist }
		osMkdirAll = func(string, os.FileMode) error { return errors.New("mkdir fail") }

		err := WriteFile("any/path.txt", []byte("x"), 0644)
		require.Error(t, err)
		require.Contains(t, err.Error(), "error making directory")
	})

	t.Run("WriteFileError", func(t *testing.T) {
		hooks := snapshotFSHooks()
		t.Cleanup(hooks.restore)

		osStat = func(string) (os.FileInfo, error) { return nil, nil }
		osWriteFile = func(string, []byte, os.FileMode) error { return errors.New("write fail") }

		err := WriteFile("any/path.txt", []byte("x"), 0644)
		require.Error(t, err)
		require.Contains(t, err.Error(), "error writing pdf")
	})
}

func TestCloseCloser_IgnoreCloseError(t *testing.T) {
	closeCloser(errCloser{})
}

func TestRemoveFile_IgnoreRemoveError(t *testing.T) {
	hooks := snapshotFSHooks()
	t.Cleanup(hooks.restore)

	osRemove = func(string) error {
		return errors.New("remove failed")
	}

	removeFile("missing")
}

func TestAtomicCopyFileContent(t *testing.T) {
	t.Run("SuccessfulCopy", func(t *testing.T) {
		dir := t.TempDir()
		src := filepath.Join(dir, "src.txt")
		dest := filepath.Join(dir, "dest.txt")

		require.NoError(t, os.WriteFile(src, []byte("hello world"), 0644))
		require.NoError(t, AtomicCopyFileContent(src, dest))

		content, err := os.ReadFile(dest)
		require.NoError(t, err)
		require.Equal(t, "hello world", string(content))
	})

	t.Run("ErrorOnSrcFileNotFound", func(t *testing.T) {
		err := AtomicCopyFileContent("non-existent.txt", "any.txt")
		require.Error(t, err)
	})

	t.Run("ErrorFailToCreatTmpFile", func(t *testing.T) {
		dir := t.TempDir()
		src := filepath.Join(dir, "src.txt")
		block := filepath.Join(dir, "blocked")
		dest := filepath.Join(block, "dest.txt")

		require.NoError(t, os.WriteFile(src, []byte("data"), 0644))
		require.NoError(t, os.WriteFile(block, []byte("file not dir"), 0644)) // block path as file

		err := AtomicCopyFileContent(src, dest)
		require.Error(t, err)
	})

	t.Run("ErrorIfCopyFails", func(t *testing.T) {
		dir := t.TempDir()
		src := filepath.Join(dir, "src.txt")
		dest := filepath.Join(dir, "dest.txt")

		require.NoError(t, os.WriteFile(src, []byte("abc"), 0000)) // unreadable

		err := AtomicCopyFileContent(src, dest)
		require.Error(t, err)
	})

	t.Run("ErrorIfSyncFails", func(t *testing.T) {
		// Cannot simulate tmpFile.Sync() error without a mock FS or syscall trick.
		// Instead, test safe fallback cleanup by denying rename:
		dir := t.TempDir()
		src := filepath.Join(dir, "src.txt")
		dest := filepath.Join(dir, "dest.txt")

		require.NoError(t, os.WriteFile(src, []byte("abc"), 0644))
		require.NoError(t, os.WriteFile(dest+".tmp", []byte("block"), 0444))

		err := AtomicCopyFileContent(src, dest)
		require.Error(t, err)
		_ = os.Remove(dest + ".tmp")
	})

	t.Run("ErrorOnRenameFail", func(t *testing.T) {
		dir := t.TempDir()
		src := filepath.Join(dir, "src.txt")
		dest := filepath.Join(dir, "readonly", "dest.txt")

		require.NoError(t, os.WriteFile(src, []byte("xyz"), 0644))
		require.NoError(t, os.MkdirAll(filepath.Dir(dest), 0000)) // no write permission

		err := AtomicCopyFileContent(src, dest)
		require.Error(t, err)
	})

	t.Run("ErrorOnCopyFail", func(t *testing.T) {
		hooks := snapshotFSHooks()
		t.Cleanup(hooks.restore)

		dir := t.TempDir()
		src := filepath.Join(dir, "src.txt")
		dest := filepath.Join(dir, "dest.txt")

		require.NoError(t, os.WriteFile(src, []byte("data"), 0644))

		ioCopy = func(io.Writer, io.Reader) (int64, error) {
			return 0, errors.New("copy fail")
		}

		err := AtomicCopyFileContent(src, dest)
		require.Error(t, err)
	})

	t.Run("ErrorOnSyncFail", func(t *testing.T) {
		hooks := snapshotFSHooks()
		t.Cleanup(hooks.restore)

		dir := t.TempDir()
		src := filepath.Join(dir, "src.txt")
		dest := filepath.Join(dir, "dest.txt")

		require.NoError(t, os.WriteFile(src, []byte("data"), 0644))

		fileSync = func(*os.File) error { return errors.New("sync fail") }

		err := AtomicCopyFileContent(src, dest)
		require.Error(t, err)
	})

	t.Run("ErrorOnCloseFail", func(t *testing.T) {
		hooks := snapshotFSHooks()
		t.Cleanup(hooks.restore)

		dir := t.TempDir()
		src := filepath.Join(dir, "src.txt")
		dest := filepath.Join(dir, "dest.txt")

		require.NoError(t, os.WriteFile(src, []byte("data"), 0644))

		fileClose = func(*os.File) error { return errors.New("close fail") }

		err := AtomicCopyFileContent(src, dest)
		require.Error(t, err)
	})

	t.Run("ErrorOnRenameFailInjected", func(t *testing.T) {
		hooks := snapshotFSHooks()
		t.Cleanup(hooks.restore)

		dir := t.TempDir()
		src := filepath.Join(dir, "src.txt")
		dest := filepath.Join(dir, "dest.txt")

		require.NoError(t, os.WriteFile(src, []byte("data"), 0644))

		osRename = func(string, string) error { return errors.New("rename fail") }

		err := AtomicCopyFileContent(src, dest)
		require.Error(t, err)
	})
}

func TestIsDirectory(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "testdir")
	require.NoError(t, err)
	defer func() {
		require.NoError(t, os.RemoveAll(tmpDir))
	}()

	tmpFile, err := os.CreateTemp(tmpDir, "testfile")
	require.NoError(t, err)
	require.NoError(t, tmpFile.Close())
	defer func() {
		require.NoError(t, os.Remove(tmpFile.Name()))
	}()

	isDir, err := IsDirectory(tmpDir)
	require.NoError(t, err)
	assert.True(t, isDir, "Expected directory to be identified as a directory")

	isDir, err = IsDirectory(tmpFile.Name())
	require.NoError(t, err)
	assert.False(t, isDir, "Expected file to not be identified as a directory")

	isDir, err = IsDirectory("/nonexistent")
	assert.Error(t, err)
	assert.False(t, isDir, "Expected non-existent path to return error and false")
}

func TestGlobWithDepth(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test")
	require.NoError(t, err)
	defer func() {
		require.NoError(t, os.RemoveAll(tmpDir))
	}()

	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "file.txt"), []byte("content"), 0644))
	require.NoError(t, os.Mkdir(filepath.Join(tmpDir, "subdir"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "subdir", "file2.txt"), []byte("content"), 0644))

	matches, err := GlobWithDepth(tmpDir, "*.txt")
	require.NoError(t, err)
	assert.Len(t, matches, 2, "Expected to find two txt files")

	matches, err = GlobWithDepth(tmpDir, "*.jpg")
	require.NoError(t, err)
	assert.Empty(t, matches, "Expected no files to match the jpg pattern")
}

func TestGlobWithDepth_WalkError(t *testing.T) {
	hooks := snapshotFSHooks()
	t.Cleanup(hooks.restore)

	filepathWalk = func(string, filepath.WalkFunc) error {
		return errors.New("walk fail")
	}

	matches, err := GlobWithDepth("any", "*.txt")
	require.Error(t, err)
	require.Empty(t, matches)
}

func TestGlobWithDepth_VisitorError(t *testing.T) {
	hooks := snapshotFSHooks()
	t.Cleanup(hooks.restore)

	filepathWalk = func(root string, fn filepath.WalkFunc) error {
		return fn(root, nil, errors.New("visit fail"))
	}

	_, err := GlobWithDepth("any", "*.txt")
	require.Error(t, err)
}

type fakeFileInfo struct {
	name  string
	size  int64
	mode  os.FileMode
	isDir bool
}

func (f fakeFileInfo) Name() string       { return f.name }
func (f fakeFileInfo) Size() int64        { return f.size }
func (f fakeFileInfo) Mode() os.FileMode  { return f.mode }
func (f fakeFileInfo) ModTime() time.Time { return time.Time{} }
func (f fakeFileInfo) IsDir() bool        { return f.isDir }
func (f fakeFileInfo) Sys() interface{}   { return nil }

func TestCreateZipFile(t *testing.T) {
	sourceDir, err := os.MkdirTemp("", "source")
	require.NoError(t, err)
	defer func() {
		require.NoError(t, os.RemoveAll(sourceDir))
	}()

	targetZip := filepath.Join(sourceDir, "target.zip")
	require.NoError(t, os.WriteFile(filepath.Join(sourceDir, "file.txt"), []byte("content"), 0644))

	err = CreateZipFile(sourceDir, targetZip)
	require.NoError(t, err)

	file, err := os.Open(targetZip)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, file.Close())
	}()

	info, err := file.Stat()
	require.NoError(t, err)

	zipReader, err := zip.NewReader(file, info.Size())
	require.NoError(t, err)

	found := false
	for _, f := range zipReader.File {
		if f.Name == "file.txt" {
			found = true
			rc, err := f.Open()
			require.NoError(t, err)
			content, err := io.ReadAll(rc)
			require.NoError(t, rc.Close())
			require.NoError(t, err)
			assert.Equal(t, "content", string(content))
		}
	}
	assert.True(t, found, "Expected file.txt to be in the zip archive")
}

func TestCreateZipFile_Errors(t *testing.T) {
	t.Run("CreateError", func(t *testing.T) {
		hooks := snapshotFSHooks()
		t.Cleanup(hooks.restore)

		osCreate = func(string) (*os.File, error) {
			return nil, errors.New("create fail")
		}

		err := CreateZipFile("source", "target.zip")
		require.Error(t, err)
	})

	t.Run("FileInfoError", func(t *testing.T) {
		hooks := snapshotFSHooks()
		t.Cleanup(hooks.restore)

		filepathWalk = func(root string, fn filepath.WalkFunc) error {
			info := fakeFileInfo{name: "file.txt", mode: 0644}
			return fn(filepath.Join(root, "file.txt"), info, nil)
		}
		zipFileInfo = func(os.FileInfo) (*zip.FileHeader, error) {
			return nil, errors.New("fileinfo fail")
		}

		tmp := t.TempDir()
		err := CreateZipFile(tmp, filepath.Join(tmp, "out.zip"))
		require.Error(t, err)
	})

	t.Run("RelError", func(t *testing.T) {
		hooks := snapshotFSHooks()
		t.Cleanup(hooks.restore)

		filepathWalk = func(root string, fn filepath.WalkFunc) error {
			info := fakeFileInfo{name: "file.txt", mode: 0644}
			return fn(filepath.Join(root, "file.txt"), info, nil)
		}
		filepathRel = func(string, string) (string, error) {
			return "", errors.New("rel fail")
		}

		tmp := t.TempDir()
		err := CreateZipFile(tmp, filepath.Join(tmp, "out.zip"))
		require.Error(t, err)
	})

	t.Run("CreateHeaderError", func(t *testing.T) {
		hooks := snapshotFSHooks()
		t.Cleanup(hooks.restore)

		filepathWalk = func(root string, fn filepath.WalkFunc) error {
			info := fakeFileInfo{name: "file.txt", mode: 0644}
			return fn(filepath.Join(root, "file.txt"), info, nil)
		}
		newZipWriter = func(io.Writer) zipWriter {
			return &mockZipWriter{createErr: errors.New("header fail")}
		}

		tmp := t.TempDir()
		err := CreateZipFile(tmp, filepath.Join(tmp, "out.zip"))
		require.Error(t, err)
	})

	t.Run("OpenError", func(t *testing.T) {
		hooks := snapshotFSHooks()
		t.Cleanup(hooks.restore)

		tmp := t.TempDir()
		source := filepath.Join(tmp, "src")
		require.NoError(t, os.MkdirAll(source, 0755))
		require.NoError(t, os.WriteFile(filepath.Join(source, "file.txt"), []byte("x"), 0644))

		osOpen = func(string) (*os.File, error) {
			return nil, errors.New("open fail")
		}

		err := CreateZipFile(source, filepath.Join(tmp, "out.zip"))
		require.Error(t, err)
	})

	t.Run("CopyError", func(t *testing.T) {
		hooks := snapshotFSHooks()
		t.Cleanup(hooks.restore)

		tmp := t.TempDir()
		source := filepath.Join(tmp, "src")
		require.NoError(t, os.MkdirAll(source, 0755))
		require.NoError(t, os.WriteFile(filepath.Join(source, "file.txt"), []byte("x"), 0644))

		ioCopy = func(io.Writer, io.Reader) (int64, error) {
			return 0, errors.New("copy fail")
		}

		err := CreateZipFile(source, filepath.Join(tmp, "out.zip"))
		require.Error(t, err)
	})
}

type mockZipWriter struct {
	createErr error
}

func (m *mockZipWriter) CreateHeader(*zip.FileHeader) (io.Writer, error) {
	return nil, m.createErr
}

func (m *mockZipWriter) Close() error {
	return nil
}

func TestCreateZipFile_DirectoryEntry(t *testing.T) {
	tmp := t.TempDir()
	source := filepath.Join(tmp, "src")
	require.NoError(t, os.MkdirAll(filepath.Join(source, "dir"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(source, "dir", "file.txt"), []byte("x"), 0644))

	err := CreateZipFile(source, filepath.Join(tmp, "out.zip"))
	require.NoError(t, err)
}

func TestCreateZipFile_WalkError(t *testing.T) {
	hooks := snapshotFSHooks()
	t.Cleanup(hooks.restore)

	filepathWalk = func(root string, fn filepath.WalkFunc) error {
		info := fakeFileInfo{name: "file.txt", mode: 0644}
		return fn(filepath.Join(root, "file.txt"), info, errors.New("walk fail"))
	}

	tmp := t.TempDir()
	err := CreateZipFile(tmp, filepath.Join(tmp, "out.zip"))
	require.Error(t, err)
}

func TestIsReadableFile_AllCases(t *testing.T) {
	// Case 1: File does not exist
	ok, err := IsReadableFile("non_existent_file.txt")
	if ok || err == nil || err.Error() != "file does not exist at provided path" {
		t.Errorf("Expected error for non-existent file, got ok=%v, err=%v", ok, err)
	}

	// Case 2: Path is a directory
	dir := t.TempDir()
	ok, err = IsReadableFile(dir)
	if ok || err == nil || err.Error() != "directory found instead of file at provided path" {
		t.Errorf("Expected error for directory, got ok=%v, err=%v", ok, err)
	}

	// Case 3: Readable file
	readableFile := filepath.Join(dir, "readable.txt")
	err = os.WriteFile(readableFile, []byte("hello"), 0644)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}
	ok, err = IsReadableFile(readableFile)
	if !ok || err != nil {
		t.Errorf("Expected readable file, got ok=%v, err=%v", ok, err)
	}

	// Case 4: Non-readable file
	unreadableFile := filepath.Join(dir, "unreadable.txt")
	err = os.WriteFile(unreadableFile, []byte("restricted"), 0000)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}
	defer func() {
		require.NoError(t, os.Chmod(unreadableFile, 0644))
	}()
	ok, err = IsReadableFile(unreadableFile)
	if ok || err == nil || err.Error() != "file exists, but not readable" {
		t.Errorf("Expected unreadable file error, got ok=%v, err=%v", ok, err)
	}
}

func TestIsReadableFile_StatError(t *testing.T) {
	hooks := snapshotFSHooks()
	t.Cleanup(hooks.restore)

	osStat = func(string) (os.FileInfo, error) {
		return nil, errors.New("stat fail")
	}

	ok, err := IsReadableFile("any")
	require.Error(t, err)
	require.False(t, ok)
}

func TestGetFileChecksum(t *testing.T) {
	t.Run("returns checksum for valid file", func(t *testing.T) {
		tmp := t.TempDir()
		file := filepath.Join(tmp, "data.txt")
		content := []byte("hello world")
		require.NoError(t, os.WriteFile(file, content, 0644))

		hash, err := GetFileChecksum(file)
		require.NoError(t, err)

		expected := sha256.Sum256(content)
		require.Equal(t, hex.EncodeToString(expected[:]), hash)
	})

	t.Run("returns error for missing file", func(t *testing.T) {
		_, err := GetFileChecksum("nonexistent.txt")
		require.Error(t, err)
	})

	t.Run("returns error for read failure", func(t *testing.T) {
		hooks := snapshotFSHooks()
		t.Cleanup(hooks.restore)

		tmp := t.TempDir()
		file := filepath.Join(tmp, "data.txt")
		require.NoError(t, os.WriteFile(file, []byte("content"), 0644))

		ioCopy = func(io.Writer, io.Reader) (int64, error) {
			return 0, errors.New("read fail")
		}

		_, err := GetFileChecksum(file)
		require.Error(t, err)
	})
}

func TestIsSymlink(t *testing.T) {
	t.Run("returns true for a symlink", func(t *testing.T) {
		tmp := t.TempDir()
		target := filepath.Join(tmp, "real.txt")
		symlink := filepath.Join(tmp, "link.txt")

		require.NoError(t, os.WriteFile(target, []byte("data"), 0644))
		require.NoError(t, os.Symlink(target, symlink))

		ok, err := IsSymlink(symlink)
		require.NoError(t, err)
		require.True(t, ok)
	})

	t.Run("returns false for regular file", func(t *testing.T) {
		tmp := t.TempDir()
		regular := filepath.Join(tmp, "file.txt")
		require.NoError(t, os.WriteFile(regular, []byte("x"), 0644))

		ok, err := IsSymlink(regular)
		require.NoError(t, err)
		require.False(t, ok)
	})

	t.Run("returns error if file is missing", func(t *testing.T) {
		_, err := IsSymlink("missing.txt")
		require.Error(t, err)
	})
}

func TestCreateDirIfNotExists(t *testing.T) {
	t.Run("creates directory if not exists", func(t *testing.T) {
		tmp := t.TempDir()
		target := filepath.Join(tmp, "newdir")

		err := CreateDirIfNotExists(target, 0755)
		require.NoError(t, err)

		info, err := os.Stat(target)
		require.NoError(t, err)
		require.True(t, info.IsDir())
	})

	t.Run("returns nil if directory already exists", func(t *testing.T) {
		tmp := t.TempDir()
		err := CreateDirIfNotExists(tmp, 0755)
		require.NoError(t, err)
	})

	t.Run("returns error if path exists but is not a directory", func(t *testing.T) {
		tmp := t.TempDir()
		file := filepath.Join(tmp, "file.txt")
		require.NoError(t, os.WriteFile(file, []byte("x"), 0644))

		err := CreateDirIfNotExists(file, 0755)
		require.Error(t, err)
	})

	t.Run("returns error if parent dir doesn’t exist and can’t be created", func(t *testing.T) {
		tmp := t.TempDir()
		blocker := filepath.Join(tmp, "block")
		require.NoError(t, os.WriteFile(blocker, []byte("x"), 0644))

		badPath := filepath.Join(blocker, "fail")
		err := CreateDirIfNotExists(badPath, 0755)
		require.Error(t, err)
	})

	t.Run("returns error for stat failure", func(t *testing.T) {
		hooks := snapshotFSHooks()
		t.Cleanup(hooks.restore)

		osStat = func(string) (os.FileInfo, error) {
			return nil, errors.New("stat fail")
		}

		err := CreateDirIfNotExists("any", 0755)
		require.Error(t, err)
	})
}

type mockZipFile struct {
	name    string
	isDir   bool
	content string
	mode    os.FileMode
}

func createTestZip(t *testing.T, path string, files []mockZipFile) {
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)

	for _, file := range files {
		hdr := &zip.FileHeader{
			Name: file.name,
		}
		hdr.SetMode(file.mode)
		if file.isDir {
			hdr.Name += "/"
		}
		f, err := w.CreateHeader(hdr)
		if err != nil {
			t.Fatalf("error creating header: %v", err)
		}
		if !file.isDir {
			_, err := f.Write([]byte(file.content))
			if err != nil {
				t.Fatalf("error writing content: %v", err)
			}
		}
	}

	if err := w.Close(); err != nil {
		t.Fatalf("error closing writer: %v", err)
	}
	if err := os.WriteFile(path, buf.Bytes(), 0644); err != nil {
		t.Fatalf("error writing zip to file: %v", err)
	}
}

func TestUnzipFile_AllCases(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("Invalid zip path", func(t *testing.T) {
		err := UnzipFile("/invalid/path.zip", tempDir)
		if err == nil || !strings.Contains(err.Error(), "error opening zip file") {
			t.Errorf("Expected error for invalid zip path, got: %v", err)
		}
	})

	t.Run("Zip slip attack", func(t *testing.T) {
		zipPath := filepath.Join(tempDir, "zip_slip.zip")
		createTestZip(t, zipPath, []mockZipFile{
			{name: "../malicious.txt", content: "bad", mode: 0644},
		})
		err := UnzipFile(zipPath, tempDir)
		if err == nil || !strings.Contains(err.Error(), "illegal target path") {
			t.Errorf("Expected illegal target path error, got: %v", err)
		}
	})

	t.Run("Directory creation error", func(t *testing.T) {
		zipPath := filepath.Join(tempDir, "mkdirfail.zip")
		createTestZip(t, zipPath, []mockZipFile{
			{name: "nested/fail.txt", content: "oops", mode: 0644},
		})
		protected := filepath.Join(tempDir, "readonly")
		if err := os.MkdirAll(protected, 0000); err != nil {
			t.Fatalf("could not make protected dir: %v", err)
		}
		defer func() {
			require.NoError(t, os.Chmod(protected, 0755))
		}()
		err := UnzipFile(zipPath, protected)
		if err == nil || !strings.Contains(err.Error(), "error making directory") {
			t.Errorf("Expected mkdir error, got: %v", err)
		}
	})

	t.Run("Successful unzip", func(t *testing.T) {
		zipPath := filepath.Join(tempDir, "ok.zip")
		createTestZip(t, zipPath, []mockZipFile{
			{name: "file1.txt", content: "hello", mode: 0644},
			{name: "dir/", isDir: true, mode: 0755},
			{name: "dir/file2.txt", content: "world", mode: 0644},
		})
		target := filepath.Join(tempDir, "out")
		err := UnzipFile(zipPath, target)
		if err != nil {
			t.Errorf("Unexpected error on valid unzip: %v", err)
		}
		if _, err := os.Stat(filepath.Join(target, "file1.txt")); err != nil {
			t.Errorf("Expected file1.txt, but not found")
		}
		if _, err := os.Stat(filepath.Join(target, "dir/file2.txt")); err != nil {
			t.Errorf("Expected file2.txt, but not found")
		}
	})

	t.Run("OpenFileError", func(t *testing.T) {
		hooks := snapshotFSHooks()
		t.Cleanup(hooks.restore)

		zipPath := filepath.Join(tempDir, "openfail.zip")
		createTestZip(t, zipPath, []mockZipFile{
			{name: "file.txt", content: "hello", mode: 0644},
		})

		osOpenFile = func(string, int, os.FileMode) (*os.File, error) {
			return nil, errors.New("open file fail")
		}

		err := UnzipFile(zipPath, filepath.Join(tempDir, "out-open"))
		require.Error(t, err)
	})

	t.Run("ZipFileOpenError", func(t *testing.T) {
		hooks := snapshotFSHooks()
		t.Cleanup(hooks.restore)

		zipPath := filepath.Join(tempDir, "fileopenfail.zip")
		createTestZip(t, zipPath, []mockZipFile{
			{name: "file.txt", content: "hello", mode: 0644},
		})

		zipFileOpen = func(*zip.File) (io.ReadCloser, error) {
			return nil, errors.New("zip open fail")
		}

		err := UnzipFile(zipPath, filepath.Join(tempDir, "out-zipopen"))
		require.Error(t, err)
	})

	t.Run("CopyError", func(t *testing.T) {
		hooks := snapshotFSHooks()
		t.Cleanup(hooks.restore)

		zipPath := filepath.Join(tempDir, "copyfail.zip")
		createTestZip(t, zipPath, []mockZipFile{
			{name: "file.txt", content: "hello", mode: 0644},
		})

		ioCopy = func(io.Writer, io.Reader) (int64, error) {
			return 0, errors.New("copy fail")
		}

		err := UnzipFile(zipPath, filepath.Join(tempDir, "out-copy"))
		require.Error(t, err)
	})

	t.Run("MkdirAllError", func(t *testing.T) {
		hooks := snapshotFSHooks()
		t.Cleanup(hooks.restore)

		zipPath := filepath.Join(tempDir, "mkdirfail2.zip")
		createTestZip(t, zipPath, []mockZipFile{
			{name: "dir/file.txt", content: "hello", mode: 0644},
		})

		osMkdirAll = func(string, os.FileMode) error {
			return errors.New("mkdir fail")
		}

		err := UnzipFile(zipPath, filepath.Join(tempDir, "out-mkdir"))
		require.Error(t, err)
	})

	t.Run("DirEntryMkdirError", func(t *testing.T) {
		hooks := snapshotFSHooks()
		t.Cleanup(hooks.restore)

		zipPath := filepath.Join(tempDir, "direntry.zip")
		createTestZip(t, zipPath, []mockZipFile{
			{name: "dir/", isDir: true, mode: 0755},
		})

		osMkdirAll = func(string, os.FileMode) error {
			return errors.New("mkdir fail")
		}

		err := UnzipFile(zipPath, filepath.Join(tempDir, "out-dir"))
		require.Error(t, err)
	})

	t.Run("TargetAbsError", func(t *testing.T) {
		hooks := snapshotFSHooks()
		t.Cleanup(hooks.restore)

		zipPath := filepath.Join(tempDir, "targetabs.zip")
		createTestZip(t, zipPath, []mockZipFile{
			{name: "file.txt", content: "hello", mode: 0644},
		})

		filepathAbs = func(string) (string, error) {
			return "", errors.New("target abs fail")
		}

		err := UnzipFile(zipPath, filepath.Join(tempDir, "out-target-abs"))
		require.Error(t, err)
		require.Contains(t, err.Error(), "error resolving target path")
	})

	t.Run("ZipEntryAbsError", func(t *testing.T) {
		hooks := snapshotFSHooks()
		t.Cleanup(hooks.restore)

		zipPath := filepath.Join(tempDir, "entryabs.zip")
		createTestZip(t, zipPath, []mockZipFile{
			{name: "file.txt", content: "hello", mode: 0644},
		})

		calls := 0
		filepathAbs = func(path string) (string, error) {
			calls++
			if calls == 1 {
				return hooks.filepathAbs(path)
			}
			return "", errors.New("entry abs fail")
		}

		err := UnzipFile(zipPath, filepath.Join(tempDir, "out-entry-abs"))
		require.Error(t, err)
		require.Contains(t, err.Error(), "error resolving zip entry path")
	})
}

func TestSanitizeFilename_Basic(t *testing.T) {
	name := "report.txt"
	safe, enc := SanitizeFilename(name, "download")
	if safe != name {
		t.Fatalf("safe mismatch: got %q want %q", safe, name)
	}
	wantEnc := url.PathEscape(name)
	if enc != wantEnc {
		t.Fatalf("enc mismatch: got %q want %q", enc, wantEnc)
	}
}

func TestSanitizeFilename_UnsafeCharsReplaced(t *testing.T) {
	// quotes, backslash, CR, LF
	name := "weird\"name\\line\r\nend.txt"
	safe, enc := SanitizeFilename(name, "download")
	wantSafe := "weird_name_line__end.txt"
	if safe != wantSafe {
		t.Fatalf("safe mismatch: got %q want %q", safe, wantSafe)
	}
	if enc != url.PathEscape(wantSafe) {
		t.Fatalf("enc mismatch: got %q want %q", enc, url.PathEscape(wantSafe))
	}
}

func TestSanitizeFilename_ControlCharsStripped(t *testing.T) {
	// include SOH(0x01), STX(0x02), DEL(0x7F)
	name := "pre" + string(rune(0x01)) + "mid" + string(rune(0x02)) + "del" + string(rune(0x7F)) + "post"
	safe, enc := SanitizeFilename(name, "download")
	wantSafe := "premiddelpost"
	if safe != wantSafe {
		t.Fatalf("safe mismatch: got %q want %q", safe, wantSafe)
	}
	if enc != url.PathEscape(wantSafe) {
		t.Fatalf("enc mismatch: got %q want %q", enc, url.PathEscape(wantSafe))
	}
}

func TestSanitizeFilename_EmptyFallsBack(t *testing.T) {
	def := "download"
	safe, enc := SanitizeFilename("", def)
	if safe != def {
		t.Fatalf("safe mismatch: got %q want %q", safe, def)
	}
	if enc != url.PathEscape(def) {
		t.Fatalf("enc mismatch: got %q want %q", enc, url.PathEscape(def))
	}
}

func TestSanitizeFilename_UTF8Preserved(t *testing.T) {
	name := "招待状 2025 résumé.pdf"
	safe, enc := SanitizeFilename(name, "download")
	if safe != name {
		t.Fatalf("safe mismatch: got %q want %q", safe, name)
	}
	if enc != url.PathEscape(name) {
		t.Fatalf("enc mismatch: got %q want %q", enc, url.PathEscape(name))
	}
}

func TestSanitizeFilename_AllRemovedFallsBack(t *testing.T) {
	def := "fallback"
	onlyControls := string([]rune{0x01, 0x02, 0x03, 0x7F})
	safe, enc := SanitizeFilename(onlyControls, def)
	if safe != def {
		t.Fatalf("safe mismatch: got %q want %q", safe, def)
	}
	if enc != url.PathEscape(def) {
		t.Fatalf("enc mismatch: got %q want %q", enc, url.PathEscape(def))
	}
}
