package cmd

import (
	"bytes"
	"os/exec"
	"strings"
	"unicode"

	"github.com/spf13/cobra"
	"github.com/twpayne/go-shell"

	"github.com/twpayne/chezmoi/chezmoi2/internal/chezmoi"
)

type diffCmdConfig struct {
	include   *chezmoi.IncludeSet
	recursive bool
	NoPager   bool
	Pager     string
}

func (c *Config) newDiffCmd() *cobra.Command {
	diffCmd := &cobra.Command{
		Use:     "diff [target]...",
		Short:   "Print the diff between the target state and the destination state",
		Long:    mustLongHelp("diff"),
		Example: example("diff"),
		RunE:    c.runDiffCmd,
		Annotations: map[string]string{
			persistentStateMode: persistentStateModeReadMockWrite,
		},
	}

	flags := diffCmd.Flags()
	flags.VarP(c.Diff.include, "include", "i", "include entry types")
	flags.BoolVar(&c.Diff.NoPager, "no-pager", c.Diff.NoPager, "disable pager")
	flags.BoolVarP(&c.Diff.recursive, "recursive", "r", c.Diff.recursive, "recursive")

	return diffCmd
}

func (c *Config) runDiffCmd(cmd *cobra.Command, args []string) error {
	sb := strings.Builder{}
	dryRunSystem := chezmoi.NewDryRunSystem(c.destSystem)
	gitDiffSystem := chezmoi.NewGitDiffSystem(dryRunSystem, &sb, c.destDirAbsPath, c.color)
	if err := c.applyArgs(gitDiffSystem, c.destDirAbsPath, args, applyArgsOptions{
		include:   c.Diff.include,
		recursive: c.Diff.recursive,
		umask:     c.Umask.FileMode(),
	}); err != nil {
		return err
	}

	if c.Diff.NoPager || c.Diff.Pager == "" {
		return c.writeOutputString(sb.String())
	}

	// If the pager command contains any spaces, assume that it is a full
	// shell command to be executed via the user's shell. Otherwise, execute
	// it directly.
	var pagerCmd *exec.Cmd
	if strings.IndexFunc(c.Diff.Pager, unicode.IsSpace) != -1 {
		shell, _ := shell.CurrentUserShell()
		pagerCmd = exec.Command(shell, "-c", c.Diff.Pager)
	} else {
		//nolint:gosec
		pagerCmd = exec.Command(c.Diff.Pager)
	}
	pagerCmd.Stdin = bytes.NewReader([]byte(sb.String()))
	pagerCmd.Stdout = c.stdout
	pagerCmd.Stderr = c.stderr
	return pagerCmd.Run()
}
