package main

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/mikefarah/yq/v4/pkg/yqlib"
	"github.com/spf13/cobra"
	"gopkg.in/op/go-logging.v1"
)

func init() {
	logging.SetLevel(logging.INFO, "yq-lib")
}

var (
	funcMap = template.FuncMap{
		"includeOAPISchema": oapiInclude(".components.schemas"),
		"includeOAPIPaths":  oapiInclude(".paths"),
		"includeYQ":         includeWithYQSelector,
		"includeVerbatim": func(filepath string) (string, error) {
			return includeWithYQSelector("", filepath)
		},
	}

	// Regular expression to match $ref with file paths
	// Matches patterns like "$ref": "somefile.json#/components/schemas/User"
	// or "$ref": "../folder/file.yaml#/components/schemas/User"
	// or $ref: "../folder/file.yaml#/components/schemas/User"
	// The first capture group named ref captures everything until a file path,
	// the second the file path and the third named fragment captures the component schema fragment.
	refPathReplaceRegex               = regexp.MustCompile(`(?P<ref>"?\$ref"?\s*:\s*")(.*?)(?P<fragment>#/.*".*)`)
	refPathReplaceRegexExpandTemplate = "$ref$fragment"

	rootCmd = &cobra.Command{
		Use:   "gotempl",
		Short: "gotempl is a CLI tool to help with OpenAPI",
		Long:  `gotempl is a CLI tool to help with OpenAPI. It offers the ability to merge referenced files into one.`,
	}

	mergeCmd = &cobra.Command{
		Use:   "template [template] [out]",
		Short: "template merges any referenced files into one",
		Long:  `template opens the passed filepath as a go template and executes that template, outputting into out. It offers the custom "include(YQ|Verbatim|OAPISchema|OAPIPaths)" functions taking either an absolute or relative path (to the template) and templates that file's content into the template.`,
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			yqlib.InitExpressionParser()
			tpl := template.New(args[0]).Funcs(funcMap).Funcs(sprig.TxtFuncMap())

			tpl, err := tpl.ParseFiles(args[0])
			if err != nil {
				log.Fatalf("failed to parse template file %q: %s", args[1], err)
			}
			tpl = tpl.Templates()[0]

			out, err := os.Create(args[1])
			if err != nil {
				log.Fatalf("failed to open output file %q: %s", args[1], err)
			}
			defer out.Close()

			if err := tpl.Execute(out, nil); err != nil {
				log.Fatalf("failed to execute template: %s", err)
			}
		},
	}
)

func main() {
	rootCmd.AddCommand(mergeCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

// includeWithYQSelector reads a file, applies the yq path selector on it and returns the filtered
// contents.
func includeWithYQSelector(selector, filepath string) (string, error) {
	contentBytes, err := os.ReadFile(filepath)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %w", filepath, err)
	}
	contents := string(contentBytes)

	if selector != "" {
		preferences := yqlib.NewDefaultYamlPreferences()
		encoder := yqlib.NewYamlEncoder(preferences)
		decoder := yqlib.NewYamlDecoder(preferences)

		contents, err = yqlib.NewStringEvaluator().Evaluate(selector, string(contents), encoder, decoder)
		if err != nil {
			return "", err
		}
	}

	return contents, nil
}

func oapiInclude(selector string) func(string) (string, error) {
	return func(filepath string) (string, error) {
		contents, err := includeWithYQSelector(selector, filepath)
		if err != nil {
			return "", err
		}

		return oapiAbsoluteRefs(contents), nil
	}
}

func oapiAbsoluteRefs(input string) string {
	return refPathReplaceRegex.ReplaceAllStringFunc(input, func(match string) string {
		index := refPathReplaceRegex.FindStringSubmatchIndex(match)
		if index != nil {
			return string(refPathReplaceRegex.ExpandString(nil, refPathReplaceRegexExpandTemplate, match, index))
		}
		return match
	})
}
