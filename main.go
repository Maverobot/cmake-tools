package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	prompt "github.com/c-bata/go-prompt"
	"github.com/fatih/color"
	"github.com/pkg/errors"
	survey "gopkg.in/AlecAivazis/survey.v1"
)

const startSize = 8

const template = `
## ClangTools
include(${CMAKE_CURRENT_LIST_DIR}/cmake/ClangTools.cmake OPTIONAL
  RESULT_VARIABLE CLANG_TOOLS
)
if(CLANG_TOOLS)
  ${GLOB_SOURCE_SNIPPET}
  ${GLOB_HEADER_SNIPPET}
  add_format_target(${PROJECT_NAME} FILES ${SOURCES} ${HEADERS})
  add_tidy_target(${PROJECT_NAME}
    FILES ${SOURCES}
    DEPENDS ${TARGETS}
  )
endif()
`

const sourceSnippetTemplate = `file(GLOB_RECURSE SOURCES
    $${GLOB_SOURCES}
  )`
const headerSnippetTemplate = `file(GLOB_RECURSE HEADERS
    $${GLOB_HEADERS}
  )`

var configFileNames = [3]string{".clang-format", ".clang-tidy", "cmake"}

func main() {
	listFilePath := flag.String("path", "", "path to a CMakeLists.txt file")
	flag.Parse()
	if len(os.Args) < 2 || len(*listFilePath) == 0 {
		flag.Usage()
		return
	}
	newTemplate := getTemplate(*listFilePath)
	fmt.Print(newTemplate)

	// Ask for a directory to find cmake/ClangTools.cmake, .clang-format and .clang-tidy
	d := color.New(color.FgGreen, color.Bold)
	_, err := d.Printf("? ")
	d = color.New(color.FgWhite, color.Bold)
	_, err = d.Printf("Please type the path to cmake-tools: \n")
	if err != nil {
		panic(errors.Wrap(err, "color print failed"))
	}

	// Path autocompletion
	var options []string
	options = append(options, getExecPath())
	srcDir := prompt.Input("> ", createCompleter(options))

	if _, err := os.Stat(srcDir); !os.IsNotExist(err) {
		fmt.Println(srcDir + " is a path.")
	} else {
		fmt.Println(srcDir + " is not a path.")
		return
	}

	// Paths of the files to be copied
	srcPaths := make([]string, len(configFileNames))
	for i, n := range configFileNames {
		srcPaths[i] = filepath.Join(srcDir, n)
	}

	// Paths wherethe files to be copied to
	dstDir := filepath.Dir(*listFilePath)
	dstPaths := make([]string, len(configFileNames))
	for i, n := range configFileNames {
		dstPaths[i] = filepath.Join(dstDir, n)
	}

	for i, src := range srcPaths {
		t := getPathType(src)
		var err error
		switch t {
		case filePath:
			err = copyFile(src, dstPaths[i])
		case dirPath:
			err = copyDir(src, dstPaths[i])
		case noPath:
			panic(errors.Errorf("%s does not exist", src))
		}
		if err != nil {
			panic(errors.Wrap(err, "copy failed"))
		}
	}

}

func getTemplate(listFilePath string) string {
	// Read files into strings
	contentList, err := ioutil.ReadFile(listFilePath)
	if err != nil {
		panic(errors.Wrap(err, "read file failed"))
	}

	targets := make(chan []string)
	projectName := make(chan string)

	// Find library or executable names
	go findLibraryNames(string(contentList), targets)

	// Find project name
	go findProjectName(string(contentList), projectName)

	rootDir := filepath.Dir(listFilePath)

	// Find source folder paths
	srcAbsDirs := findSourceDirs(rootDir)
	srcRelDirs := getRootChildren(rootDir, srcAbsDirs)
	srcRelDirs = getUnique(srcRelDirs)
	srcRelDirs = userFilterOptions("source filter", "Verify your source folder(s): \n", srcRelDirs)

	// Find header folder paths
	headerAbsDirs := findHeaderDirs(rootDir)
	headerRelDirs := getRootChildren(rootDir, headerAbsDirs)
	headerRelDirs = getUnique(headerRelDirs)
	headerRelDirs = userFilterOptions("header filter", "Verify your header folder(s): \n", headerRelDirs)

	// Get source glob config
	srcConfArray := getGlobConfArray(srcRelDirs, "cpp")
	headerConfArray := getGlobConfArray(headerRelDirs, "h")

	// Get source and header snippets
	sourceSnippet := ""
	if len(srcConfArray) != 0 {
		sourceSnippet = replaceString(sourceSnippetTemplate,
			`\$\{GLOB_SOURCES\}`,
			strings.Join(srcConfArray, "\n    $$"))
	}
	headerSnippet := ""
	if len(headerConfArray) != 0 {
		headerSnippet = replaceString(headerSnippetTemplate,
			`\$\{GLOB_HEADERS\}`,
			strings.Join(headerConfArray, "\n    $$"))
	}

	output := replaceString(template, `\$\{GLOB_SOURCE_SNIPPET\}`, sourceSnippet)
	output = replaceString(output, `\$\{GLOB_HEADER_SNIPPET\}`, headerSnippet)

	// Puts project name into output
	output = replaceString(output, `\$\{PROJECT_NAME\}`, <-projectName)

	// Puts project name into output
	return replaceString(output, `\$\{TARGETS\}`, strings.Join(<-targets, " "))
}

// findLibraryNames finds the names of libraries and executables defined
// by add_library and add_executable
func findLibraryNames(text string, names chan<- []string) {
	libNames := make([]string, 0, startSize)
	targetMatch := string(` *add_(?:library|executable)\( *(\w*)`)
	r := regexp.MustCompile(targetMatch)
	matches := r.FindAllStringSubmatch(text, -1)
	for _, match := range matches {
		libNames = append(libNames, match[1])
	}
	names <- libNames
}

