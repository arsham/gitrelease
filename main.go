package main

import (
	"os"

	"github.com/arsham/gitrelease/commit"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	tag     string
	rootCmd = &cobra.Command{
		Use:   "gitrelease",
		Short: "Print the release information of a tag",
		RunE: func(cmd *cobra.Command, _ []string) error {
			token := os.Getenv("GITHUB_TOKEN")
			if token == "" {
				return errors.New("please export GITHUB_TOKEN")
			}
			g := &commit.Git{}

			user, repo, err := g.RepoInfo(cmd.Context())
			if err != nil {
				return errors.Wrap(err, "can't get repo name")
			}

			tag1, err := g.PreviousTag(cmd.Context(), tag)
			if err != nil {
				return errors.Wrap(err, "getting previous tag")
			}

			logs, err := g.Commits(cmd.Context(), tag1, tag)
			if err != nil {
				return err
			}
			desc := commit.ParseGroups(logs)
			if tag == "@" {
				tag, err = g.LatestTag(cmd.Context())
				if err != nil {
					return err
				}
			}

			return g.Release(token, user, repo, tag, desc)
		},
	}
)

func main() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	cobra.OnInitialize(viper.AutomaticEnv)
	rootCmd.PersistentFlags().StringVarP(&tag, "tag", "t", "@", "tag to produce the logs for. Leave empty for current tag.")
}
