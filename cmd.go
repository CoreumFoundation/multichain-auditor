package main

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

// flags defined for cmd.
const (
	chainIDFlag              = "chain-id"
	fromDateTimeFlag         = "from-date-time"
	toDateTimeFlag           = "to-date-time"
	coreumNodeFlag           = "coreum-node"
	coreumAccountFlag        = "coreum-account"
	xrplRPCAPIURLFlag        = "xrpl-rpc-api-url"
	xrplHistoricalAPIURLFlag = "xrpl-historical-api-url"
	xrplAccountFlag          = "xrpl-account"
	xrplCurrencyFlag         = "xrpl-currency"
	xrplIssuerFlag           = "xrpl-issuer"
	bridgeChainIndexFlag     = "bridge-chain-index"
	outputDocumentFlag       = "output-document"
	includeAllFlag           = "include-all"
)

var (
	defaultFromDateTime = time.Now()
	defaultToDateTime   = time.Date(2023, time.Month(3), 1, 0, 0, 0, 0, time.UTC)
)

func rootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Short: "Multichain Auditor",
	}

	cmd.AddCommand(coreumCmd())
	cmd.AddCommand(xrplCmd())
	cmd.AddCommand(discrepancyCmd())

	cmd.PersistentFlags().String(chainIDFlag, "coreum-mainnet-1", "chain id (coreum-mainnet-1,coreum-testnet-1)")
	cmd.PersistentFlags().String(coreumNodeFlag, "", "coreum rpc address")
	cmd.PersistentFlags().String(coreumAccountFlag, "", "multichain account on coreum")
	cmd.PersistentFlags().String(fromDateTimeFlag, defaultFromDateTime.Format(time.DateTime), fmt.Sprintf("UTC date and time to fetch from, format: %s", time.DateTime))
	cmd.PersistentFlags().String(toDateTimeFlag, defaultToDateTime.Format(time.DateTime), fmt.Sprintf("UTC date and time to fetch to, format: %s", time.DateTime))
	cmd.PersistentFlags().String(xrplRPCAPIURLFlag, "", "xrpl RPC address")
	cmd.PersistentFlags().String(xrplHistoricalAPIURLFlag, "", "xrpl historical API address")
	cmd.PersistentFlags().String(xrplAccountFlag, "", "xrpl account")
	cmd.PersistentFlags().String(xrplCurrencyFlag, "", "xrpl hex currency")
	cmd.PersistentFlags().String(xrplIssuerFlag, "", "xrpl issuer")
	cmd.PersistentFlags().String(bridgeChainIndexFlag, "", "xrpl chain index")

	return cmd
}

func coreumCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "coreum",
		Short: "Fetch transactions from multichain's coreum account",
	}

	cmd.AddCommand(
		coreumOutgoingCmd(),
		coreumIncomingCmd(),
	)

	return cmd
}

func coreumOutgoingCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export-outgoing",
		Short: "Write outgoing transactions from coreum wallet to csv file",
		RunE: func(cmd *cobra.Command, args []string) error {
			config, ctx, log, err := Setup(cmd)
			if err != nil {
				return err
			}
			clientCtx := createClientContext(config)
			log.Info("Fetching outgoing transactions from multichain's coreum wallet")
			coreumAuditTxs, err := GetCoreumAuditTransactions(
				ctx,
				clientCtx,
				fmt.Sprintf("coin_spent.spender='%s'", config.CoreumAccount),
				config.Denom,
				config.FromDateTime,
				config.ToDateTime,
			)
			if err != nil {
				return err
			}

			err = WriteAuditTxsToCSV(coreumAuditTxs, config.OutputDocument)
			if err != nil {
				return err
			}

			return nil
		},
	}

	cmd.PersistentFlags().String(outputDocumentFlag, "datafiles/outgoing-on-coreum.csv", "output file")

	return cmd
}

func coreumIncomingCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export-incoming",
		Short: "Write incoming transactions from multichain's coreum wallet to csv file",
		RunE: func(cmd *cobra.Command, args []string) error {
			config, ctx, log, err := Setup(cmd)
			if err != nil {
				return err
			}

			clientCtx := createClientContext(config)

			log.Info("Fetching incoming transactions to multichain coreum wallet")
			coreumAuditTxs, err := GetCoreumAuditTransactions(
				ctx,
				clientCtx,
				fmt.Sprintf("coin_spent.spender='%s'", config.CoreumAccount),
				config.Denom,
				config.FromDateTime,
				config.ToDateTime,
			)
			if err != nil {
				return err
			}

			err = WriteAuditTxsToCSV(coreumAuditTxs, config.OutputDocument)
			if err != nil {
				return err
			}

			return nil
		},
	}

	cmd.PersistentFlags().String(outputDocumentFlag, "datafiles/incoming-on-coreum.csv", "output file")

	return cmd
}

func xrplCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "xrpl",
		Short: "Fetch xrpl account for transactions",
	}

	cmd.AddCommand(
		xrplIncomingCmd(),
	)

	return cmd
}

func xrplIncomingCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export-incoming",
		Short: "Write incoming transactions from xrpl address to csv file",
		RunE: func(cmd *cobra.Command, args []string) error {
			config, ctx, log, err := Setup(cmd)
			if err != nil {
				return err
			}

			log.Info(fmt.Sprintf("Fetching incoming transactions for %s xrpl account", config.XrplAccount))
			xrplAuditTxs, err := GetXRPLAuditTransactions(
				ctx,
				config.XrplRPCAPIURL,
				config.XrplHistoricalAPIURL,
				config.XrplAccount,
				config.XrplCurrency,
				config.XrplIssuer,
				config.BridgeChainIndex,
				config.FromDateTime,
				config.ToDateTime,
			)
			if err != nil {
				return err
			}

			err = WriteAuditTxsToCSV(xrplAuditTxs, config.OutputDocument)
			if err != nil {
				return err
			}

			return nil
		},
	}

	cmd.PersistentFlags().String(outputDocumentFlag, "datafiles/incoming-on-xrpl.csv", "output file")

	return cmd
}

func discrepancyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "discrepancy",
		Short: "Fetch xrpl and coreum discrepancies",
	}

	cmd.AddCommand(
		discrepancyAllCmd(),
	)

	return cmd
}

func discrepancyAllCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export",
		Short: "Write all transactions xrpl and coreum discrepancies to csv file",
		RunE: func(cmd *cobra.Command, args []string) error {
			config, ctx, log, err := Setup(cmd)
			if err != nil {
				return err
			}

			log.Info(fmt.Sprintf("Fetching incoming transactions for %s xrpl account", config.XrplAccount))
			xrplAuditTxs, err := GetXRPLAuditTransactions(
				ctx,
				config.XrplRPCAPIURL,
				config.XrplHistoricalAPIURL,
				config.XrplAccount,
				config.XrplCurrency,
				config.XrplIssuer,
				config.BridgeChainIndex,
				config.FromDateTime,
				config.ToDateTime,
			)
			if err != nil {
				return err
			}

			clientCtx := createClientContext(config)
			log.Info("Fetching outgoing transactions from multichain coreum wallet")
			coreumAuditTxs, err := GetCoreumAuditTransactions(
				ctx,
				clientCtx,
				fmt.Sprintf("coin_spent.spender='%s'", config.CoreumAccount),
				config.Denom,
				config.FromDateTime,
				config.ToDateTime,
			)
			if err != nil {
				return err
			}

			discrepancies := FindAuditTxDiscrepancies(xrplAuditTxs, coreumAuditTxs, config.FeeConfigs, config.IncludeAll)
			log.Info(fmt.Sprintf("Found %d discrepancies", len(discrepancies)))
			err = WriteTxsDiscrepancyToCSV(discrepancies, config.OutputDocument)
			if err != nil {
				return err
			}

			return nil
		},
	}

	cmd.PersistentFlags().String(outputDocumentFlag, "datafiles/discrepancies.csv", "output file")
	cmd.PersistentFlags().Bool(includeAllFlag, false, "add all tx to output file even if no discrepancies are found")

	return cmd
}
