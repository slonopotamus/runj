package main

import (
	"fmt"
	"os"
	"path/filepath"

	"go.sbk.wtf/runj/oci"

	"go.sbk.wtf/runj/state"

	"go.sbk.wtf/runj/jail"

	"github.com/spf13/cobra"
)

// createCommand implements the OCI "create" command
//
// create <container-id> <path-to-bundle>
//
// This operation MUST generate an error if it is not provided a path to the
// bundle and the container ID to associate with the container. If the ID
// provided is not unique across all containers within the scope of the runtime,
// or is not valid in any other way, the implementation MUST generate an error
// and a new container MUST NOT be created. This operation MUST create
// a new container.
//
// All of the properties configured in config.json except for process MUST be
// applied. process.args MUST NOT be applied until triggered by the start
// operation. The remaining process properties MAY be applied by this operation.
// If the runtime cannot apply a property as specified in the configuration, it
// MUST generate an error and a new container MUST NOT be created.
//
// The runtime MAY validate config.json against this spec, either generically or
// with respect to the local system capabilities, before creating the container
// (step 2). Runtime callers who are interested in pre-create validation can run
// bundle-validation tools before invoking the create operation.
//
// Any changes made to the config.json file after this operation will not have
// an effect on the container.
func createCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "create <container-id> <path-to-bundle>",
		Short: "Create a new container with given ID and bundle",
		Long:  "Create a new container with given ID and bundle.  IDs must be unique.",
		Args:  cobra.ExactArgs(2),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			bundle := args[1]
			bundleConfig := filepath.Join(bundle, oci.ConfigFileName)
			fInfo, err := os.Stat(bundleConfig)
			if err != nil {
				return err
			}
			if fInfo.Mode()&os.ModeType != 0 {
				return fmt.Errorf("%q should be a regular file", bundleConfig)
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			disableUsage(cmd)
			id := args[0]
			bundle := args[1]
			err = state.Create(id)
			if err != nil {
				return err
			}
			defer func() {
				if err != nil {
					state.Remove(id)
				}
			}()
			var confPath string
			confPath, err = jail.CreateConfig(id, filepath.Join(bundle, "root"))
			if err != nil {
				return err
			}
			err = oci.StoreConfig(id, bundle)
			if err != nil {
				return err
			}
			return jail.CreateJail(cmd.Context(), confPath)
		},
	}
}