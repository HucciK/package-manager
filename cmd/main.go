package main

import (
	"fmt"
	"github.com/HucciK/package-manager/config"
	"github.com/HucciK/package-manager/internal"
	"github.com/HucciK/package-manager/internal/core"
	"github.com/spf13/cobra"
	"os"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   `pm`,
		Short: `pm is package manager for test task`,
		Long:  `pm create or update specified packages`,
	}

	rootCmd.AddCommand(CreateCmd())
	rootCmd.AddCommand(UpdateCmd())

	if err := rootCmd.Execute(); err != nil {
		return
	}
}

func CreateCmd() *cobra.Command {

	command := &cobra.Command{
		Use:   `create`,
		Short: `create - creates zip by paths specified in json`,
		Long:  `create - creates zip by paths specified in json`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunCreate(args)
		},
	}

	return command
}

func RunCreate(args []string) error {
	cfg, err := config.NewConfig()
	if err != nil {
		return fmt.Errorf("can't initialize config: %w", err)
	}

	packetPath := args[0]
	packet, err := core.NewPacket(packetPath)
	if err != nil {
		return fmt.Errorf("can't create packet: %w", err)
	}

	zipFile, err := os.OpenFile(packet.ZipName(), os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return fmt.Errorf("can't create zip file for packet: %w", err)
	}
	defer zipFile.Close()

	client, err := internal.NewSSHClient(cfg)
	if err != nil {
		return fmt.Errorf("can't initialize ssh client: %w", err)
	}

	zip := internal.NewZipHandler(zipFile)

	manager := internal.NewManager(zip, client)
	if err := manager.CreatePacket(packet); err != nil {
		return fmt.Errorf("can't create packet: %w", err)
	}

	return nil
}

func UpdateCmd() *cobra.Command {

	command := &cobra.Command{
		Use:   `update`,
		Short: `update - updates packets`,
		Long:  `update - updates packet by name to specified ver`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunUpdate(args)
		},
	}

	return command
}

func RunUpdate(args []string) error {
	cfg, err := config.NewConfig()
	if err != nil {
		return fmt.Errorf("can't initialize config: %w", err)
	}

	packagesPath := args[0]
	packages, err := core.NewPackages(packagesPath)
	if err != nil {
		return fmt.Errorf("can't parse packages info: %w", err)
	}

	client, err := internal.NewSSHClient(cfg)
	if err != nil {
		return fmt.Errorf("can't initialize ssh client: %w", err)
	}

	zip := internal.NewZipHandler(nil)

	manager := internal.NewManager(zip, client)
	if err := manager.UpdatePackets(packages); err != nil {
		return fmt.Errorf("can't update packets: %w", err)
	}

	return nil
}
