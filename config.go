package main

import (
	"context"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/CoreumFoundation/coreum/pkg/config"
	"github.com/CoreumFoundation/coreum/pkg/config/constant"
	"github.com/CoreumFoundation/faucet/pkg/logger"
)

const (
	defaultCoreumTestnetRPC    = "https://full-node.testnet-1.coreum.dev:26657"
	defaultCoreumMainnetRPC    = "https://full-node.mainnet-1.coreum.dev:26657"
	defaultCoreumWalletTestnet = "testcore1pykqce6sh6szm8mkzmsjweyucshahe5gjeykxr"
	defaultCoreumWalletMainnet = "core1ssh2d2ft6hzrgn9z6k7mmsamy2hfpxl9y8re5x"

	defaultXrplTestnetRPCAPIURL        = "https://xrplcluster.com/"
	defaultXrplTestnetHistoricalAPIURL = "https://data.ripple.com"
	defaultXrplTestnetAccount          = "rcoreNywaoz2ZCQ8Lg2EbSLnGuRBmun6D"
	defaultXrplTestnetCurrency         = "434F524500000000000000000000000000000000"
	defaultXrplTestnetIssuer           = "rcoreNywaoz2ZCQ8Lg2EbSLnGuRBmun6D"
	defaultXrplTestnetChainIndex       = "1007961752909"

	// FIXME(dhil) replace to testnet settings
	defaultXrplMainnetRPCAPIURL        = "https://xrplcluster.com/"
	defaultXrplMainnetHistoricalAPIURL = "https://data.ripple.com"
	defaultXrplMainnetAccount          = "rcoreNywaoz2ZCQ8Lg2EbSLnGuRBmun6D"
	defaultXrplMainnetCurrency         = "434F524500000000000000000000000000000000"
	defaultXrplMainnetIssuer           = "rcoreNywaoz2ZCQ8Lg2EbSLnGuRBmun6D"
	defaultXrplMainnetChainIndex       = "1007961752909"
)

type Config struct {
	chainID   string
	startDate time.Time

	denom            string
	coreumWallet     string
	coreumRPCAddress string

	xrplRPCAPIURL        string
	xrplHistoricalAPIURL string
	xrplAccount          string
	xrplCurrency         string
	xrplIssuer           string
	xrplChainIndex       string
}

func setup(cmd *cobra.Command) (Config, context.Context, *zap.Logger, error) {
	loggerConfig, _ := logger.ConfigureWithCLI(logger.ToolDefaultConfig)
	log := logger.New(loggerConfig)
	ctx := logger.WithLogger(context.Background(), log)

	config, err := getConfig(cmd)
	if err != nil {
		return Config{}, nil, nil, err
	}

	return config, ctx, log, nil
}

func getConfig(cmd *cobra.Command) (Config, error) {
	chainID, err := cmd.Flags().GetString(chainIDFlag)
	if err != nil {
		return Config{}, err
	}

	startDate := time.Date(2023, time.Month(1), 1, 1, 0, 0, 0, time.UTC)
	startDateString, err := cmd.Flags().GetString(startDateFlag)
	if err != nil {
		return Config{}, err
	}
	if startDateString != "" {
		startDate, err = time.Parse(time.DateOnly, startDateString)
		if err != nil {
			return Config{}, err
		}
	}

	network, err := config.NetworkByChainID(constant.ChainID(chainID))
	if err != nil {
		return Config{}, err
	}

	network.SetSDKConfig()

	coreumRPCAddress, err := cmd.Flags().GetString(coreumNodeFlag)
	if err != nil {
		return Config{}, err
	}

	if coreumRPCAddress == "" {
		switch constant.ChainID(chainID) {
		case constant.ChainIDTest:
			coreumRPCAddress = defaultCoreumTestnetRPC
		case constant.ChainIDMain:
			coreumRPCAddress = defaultCoreumMainnetRPC
		}
	}

	coreumWallet, err := cmd.Flags().GetString(coreumWalletFlag)
	if err != nil {
		return Config{}, err
	}

	if coreumWallet == "" {
		switch constant.ChainID(chainID) {
		case constant.ChainIDTest:
			coreumWallet = defaultCoreumWalletTestnet
		case constant.ChainIDMain:
			coreumWallet = defaultCoreumWalletMainnet
		}
	}

	xrplRPCAPIURL, err := cmd.Flags().GetString(xrplRPCAPIURLFlag)
	if err != nil {
		return Config{}, err
	}

	if xrplRPCAPIURL == "" {
		switch constant.ChainID(chainID) {
		case constant.ChainIDTest:
			xrplRPCAPIURL = defaultXrplTestnetRPCAPIURL
		case constant.ChainIDMain:
			xrplRPCAPIURL = defaultXrplMainnetRPCAPIURL
		}
	}

	xrplHistoricalAPIURL, err := cmd.Flags().GetString(xrplHistoricalAPIURLFlag)
	if err != nil {
		return Config{}, err
	}
	if xrplHistoricalAPIURL == "" {
		switch constant.ChainID(chainID) {
		case constant.ChainIDTest:
			xrplHistoricalAPIURL = defaultXrplTestnetHistoricalAPIURL
		case constant.ChainIDMain:
			xrplHistoricalAPIURL = defaultXrplMainnetHistoricalAPIURL
		}
	}

	xrplAccount, err := cmd.Flags().GetString(xrplAccountFlag)
	if err != nil {
		return Config{}, err
	}
	if xrplAccount == "" {
		switch constant.ChainID(chainID) {
		case constant.ChainIDTest:
			xrplAccount = defaultXrplTestnetAccount
		case constant.ChainIDMain:
			xrplAccount = defaultXrplMainnetAccount
		}
	}

	xrplCurrency, err := cmd.Flags().GetString(xrplCurrencyFlag)
	if err != nil {
		return Config{}, err
	}
	if xrplCurrency == "" {
		switch constant.ChainID(chainID) {
		case constant.ChainIDTest:
			xrplCurrency = defaultXrplTestnetCurrency
		case constant.ChainIDMain:
			xrplCurrency = defaultXrplMainnetCurrency
		}
	}

	xrplIssuer, err := cmd.Flags().GetString(xrplIssuerFlag)
	if err != nil {
		return Config{}, err
	}
	if xrplIssuer == "" {
		switch constant.ChainID(chainID) {
		case constant.ChainIDTest:
			xrplIssuer = defaultXrplTestnetIssuer
		case constant.ChainIDMain:
			xrplIssuer = defaultXrplMainnetIssuer
		}
	}

	xrplChainIndex, err := cmd.Flags().GetString(xrplChainIndexFlag)
	if err != nil {
		return Config{}, err
	}
	if xrplChainIndex == "" {
		switch constant.ChainID(chainID) {
		case constant.ChainIDTest:
			xrplChainIndex = defaultXrplTestnetChainIndex
		case constant.ChainIDMain:
			xrplChainIndex = defaultXrplMainnetChainIndex
		}
	}

	return Config{
		chainID:          chainID,
		startDate:        startDate,
		denom:            network.Denom(),
		coreumWallet:     coreumWallet,
		coreumRPCAddress: coreumRPCAddress,

		xrplRPCAPIURL:        xrplRPCAPIURL,
		xrplHistoricalAPIURL: xrplHistoricalAPIURL,
		xrplAccount:          xrplAccount,
		xrplCurrency:         xrplCurrency,
		xrplIssuer:           xrplIssuer,
		xrplChainIndex:       xrplChainIndex,
	}, nil
}
