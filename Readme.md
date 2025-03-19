# gotempl

gotempl is a small cli tool allowing you to apply go templating to arbitrary files.

Specifically, it provides 4 templating functions that open an arbitrary file and include its content in the templated position.
Both absolute and relative filepaths are accepted, relative paths are interpreted as relative to the current working directory from which
gotempl is called.

```
# includeOAPISchema includes the content in file.foo under the .components.schemas yq selector and removes all file paths from replace directives.
{{ includeOAPISchema "../path/to/file.foo" }}

# includeOAPIPaths includes the content in file.foo under the .paths yq selector and removes all file paths from replace directives.
{{ includeOAPIPaths "../path/to/file.foo" }}

# includeYQ includes the content in file.foo under the passed yq selector. If the selector is empty everything is included.
{{ includeYQ "../path/to/file.foo" ".yqselector" }}

# includeVerbatim includes the file's content verbatim.
{{ includeVerbatim "../path/to/file.foo" }}
```

In addition to these 4 functions all of [sprig's](https://github.com/Masterminds/sprig) text functions are available.

The tool's `template` subcommand accepts two positional arguments, first the path to the template file, second the path to the output file.
If the output file does not exist it will be created, if it exists it will be overwritten.

An example of the usage is (run from inside the test folder):
`gotempl template templ.bar out.txt`