// findProjectName finds the name of project in CMakeLists.txt
func findProjectName(text string, name chan<- string) {
	targetMatch := string(`\s*project\(((?:\w|-)*)\s*.*\)`)
	r := regexp.MustCompile(targetMatch)
	matches := r.FindAllStringSubmatch(text, -1)
	if len(matches) != 1 {
		name <- ""
	}
	// TODO: throw error if no matches are found
	name <- matches[0][1]
}

func replaceString(src string, pattern string, repl string) string {
	r := regexp.MustCompile(pattern)
	return r.ReplaceAllString(src, repl)
}

// findSourceDirs finds the folder paths of cpp files in cmake project
func findSourceDirs(path string) []string {
	pattern := `\.cpp$`
	return findAllMatchDirs(path, pattern)
}

// findHeaderDirs finds the folder paths of h/hpp files in cmake project
func findHeaderDirs(path string) []string {
	pattern := `\.(?:h|hpp)$`
	return findAllMatchDirs(path, pattern)
}

func findAllMatchDirs(path string, pattern string) []string {
	dirMap := make(map[string]struct{})
	r := regexp.MustCompile(pattern)

	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Fatal(err)
		}
		if r.MatchString(path) {
			dir := filepath.Dir(path)
			if _, ok := dirMap[dir]; !ok {
				dirMap[dir] = struct{}{}
			} else {
			}
		}
		return nil
	})
	if err != nil {
		panic(errors.Wrap(err, "file walk failed"))
	}

	dirs := make([]string, len(dirMap))
	i := 0
	for key := range dirMap {
		dirs[i] = key
		i++
	}
	return dirs
}

func getUnique(src []string) []string {
	uMap := make(map[string]struct{})
	for _, v := range src {
		if _, ok := uMap[v]; !ok {
			uMap[v] = struct{}{}
		}
	}
	uArr := make([]string, len(uMap))
	i := 0
	for key := range uMap {
		uArr[i] = key
		i++
	}
	return uArr
}

func getRootChildren(rootDir string, absDirs []string) []string {
	relDirs := make([]string, len(absDirs))
	for i, absDir := range absDirs {
		tmp := strings.Replace(absDir, rootDir, "", 1)
		dirs := strings.Split(tmp, `/`)
		relDirs[i] = dirs[1]
	}
	return relDirs
}

func getGlobConfArray(folderNames []string, ext string) []string {
	template := `$${CMAKE_CURRENT_SOURCE_DIR}/%s/*.%s`
	conf := make([]string, len(folderNames))
	for i, v := range folderNames {
		conf[i] = fmt.Sprintf(template, v, ext)
	}
	return conf
}

func createMultiSelectQuestion(name string, message string, options []string) []*survey.Question {
	return []*survey.Question{
		{
			Name: name,
			Prompt: &survey.MultiSelect{
				Message: message,
				Options: options,
				Default: options,
			},
		},
	}
}

func userFilterOptions(name string, info string, src []string) []string {
	answers := []string{}
	question := createMultiSelectQuestion(name, info, src)
	err := survey.Ask(question[:], &answers)
	if err != nil {
		panic(errors.Wrap(err, "survey ask failed"))
	}
	return answers
}

func getSuggestionsPath(path string) []string {
	// remove the letters after and inclusive the last "/"
	var re = regexp.MustCompile(`\w*$`)
	path = re.ReplaceAllString(path, ``)

	info, err := os.Stat(path)
	if err != nil {
		return []string{}
	}

	var children []string
	if info.IsDir() {
		files, err := ioutil.ReadDir(path)
		if err != nil {
			return []string{}
		}
		// Strip "/" at the end of path
		if strings.HasSuffix(path, "/") {
			path = path[0 : len(path)-1]
		}

		for _, file := range files {
			children = append(children, path+"/"+file.Name())
		}
	}

	return children
}

func createCompleter(textList []string) prompt.Completer {

	completer := func(d prompt.Document) []prompt.Suggest {
		var s []prompt.Suggest
		for _, value := range textList {
			s = append(s, prompt.Suggest{Text: value, Description: ""})
		}

		children := getSuggestionsPath(d.GetWordBeforeCursor())

		for _, value := range children {
			s = append(s, prompt.Suggest{Text: value, Description: ""})
		}

		return prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), true)
	}

	return completer
}

func getExecPath() string {
	ex, err := os.Executable()
	if err != nil {
		panic(errors.Wrap(err, "get exec path failed"))
	}
	return filepath.Dir(ex)
}

func getAbsPaths(root string, children []string) []string {
	res := make([]string, len(children))
	for i, child := range children {
		res[i] = filepath.Join(root, child)
	}
	return res
}

// File copies a single file from src to dst
func copyFile(src, dst string) error {
	var err error
	var srcfd *os.File
	var dstfd *os.File
	var srcinfo os.FileInfo

	if srcfd, err = os.Open(src); err != nil {
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

// Dir copies a whole directory recursively
func copyDir(src string, dst string) error {
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
			if err = copyDir(srcfp, dstfp); err != nil {
				fmt.Println(err)
			}
		} else {
			if err = copyFile(srcfp, dstfp); err != nil {
				fmt.Println(err)
			}
		}
	}
	return nil
}

func pathExists(path string) bool {
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		return true
	}
	return false
}

type pathType int

const (
	filePath pathType = iota
	dirPath
	noPath
)

func getPathType(path string) pathType {
	fi, err := os.Stat(path)
	if err != nil {
		return noPath
	}
	switch mode := fi.Mode(); {
	case mode.IsDir():
		return dirPath
	case mode.IsRegular():
		return filePath
	}
	return noPath
}
