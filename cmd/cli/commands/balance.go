package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewBalanceCmd creates the BUNKER balance command.
func NewBalanceCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "balance",
		Short: "BUNKER token balance",
		Long:  "Display your BUNKER token balance. Works via HTTP API without a daemon.",
		RunE:  runBalance,
	}
}

func runBalance(cmd *cobra.Command, args []string) error {
	c, err := GetClient()
	if err != nil {
		return err
	}
	defer c.Close()

	bal, err := c.Balance()
	if err != nil {
		return fmt.Errorf("failed to get balance: %w", err)
	}

	fmt.Println(StatusBox("BUNKER Balance", [][2]string{
		{"Address", FormatAddress(bal.Address)},
		{"Available", bal.Available + " BUNKER"},
		{"Reserved", bal.Reserved + " BUNKER"},
		{"Staked", bal.Staked + " BUNKER"},
		{"Total", bal.Total + " BUNKER"},
	}))

	return nil
}
