package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const startSize = 8

const template = `
## ClangTools
include(${CMAKE_CURRENT_LIST_DIR}/../cmake/ClangTools.cmake OPTIONAL
  RESULT_VARIABLE CLANG_TOOLS
)
if(CLANG_TOOLS)
  file(GLOB_RECURSE SOURCES
    ${GLOB_SOURCES}
  )
  file(GLOB_RECURSE HEADERS
    ${GLOB_HEADERS}
  )
  add_format_target(${PROJECT_NAME} FILES ${SOURCES} ${HEADERS})
  add_tidy_target(${PROJECT_NAME}
    FILES ${SOURCES}
    DEPENDS ${TARGETS}
  )
endif()
`

func main() {

	listFilePath := flag.String("path", "", "path to a CMakeLists.txt file")

	flag.Parse()

	if len(os.Args) < 2 || len(*listFilePath) == 0 {
		flag.Usage()
		return
	}

	newTemplate := getTemplate(*listFilePath)

	fmt.Print(newTemplate)
}

func getTemplate(listFilePath string) string {
	// Read files into strings
	contentList, err := ioutil.ReadFile(listFilePath)
	if err != nil {
		panic(err)
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

	// Find header folder paths
	headerAbsDirs := findHeaderDirs(rootDir)
	headerRelDirs := getRootChildren(rootDir, headerAbsDirs)

	// Print relative paths
	// fmt.Printf("src dirs: %v\n", srcRelDirs)
	// fmt.Printf("header dirs: %v\n", headerRelDirs)

	// Get source glob config
	srcConfArray := getGlobConfArray(srcRelDirs, "cpp")
	headerConfArray := getGlobConfArray(headerRelDirs, "h")

	output := replaceString(template, `\$\{GLOB_SOURCES\}`, strings.Join(srcConfArray, "\n    "))
	output = replaceString(output, `\$\{GLOB_HEADERS\}`, strings.Join(headerConfArray, "\n    "))

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
	targetMatch := string(` *project\((\w*)\)`)
	r := regexp.MustCompile(targetMatch)
	matches := r.FindAllStringSubmatch(text, -1)
	if len(matches) > 1 {
		name <- ""
	} else if len(matches) < 1 {
		name <- ""
	}
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
		panic(err)
	}

	dirs := make([]string, len(dirMap))
	i := 0
	for key := range dirMap {
		dirs[i] = key
		i++
	}
	return dirs
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
