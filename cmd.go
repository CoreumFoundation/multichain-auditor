package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

// flags defined for cmd.
const (
	chainIDFlag        = "chain-id"
	coreumNodeFlag     = "coreum-node"
	coreumAccountFlag  = "coreum-account"
	outputDocumentFlag = "output-document"
)

func rootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Short: "Multichain Auditor",
	}

	cmd.AddCommand(coreumCmd())

	cmd.PersistentFlags().String(chainIDFlag, "coreum-mainnet-1", "specify the chain id (coreum-mainnet-1,coreum-testnet-1).")
	cmd.PersistentFlags().String(coreumNodeFlag, "", "specify the rpc address of the coreum node.")
	cmd.PersistentFlags().String(coreumAccountFlag, "", "specify multichain's account address on coreum.")
	return cmd
}

func coreumCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "coreum",
		Short: "Fetch transactions from multichain's coreum account and write them to file",
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
			log.Info("Fetching outgoing transactions from multichain's coreum wallet")
			spentTxs, err := findTxsWithSingleBankSend(ctx, clientCtx, fmt.Sprintf("coin_spent.spender='%s'", config.coreumAccount))
			if err != nil {
				return err
			}

			err = writeCoreumTxsToCSV(spentTxs, config.denom, config.outputDocument)
			if err != nil {
				return err
			}

			return nil
		},
	}

	cmd.PersistentFlags().String(outputDocumentFlag, "datafiles/outgoing-on-coreum.csv", "specify the output file")

	return cmd
}

func coreumIncomingCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "incoming",
		Short: "Write incoming transactions from multichain's coreum wallet to csv file",
		RunE: func(cmd *cobra.Command, args []string) error {
			config, ctx, log, err := setup(cmd)
			if err != nil {
				return err
			}

			clientCtx := createClientContext(config)

			log.Info("Fetching incoming transactions to multichain coreum wallet")
			receivedTxs, err := findTxsWithSingleBankSend(ctx, clientCtx, fmt.Sprintf("coin_received.receiver='%s'", config.coreumAccount))
			if err != nil {
				return err
			}

			err = writeCoreumTxsToCSV(receivedTxs, config.denom, config.outputDocument)
			if err != nil {
				return err
			}
			return nil
		},
	}

	cmd.PersistentFlags().String(outputDocumentFlag, "datafiles/incoming-on-coreum.csv", "specify the output file")

	return cmd
}
