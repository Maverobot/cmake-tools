package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
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
    ${CMAKE_CURRENT_SOURCE_DIR}/src/*.cpp)
  file(GLOB_RECURSE HEADERS
    ${CMAKE_CURRENT_SOURCE_DIR}/include/*.h
    ${CMAKE_CURRENT_SOURCE_DIR}/src/*.h
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

	//  ex, err := os.Executable()
	//  if err != nil {
	//  	panic(err)
	//  }

	//	exPath := filepath.Dir(ex)
	//	fmt.Println(exPath)

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

	// Puts project name into template
	newTemplate := replaceString(template, `\$\{PROJECT_NAME\}`, <-projectName)

	// Puts project name into template
	return replaceString(newTemplate, `\$\{TARGETS\}`, strings.Join(<-targets, " "))
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

// findLibraryNames finds the name of project in CMakeLists.txt
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
