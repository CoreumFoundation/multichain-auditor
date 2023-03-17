package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

// flags defined for cmd.
const (
	chainIDFlag      = "chain-id"
	coreumNodeFlag   = "coreum-node"
	coreumWalletFlag = "coreum-wallet"
)

func rootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Short: "Multichain watchdog",
	}

	cmd.AddCommand(coreumCmd())

	cmd.PersistentFlags().String(chainIDFlag, "coreum-mainnet-1", "specify the chain id (coreum-mainnet-1,coreum-testnet-1).")
	cmd.PersistentFlags().String(coreumNodeFlag, "", "specify the rpc address of the coreum node.")
	cmd.PersistentFlags().String(coreumWalletFlag, "", "specify multichain wallet on coreum.")
	return cmd
}

func coreumCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "coreum",
		Short: "Watch coreum wallet for transactions and write them to file",
	}

	cmd.AddCommand(
		coreumOutgoingCmd(),
		coreumIncomingCmd(),
	)

	return cmd
}

func coreumOutgoingCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "outgoing",
		Short: "Write outgoing transactions from coreum wallet to csv file",
		RunE: func(cmd *cobra.Command, args []string) error {
			config, ctx, log, err := setup(cmd)
			if err != nil {
				return err
			}
			clientCtx := createClientContext(config)
			log.Info("Fetching outgoing transactions from multichain coreum wallet")
			spentTxs, err := findTxsWithSingleBankSend(ctx, clientCtx, fmt.Sprintf("coin_spent.spender='%s'", config.coreumWallet))
			if err != nil {
				return err
			}

			err = writeCoreumTxsToCSV(spentTxs, config.denom, "datafiles/outgoing-on-coreum.csv")
			if err != nil {
				return err
			}

			return nil
		},
	}

	return cmd
}

func coreumIncomingCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "incoming",
		Short: "Write incoming transactions from coreum wallet to csv file",
		RunE: func(cmd *cobra.Command, args []string) error {
			config, ctx, log, err := setup(cmd)
			if err != nil {
				return err
			}

			clientCtx := createClientContext(config)

			log.Info("Fetching incoming transactions to multichain coreum wallet")
			receivedTxs, err := findTxsWithSingleBankSend(ctx, clientCtx, fmt.Sprintf("coin_received.receiver='%s'", config.coreumWallet))
			if err != nil {
				return err
			}

			err = writeCoreumTxsToCSV(receivedTxs, config.denom, "datafiles/incoming-on-coreum.csv")
			if err != nil {
				return err
			}
			return nil
		},
	}

	return cmd
}
