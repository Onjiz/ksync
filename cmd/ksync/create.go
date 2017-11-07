package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/dustinkirkland/golang-petname"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/vapor-ware/ksync/pkg/cli"
	"github.com/vapor-ware/ksync/pkg/input"
	"github.com/vapor-ware/ksync/pkg/ksync"
)

type createCmd struct {
	cli.FinderCmd
}

func (cmd *createCmd) new() *cobra.Command {
	long := `
    create a new sync between a local and remote directory.`
	example := ``

	cmd.Init("ksync", &cobra.Command{
		Use:     "create [flags] [local path] [remote path]",
		Short:   "create a new sync between a local and remote directory.",
		Long:    long,
		Example: example,
		Aliases: []string{"c"},
		Args:    cobra.ExactArgs(2),
		Run:     cmd.run,
		// TODO: BashCompletionFunction
	})

	if err := cmd.DefaultFlags(); err != nil {
		log.Fatal(err)
	}

	flags := cmd.Cmd.Flags()

	rand.Seed(time.Now().UnixNano())
	flags.String(
		"name",
		petname.Generate(2, "-"),
		"Friendly name to describe this sync.")
	if err := cmd.BindFlag("name"); err != nil {
		log.Fatal(err)
	}

	flags.String(
		"user",
		fmt.Sprintf("%d:%d", os.Getuid(), os.Getgid()),
		"User to run the sync as locally. Defaults to current user.")
	if err := cmd.BindFlag("user"); err != nil {
		log.Fatal(err)
	}

	flags.Bool(
		"force",
		false,
		"Force creation, ignoring similarity.")
	if err := cmd.BindFlag("force"); err != nil {
		log.Fatal(err)
	}

	return cmd.Cmd
}

// TODO: check for existence of the watcher, warn if it isn't running.
func (cmd *createCmd) run(_ *cobra.Command, args []string) {
	syncPath := input.GetSyncPath(args)

	// Usage validation ------------------------------------
	if err := cmd.Validator(); err != nil {
		log.Fatal(err)
	}
	if err := syncPath.Validator(); err != nil {
		log.Fatal(err)
	}

	specMap, err := ksync.AllSpecs()
	if err != nil {
		log.Fatal(err)
	}

	newSpec := &ksync.Spec{
		Name: cmd.Viper.GetString("name"),
		User: cmd.Viper.GetString("user"),

		Namespace:   viper.GetString("namespace"),
		Context:     viper.GetString("context"),
		KubeCfgPath: ksync.KubeCfgPath,

		Container: cmd.Viper.GetString("container"),
		Pod:       cmd.Viper.GetString("pod"),
		Selector:  cmd.Viper.GetString("selector"),

		LocalPath:  syncPath.Local,
		RemotePath: syncPath.Remote,
	}

	if err := newSpec.IsValid(); err != nil {
		log.Fatal(err)
	}

	if err := specMap.Create(
		cmd.Viper.GetString("name"),
		newSpec,
		cmd.Viper.GetBool("force")); err != nil {

		log.Fatalf("Could not create, --force to ignore: %v", err)
	}
	if err := specMap.Save(); err != nil {
		log.Fatal(err)
	}
}