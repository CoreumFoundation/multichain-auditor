package main

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
)

// flags defined for cmd.
const (
	beforeDateTimeFlag          = "before-date-time"
	afterDateTimeFlag           = "after-date-time"
	coreumNodeFlag              = "coreum-node"
	coreumAccountFlag           = "coreum-account"
	coreumFoundationAccountFlag = "coreum-foundation-account"
	xrplFetchPoolSizeFlag       = "xrpl-fetch-pool-size"
	xrplRPCAPIURLFlag           = "xrpl-rpc-api-url"
	xrplHistoricalAPIURLFlag    = "xrpl-historical-api-url"
	xrplScanAPIURLFlag          = "xrpl-scan-api-url"
	xrplAccountFlag             = "xrpl-account"
	xrplCurrencyFlag            = "xrpl-currency"
	xrplIssuerFlag              = "xrpl-issuer"
	bridgeChainIndexFlag        = "bridge-chain-index"
	outputDocumentFlag          = "output-document"
	includeAllFlag              = "include-all"
	multichainRescanAPIURLFlag  = "multichain-rescan-api-url"
)

const (
	defaultCoreumRPC               = "https://full-node.mainnet-1.coreum.dev:26657"
	defaultCoreumAccount           = "core1ssh2d2ft6hzrgn9z6k7mmsamy2hfpxl9y8re5x"
	defaultCoreumFoundationAccount = "core13xmyzhvl02xpz0pu8v9mqalsvpyy7wvs9q5f90"

	defaultXrplFetchPullSize    = 10
	defaultXrplRPCAPIURL        = "https://s2.ripple.com:51234/"
	defaultXrplHistoricalAPIURL = "https://data.ripple.com"
	defaultXrplScanAPIURL       = "https://api.xrpscan.com"

	defaultXrplAccount            = "rcoreNywaoz2ZCQ8Lg2EbSLnGuRBmun6D"
	defaultXrplCurrency           = "434F524500000000000000000000000000000000"
	defaultXrplIssuer             = "rcoreNywaoz2ZCQ8Lg2EbSLnGuRBmun6D"
	defaultBridgeChainIndex       = "1007961752909"
	defaultMultichainRescanAPIURL = "https://scanapi.multichain.org"
)

var (
	defaultBeforeDateTime = time.Now().UTC()
	defaultAfterDateTime  = time.Date(2023, time.Month(3), 1, 0, 0, 0, 0, time.UTC)
)

func rootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Short: "Multichain Auditor",
	}

	cmd.AddCommand(coreumCmd())
	cmd.AddCommand(xrplCmd())
	cmd.AddCommand(discrepancyCmd())
	cmd.AddCommand(summaryCmd())

	cmd.PersistentFlags().String(coreumNodeFlag, defaultCoreumRPC, "coreum rpc address")
	cmd.PersistentFlags().String(coreumAccountFlag, defaultCoreumAccount, "multichain account on coreum")
	cmd.PersistentFlags().String(coreumFoundationAccountFlag, defaultCoreumFoundationAccount, "foundation account on coreum")
	cmd.PersistentFlags().String(beforeDateTimeFlag, defaultBeforeDateTime.Format(time.DateTime), fmt.Sprintf("UTC date and time to fetch from, format: %s", time.DateTime))
	cmd.PersistentFlags().String(afterDateTimeFlag, defaultAfterDateTime.Format(time.DateTime), fmt.Sprintf("UTC date and time to fetch to, format: %s", time.DateTime))
	cmd.PersistentFlags().String(xrplRPCAPIURLFlag, defaultXrplRPCAPIURL, "xrpl RPC address")
	cmd.PersistentFlags().Int(xrplFetchPoolSizeFlag, defaultXrplFetchPullSize, "xrpl fetch pool size")
	cmd.PersistentFlags().String(xrplHistoricalAPIURLFlag, defaultXrplHistoricalAPIURL, "xrpl historical API address")
	cmd.PersistentFlags().String(xrplScanAPIURLFlag, defaultXrplScanAPIURL, "xrpl scan API address")
	cmd.PersistentFlags().String(xrplAccountFlag, defaultXrplAccount, "xrpl account")
	cmd.PersistentFlags().String(xrplCurrencyFlag, defaultXrplCurrency, "xrpl hex currency")
	cmd.PersistentFlags().String(xrplIssuerFlag, defaultXrplIssuer, "xrpl issuer")
	cmd.PersistentFlags().String(bridgeChainIndexFlag, defaultBridgeChainIndex, "xrpl chain index")

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
				config.BeforeDateTime,
				config.AfterDateTime,
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
				fmt.Sprintf("coin_received.receiver='%s'", config.CoreumAccount),
				config.Denom,
				config.BeforeDateTime,
				config.AfterDateTime,
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
		Short: "Fetch xrpl account transactions",
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
				config.XrplFetchPoolSize,
				config.XrplRPCAPIURL,
				config.XrplHistoricalAPIURL,
				config.XrplAccount,
				config.XrplCurrency,
				config.XrplIssuer,
				config.BridgeChainIndex,
				config.BeforeDateTime,
				config.AfterDateTime,
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
		discrepancyExportCmd(),
		discrepancyRescanCmd(),
	)

	return cmd
}

func discrepancyExportCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export",
		Short: "Write all transactions xrpl and coreum discrepancies to csv file",
		RunE: func(cmd *cobra.Command, args []string) error {
			config, ctx, log, err := Setup(cmd)
			if err != nil {
				return err
			}
			log.Info("Exporting discrepancies.")
			discrepancies, err := findTxDiscrepancies(ctx, config)
			if err != nil {
				return err
			}

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

func discrepancyRescanCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rescan",
		Short: "Rescans all transactions orphan xrpl txs",
		RunE: func(cmd *cobra.Command, args []string) error {
			config, ctx, log, err := Setup(cmd)
			if err != nil {
				return err
			}

			log.Info("Rescanning orphan discrepancies.")
			discrepancies, err := findTxDiscrepancies(ctx, config)
			if err != nil {
				return err
			}

			txHashes := make([]string, 0)
			for _, discrepancy := range discrepancies {
				if discrepancy.Discrepancy == DiscrepancyOrphanXrplTx {
					txHashes = append(txHashes, discrepancy.XrplTx.Hash)
				}
			}

			return RescanMultichainTxs(ctx, config.MultichainRescanAPIURL, txHashes)
		},
	}

	cmd.PersistentFlags().String(multichainRescanAPIURLFlag, defaultMultichainRescanAPIURL, "multichain rescan API url")

	return cmd
}

func summaryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "summary",
		Short: "Get summary data.",
	}

	cmd.AddCommand(
		summaryPrintCmd(),
	)

	return cmd
}

func summaryPrintCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "print",
		Short: "Get and print summary report.",
		RunE: func(cmd *cobra.Command, args []string) error {
			config, ctx, log, err := Setup(cmd)
			if err != nil {
				return err
			}
			log.Info("Fetching data for the report.")

			xrplSupply, err := GetXrplCurrencySupply(ctx, config.XrplScanAPIURL, config.XrplIssuer, config.XrplCurrency)
			if err != nil {
				return err
			}
			clientCtx := createClientContext(config)

			coreumBalance, err := GetCoreumAccountBalance(ctx, clientCtx, config.CoreumAccount, config.Denom)
			if err != nil {
				return err
			}

			config.IncludeAll = true
			discrepancies, err := findTxDiscrepancies(ctx, config)
			if err != nil {
				return err
			}

			coreumIncomingAuditTxs, err := GetCoreumAuditTransactions(
				ctx,
				clientCtx,
				fmt.Sprintf("coin_received.receiver='%s'", config.CoreumAccount),
				config.Denom,
				config.BeforeDateTime,
				config.AfterDateTime,
			)
			if err != nil {
				return err
			}
			foundationCoreumIncomingAuditTxs := make([]AuditTx, 0)
			for _, auditTx := range coreumIncomingAuditTxs {
				if auditTx.FromAddress == config.CoreumFoundationAccount {
					foundationCoreumIncomingAuditTxs = append(foundationCoreumIncomingAuditTxs, auditTx)
				}
			}

			summary := BuildSummary(discrepancies, foundationCoreumIncomingAuditTxs, coreumBalance, xrplSupply)

			log.Info("Summary report:")
			log.Info(fmt.Sprintf("\n%s", summary.String()))

			return nil
		},
	}

	return cmd
}

func findTxDiscrepancies(ctx context.Context, config Config) ([]TxDiscrepancy, error) {
	log := logger.Get(ctx)
	log.Info(fmt.Sprintf("Fetching incoming transactions for %s xrpl account", config.XrplAccount))
	xrplAuditTxs, err := GetXRPLAuditTransactions(
		ctx,
		config.XrplFetchPoolSize,
		config.XrplRPCAPIURL,
		config.XrplHistoricalAPIURL,
		config.XrplAccount,
		config.XrplCurrency,
		config.XrplIssuer,
		config.BridgeChainIndex,
		defaultBeforeDateTime, // for the discrepancies we export full history and filter later
		defaultAfterDateTime,
	)
	if err != nil {
		return nil, err
	}

	clientCtx := createClientContext(config)
	log.Info("Fetching outgoing transactions from multichain coreum wallet")
	coreumAuditTxs, err := GetCoreumAuditTransactions(
		ctx,
		clientCtx,
		fmt.Sprintf("coin_spent.spender='%s'", config.CoreumAccount),
		config.Denom,
		defaultBeforeDateTime, // for the discrepancies we export full history and filter later
		defaultAfterDateTime,
	)
	if err != nil {
		return nil, err
	}

	discrepancies := FindAuditTxDiscrepancies(
		xrplAuditTxs,
		coreumAuditTxs,
		config.FeeConfigs,
		config.IncludeAll,
		config.BeforeDateTime,
		config.AfterDateTime,
	)
	log.Info(fmt.Sprintf("Found %d discrepancies", len(discrepancies)))

	return discrepancies, nil
}
