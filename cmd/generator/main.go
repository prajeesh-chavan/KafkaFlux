package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"go-kafka-simulator/internal/generator"
)

func main() {
	help := flag.Bool("help", false, "Print field type reference and YAML examples")
	validate := flag.String("validate", "", "Validate one or more profile YAML files (comma-separated)")
	initTmpl := flag.String("init", "", "Generate a starter profile from a built-in template (orders, customers, clickstream, etc.)")
	entity := flag.String("entity", "", "Entity name for batch mode (use with --field)")
	fields := fieldFlags{}
	flag.Var(&fields, "field", "Field definition for batch mode: name=id,type=uuid,publish_to=pool (repeatable)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "KafkaFlux Profile Generator\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "  generator                        Start interactive profile builder\n")
		fmt.Fprintf(os.Stderr, "  generator --help                 Show field type reference\n")
		fmt.Fprintf(os.Stderr, "  generator --validate <files>     Validate profile YAML(s)\n")
		fmt.Fprintf(os.Stderr, "  generator --init <template>      Generate a starter profile\n")
		fmt.Fprintf(os.Stderr, "  generator --entity X --field ...  Batch mode (non-interactive)\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	switch {
	case *help:
		generator.PrintHelp()
	case *validate != "":
		generator.ValidateProfiles(strings.Split(*validate, ","))
	case *initTmpl != "":
		generator.GenerateTemplate(*initTmpl)
	case *entity != "" || len(fields) > 0:
		if *entity == "" {
			fmt.Fprintln(os.Stderr, "Error: --entity is required when using --field")
			os.Exit(1)
		}
		generator.RunBatch(*entity, fields)
	default:
		generator.RunInteractive()
	}
}

type fieldFlags []string

func (f *fieldFlags) String() string {
	return strings.Join(*f, ", ")
}

func (f *fieldFlags) Set(v string) error {
	*f = append(*f, v)
	return nil
}
