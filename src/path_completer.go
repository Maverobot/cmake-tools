package cmakego

import (
	"io/ioutil"
	"os"
	"regexp"
	"strings"

	prompt "github.com/c-bata/go-prompt"
)

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

// CreatePathCompleter provides a completer for prompt to autocomplete system paths
func CreatePathCompleter(textList []string) prompt.Completer {

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
