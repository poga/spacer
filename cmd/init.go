package cmd

import (
	spacer "github.com/poga/spacer/pkg"
	"github.com/spf13/cobra"
)

var targetDir string
var shallow bool

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "init a new spacer project",
	Run: func(cmd *cobra.Command, args []string) {
		// TODO:
		// 1. write nginx configs to target directory
		if !shallow {
			spacer.RestoreAssets(targetDir, "nginx")
			return
		}
		// TODO: shallow - symlink to nginx configs in the repo
		// 2. add spacer.yml
		// 3. hello world function
	},
}

func init() {
	initCmd.Flags().StringVarP(&targetDir, "target", "t", "", "Target directory to init a new spacer project")
	initCmd.Flags().BoolVarP(&shallow, "shallow", "s", false, "Create symlink instead of copying project template to target directory. Useful for development")
	RootCmd.AddCommand(initCmd)
}
