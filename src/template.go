package cmakego

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pkg/errors"
)

// GetTemplate returns a string to be added into CMakeLists.txt to configure clang tools
func GetTemplate(listFilePath string) string {
	// Read files into strings
	contentList, err := ioutil.ReadFile(filepath.Clean(listFilePath))
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
	if len(headerRelDirs) != 0 {
		headerRelDirs = userFilterOptions("header filter", "Verify your header folder(s): \n", headerRelDirs)
	}

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

	output := replaceString(clangConfigTemplate, `\$\{GLOB_SOURCE_SNIPPET\}`, sourceSnippet)
	output = replaceString(output, `\$\{GLOB_HEADER_SNIPPET\}`, headerSnippet)

	// Puts project name into output
	output = replaceString(output, `\$\{PROJECT_NAME\}`, <-projectName)

	// Puts project name into output
	return replaceString(output, `\$\{TARGETS\}`, strings.Join(<-targets, " "))
}

// findLibraryNames finds the names of libraries and executables defined
// by add_library and add_executable
func findLibraryNames(text string, names chan<- []string) {
	const startSize = 8
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
	targetMatch := string(`\s*project\s*\(((?:\w|-)*)\s*.*\)`)
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
