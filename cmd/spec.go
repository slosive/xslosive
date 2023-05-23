package cmd

import (
	"github.com/tfadeyi/sloth-simple-comments/internal/parser/strategy/golang"
	"github.com/tfadeyi/sloth-simple-comments/internal/parser/strategy/wasm"
	"os"

	"github.com/spf13/cobra"
	specoptions "github.com/tfadeyi/sloth-simple-comments/cmd/options/spec"
	"github.com/tfadeyi/sloth-simple-comments/internal/generate"
	"github.com/tfadeyi/sloth-simple-comments/internal/logging"
	"github.com/tfadeyi/sloth-simple-comments/internal/parser"
	"github.com/tfadeyi/sloth-simple-comments/internal/parser/lang"
	"github.com/tfadeyi/sloth-simple-comments/internal/parser/options"
)

func specGenerateCmd() *cobra.Command {
	opts := specoptions.New()
	var outputDir string
	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate the Sloth definition specification from source code comments.",
		Long: `The generate command parses files in the target directory for comment using the @sloth tags,
i.e: 
	// @sloth name chat-gpt-availability
	// @sloth objective 95.0
	// @sloth.sli error_query sum(rate(tenant_failed_login_operations_total{client="chat-gpt"}[{{.window}}])) OR on() vector(0)
	// @sloth.sli total_query sum(rate(tenant_login_operations_total{client="chat-gpt"}[{{.window}}]))
	// @sloth description 95% of logins to the chat-gpt app should be successful annotations.

These are then used to generate Sloth definition specifications. 
i.e:
	version: prometheus/v1
	service: "chatgpt"
	slos:
		- name: chat-gpt-availability
		  description: 95% of logins to the chat-gpt app should be successful.
		  objective: 95
		  sli:
			raw:
				error_ratio_query: ""
			events:
				error_query: sum(rate(tenant_failed_login_operations_total{client="chat-gpt"}[{{.window}}])) OR on() vector(0)
				total_query: sum(rate(tenant_login_operations_total{client="chat-gpt"}[{{.window}}]))
		  alerting:
			name: ""
`,
		SilenceErrors: true,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			// if an argument is passed to the command
			// arg 1: should be the output directory where to store the output
			output, err := os.Getwd()
			if err != nil {
				return err
			}
			if len(args) == 1 {
				output = args[0]
			}
			outputDir = output
			return opts.Complete()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			logger := logging.LoggerFromContext(cmd.Context())
			var languageParser options.Option
			switch opts.SrcLanguage {
			case lang.Wasm:
				logger.Info("The wasm parser has not been fully implemented and shouldn't be used! It will have unexpected behaviours.")
				languageParser = wasm.Parser()
			default:
				languageParser = golang.Parser()
			}

			parser, err := parser.New(
				languageParser,
				options.Logger(&logger),
				options.Include(opts.IncludedDirs...))
			if err != nil {
				return err
			}
			service, err := parser.Parse(cmd.Context())
			if err != nil {
				return err
			}

			return generate.WriteSpecification(service, opts.StdOutput, outputDir, opts.Formats...)
		},
	}
	opts = opts.Prepare(cmd)
	return cmd
}

func init() {
	rootCmd.AddCommand(specGenerateCmd())
}
