package main

import (
	"context"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/CoreumFoundation/coreum/pkg/config"
	"github.com/CoreumFoundation/coreum/pkg/config/constant"
	"github.com/CoreumFoundation/faucet/pkg/logger"
)

const (
	defaultCoreumTestnetRPC    = "https://full-node.testnet-1.coreum.dev:26657"
	defaultMainnetTestnetRPC   = "https://full-node.mainnet-1.coreum.dev:26657"
	defaultCoreumWalletTestnet = "testcore1pykqce6sh6szm8mkzmsjweyucshahe5gjeykxr"
	defaultCoreumWalletMainnet = "core1ssh2d2ft6hzrgn9z6k7mmsamy2hfpxl9y8re5x"
)

type Config struct {
	chainID          string
	denom            string
	coreumWallet     string
	coreumRPCAddress string
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
			coreumRPCAddress = defaultMainnetTestnetRPC
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

	return Config{
		chainID:          chainID,
		denom:            network.Denom(),
		coreumRPCAddress: coreumRPCAddress,
		coreumWallet:     coreumWallet,
	}, nil
}