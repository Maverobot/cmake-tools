package cmakego

import (
	"github.com/pkg/errors"
	survey "gopkg.in/AlecAivazis/survey.v1"
)

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
