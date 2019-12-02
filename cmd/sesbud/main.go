/*
 * Copyright 2018-2019 Kopano and its licensors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *	http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
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
	c.SetClientApp("kcc-go-sesbud", kcc.Version)
	sessions := make([]*kcc.Session, 0)

	for i := 1; i <= count; i++ {
		fmt.Printf("Starting session: %04d\n", i)

		session, err := kcc.NewSession(ctx, c, username, password)
		if err != nil {
			return err
		}
		sessions = append(sessions, session)
	}

	signalCh := make(chan os.Signal, 1)
	// Wait for exit or error.
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-signalCh:
		// breaks
	}

	fmt.Println("Exiting, logoff sessions")
	for _, session := range sessions {
		err := session.Destroy(ctx, true)
		if err != nil {
			fmt.Printf("Failed to destroy session: %v\n", err)
		}
	}

	return nil
}
