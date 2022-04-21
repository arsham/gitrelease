package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/arsham/gitrelease/commit"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	tag        string
	printMode  bool
	remote     string
	version    = "development"
	currentSha = "N/A"

	rootCmd = &cobra.Command{
		Use:   "gitrelease",
		Short: "Release commit information of a tag to github",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 && args[0] == "version" {
				fmt.Printf("gitrelease version %s (%s)\n", version, currentSha)
				return nil
			}

			ctx, cancel := signal.NotifyContext(cmd.Context(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
			defer cancel()
			token := os.Getenv("GITHUB_TOKEN")
			if token == "" {
				return errors.New("please export GITHUB_TOKEN")
			}
			g := &commit.Git{
				Remote: remote,
			}

			user, repo, err := g.RepoInfo(ctx)
			if err != nil {
				return errors.Wrap(err, "can't get repo name")
			}

			tag1, err := g.PreviousTag(ctx, tag)
			if err != nil {
				return errors.Wrap(err, "getting previous tag")
			}

			logs, err := g.Commits(ctx, tag1, tag)
			if err != nil {
				return err
			}
			desc := commit.ParseGroups(logs)
			if tag == "@" {
				tag, err = g.LatestTag(ctx)
				if err != nil {
					return err
				}
			}

			if printMode {
				_, err := fmt.Println(desc)
				return err
			}

			return g.Release(ctx, token, user, repo, tag, desc)
		},
	}
)

func main() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	cobra.OnInitialize(viper.AutomaticEnv)
	rootCmd.PersistentFlags().StringVarP(&tag, "tag", "t", "@", "tag to produce the logs for. Leave empty for current tag.")
	rootCmd.PersistentFlags().BoolVarP(&printMode, "print", "p", false, "only print, do not release!")
	rootCmd.PersistentFlags().StringVarP(&remote, "remote", "r", "origin", "use a different remote")

	rootCmd.SetUsageTemplate(`Usage:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}

Available Commands:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Available Commands:
  help        Help about any command
  version     Print binary version information

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`)
}
