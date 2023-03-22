package main

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

// flags defined for cmd.
const (
	chainIDFlag      = "chain-id"
	startDateFlag    = "start-date"
	coreumNodeFlag   = "coreum-node"
	coreumWalletFlag = "coreum-wallet"

	xrplRPCAPIURLFlag        = "xrpl-rpc-api-url"
	xrplHistoricalAPIURLFlag = "xrpl-historical-api-url"
	xrplAccountFlag          = "xrpl-account"
	xrplCurrencyFlag         = "xrpl-currency"
	xrplIssuerFlag           = "xrpl-issuer"
	xrplChainIndexFlag       = "xrpl-chain-index"
)

func rootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Short: "Multichain watchdog",
	}

	cmd.AddCommand(coreumCmd())
	cmd.AddCommand(xrplCmd())

	cmd.PersistentFlags().String(chainIDFlag, "coreum-mainnet-1", "chain id (coreum-mainnet-1,coreum-testnet-1)")
	cmd.PersistentFlags().String(coreumNodeFlag, "", "rpc address of the coreum node")
	cmd.PersistentFlags().String(coreumWalletFlag, "", "multichain wallet on coreum")
	cmd.PersistentFlags().String(startDateFlag, "", fmt.Sprintf("date to fetch from, format: %s", time.DateOnly))

	cmd.PersistentFlags().String(xrplRPCAPIURLFlag, "", "xrpl RPC address")
	cmd.PersistentFlags().String(xrplHistoricalAPIURLFlag, "", "xrpl historical API address")
	cmd.PersistentFlags().String(xrplAccountFlag, "", "xrpl account")
	cmd.PersistentFlags().String(xrplCurrencyFlag, "", "xrpl hex currency")
	cmd.PersistentFlags().String(xrplIssuerFlag, "", "xrpl issuer")
	cmd.PersistentFlags().String(xrplChainIndexFlag, "", "xrpl chain index")

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

func xrplCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "xrpl",
		Short: "Watch xrpl account for transactions and write them to file",
	}

	cmd.AddCommand(
		xrplIncomingCmd(),
	)

	return cmd
}

func xrplIncomingCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "incoming",
		Short: "Write incoming transactions from xrpl address to csv file",
		RunE: func(cmd *cobra.Command, args []string) error {
			config, ctx, log, err := setup(cmd)
			if err != nil {
				return err
			}

			log.Info(fmt.Sprintf("Fetching incoming transactions for %s xrpl account", config.xrplAccount))
			txs, err := GetXRPLPaymentTransactions(ctx, config.xrplRPCAPIURL, config.xrplHistoricalAPIURL, config.xrplAccount, config.xrplCurrency, config.xrplIssuer, config.startDate)
			if err != nil {
				return err
			}

			filteredTxs := FilterXRPLBridgeTransactionsAndConvertToExportItem(config.xrplChainIndex, txs)
			err = writeTxsToCSV(filteredTxs, "datafiles/incoming-on-xrpl.csv")
			if err != nil {
				return err
			}

			return nil
		},
	}

	return cmd
}
