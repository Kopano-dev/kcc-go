/*
 * Copyright 2018 Kopano and its licensors
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License, version 3,
 * as published by the Free Software Foundation.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"stash.kopano.io/kgol/kcc-go"
	"stash.kopano.io/kgol/kcc-go/cmd"
)

func main() {
	cmd.RootCmd.AddCommand(commandRun())

	if err := cmd.RootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func commandRun() *cobra.Command {
	runCmd := &cobra.Command{
		Use:   "run [...args]",
		Short: "Start sessions and periodically refresh them forever",
		Run: func(cmd *cobra.Command, args []string) {
			if err := run(cmd, args); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		},
	}
	runCmd.Flags().Int("count", 10, "Number of sessions")

	return runCmd
}

func run(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	count, _ := cmd.Flags().GetInt("count")

	username := "SYSTEM"
	password := ""
	if usernameOverride := os.Getenv("KOPANO_USERNAME"); usernameOverride != "" {
		username = usernameOverride
	}
	if passwordOverride := os.Getenv("KOPANO_PASSWORD"); passwordOverride != "" {
		password = passwordOverride
	}

	c := kcc.NewKCC(nil)
	sessions := make([]*kcc.Session, 0)

	for i := 1; i <= count; i++ {
		fmt.Printf("Starting session: %04d\n", i)

		session, err := kcc.NewSession(ctx, c, username, password)
		if err != nil {
			return err
		}
		sessions = append(sessions, session)
	}

	signalCh := make(chan os.Signal)
	// Wait for exit or error.
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-signalCh:
		// breaks
	}

	fmt.Println("Exiting, logoff sesions")
	for _, session := range sessions {
		err := session.Destroy(ctx, true)
		if err != nil {
			fmt.Printf("Failed to destroy session: %v\n", err)
		}
	}

	return nil
}
