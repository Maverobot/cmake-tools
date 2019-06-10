package cmakego

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/pkg/errors"
)

// CopyConfigFiles copies the config files and directories to the target cmake project
func CopyConfigFiles(srcDir string, listFilePath string) {
	// Paths of the files to be copied
	srcPaths := make([]string, len(configFileNames))
	for i, n := range configFileNames {
		srcPaths[i] = filepath.Join(srcDir, n)
	}

	// Paths wherethe files to be copied to
	dstDir := filepath.Dir(listFilePath)
	dstPaths := make([]string, len(configFileNames))
	for i, n := range configFileNames {
		dstPaths[i] = filepath.Join(dstDir, n)
	}

	for i, src := range srcPaths {
		t := getPathType(src)
		var err error
		switch t {
		case filePath:
			err = CopyFile(src, dstPaths[i])
		case DirPath:
			err = CopyDir(src, dstPaths[i])
		case NoPath:
			panic(errors.Errorf("%s does not exist", src))
		}
		if err != nil {
			panic(errors.Wrap(err, "copy failed"))
		}
	}

}

func AddConfigToCMakeLists() {

}

// GetExecPath returns the path where the binary was executed
func GetExecPath() string {
	ex, err := os.Executable()
	if err != nil {
		panic(errors.Wrap(err, "get exec path failed"))
	}
	return filepath.Dir(ex)
}

// GetAbsPaths returns the absolute paths of relative paths given root path
func GetAbsPaths(root string, children []string) []string {
	res := make([]string, len(children))
	for i, child := range children {
		res[i] = filepath.Join(root, child)
	}
	return res
}

// CopyFile copies a single file from src to dst
func CopyFile(src, dst string) error {
	var err error
	var srcfd *os.File
	var dstfd *os.File
	var srcinfo os.FileInfo

	if srcfd, err = os.Open(filepath.Clean(src)); err != nil {
		return err
	}
	defer srcfd.Close()

	if dstfd, err = os.Create(dst); err != nil {
		return err
	}
	defer dstfd.Close()

	if _, err = io.Copy(dstfd, srcfd); err != nil {
		return err
	}
	if srcinfo, err = os.Stat(src); err != nil {
		return err
	}
	return os.Chmod(dst, srcinfo.Mode())
}

// CopyDir copies a whole directory recursively
func CopyDir(src string, dst string) error {
	var err error
	var fds []os.FileInfo
	var srcinfo os.FileInfo

	if srcinfo, err = os.Stat(src); err != nil {
		return err
	}

	if err = os.MkdirAll(dst, srcinfo.Mode()); err != nil {
		return err
	}

	if fds, err = ioutil.ReadDir(src); err != nil {
		return err
	}
	for _, fd := range fds {
		srcfp := path.Join(src, fd.Name())
		dstfp := path.Join(dst, fd.Name())

		if fd.IsDir() {
			if err = CopyDir(srcfp, dstfp); err != nil {
				fmt.Println(err)
			}
		} else {
			if err = CopyFile(srcfp, dstfp); err != nil {
				fmt.Println(err)
			}
		}
	}
	return nil
}

// PathExists checks if a given path exists
func PathExists(path string) bool {
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		return true
	}
	return false
}

// PathType define types of of path using int
type PathType int

const (
	// filePath indicates the path is for a file
	filePath PathType = iota
	// DirPath indicates the path is for a directory
	DirPath
	// NoPath indicates the path does not exist
	NoPath
)

// getPathType returns type of path given a string of the path
func getPathType(path string) PathType {
	fi, err := os.Stat(path)
	if err != nil {
		return NoPath
	}
	switch mode := fi.Mode(); {
	case mode.IsDir():
		return DirPath
	case mode.IsRegular():
		return filePath
	}
	return NoPath
}
