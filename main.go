package main

import (
	"flag"
	"fmt"
	"os"

	prompt "github.com/c-bata/go-prompt"
	"github.com/fatih/color"
	"github.com/mbndr/figlet4go"
	"github.com/pkg/errors"

	cmakego "github.com/maverobot/cmake-tools/src"
)

func main() {
	listFilePath := flag.String("path", "", "path to a CMakeLists.txt file")
	useCustomClangConfig := flag.Bool("v", false, "Use custom clang config.")
	flag.Parse()
	if len(os.Args) < 2 || len(*listFilePath) == 0 {
		flag.Usage()
		return
	}

	// figlet fun
	figletFromString("cmake-tools")

	// get the config text for CMakeLists.txt
	newTemplate := cmakego.GetTemplate(*listFilePath)

	// determine where to get the config files
	var srcDir string
	if *useCustomClangConfig {
		// Ask for a directory to find cmake/ClangTools.cmake, .clang-format and .clang-tidy
		d := color.New(color.FgGreen, color.Bold)
		_, err := d.Printf("? ")
		if err != nil {
			panic(errors.Wrap(err, "color print failed"))
		}
		d = color.New(color.FgWhite, color.Bold)
		_, err = d.Printf("Please type the path to cmake-tools: \n")
		if err != nil {
			panic(errors.Wrap(err, "color print failed"))
		}

		// Path autocompletion
		var options []string
		options = append(options, cmakego.GetExecPath())
		srcDir = prompt.Input("> ", cmakego.CreatePathCompleter(options))

		if _, err := os.Stat(srcDir); os.IsNotExist(err) {
			fmt.Println(srcDir + " is not a path.")
			return
		}
	} else {
		srcDir = cmakego.GetExecPath()
	}

	// copy necessary files to the target cmake project
	cmakego.CopyConfigFiles(srcDir, *listFilePath)

	// Add template to the given CMakeLists
	f, err := os.OpenFile(*listFilePath, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	if _, err = f.WriteString(newTemplate); err != nil {
		panic(err)
	}

	fmt.Printf("\nThe cmake tools have been successfully configured." +
		"\n\tNow use \"-DCMAKE_EXPORT_COMPILE_COMMANDS=ON\" flag during cmake, " +
		"and try with \"make format\" and \"make tidy\".\n")
}

func figletFromString(src string) {
	ascii := figlet4go.NewAsciiRender()
	options := figlet4go.NewRenderOptions()

	options.FontColor = []figlet4go.Color{
		// Colors can be given by default ansi color codes...
		figlet4go.ColorGreen,
	}

	// The underscore would be an error
	renderStr, err := ascii.RenderOpts(src, options)
	if err != nil {
		panic(errors.Wrap(err, "figlet rendering failed"))
	}
	fmt.Print(renderStr)
}
