package main

import (
	"context"
	"math/big"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/CoreumFoundation/coreum/pkg/config"
	"github.com/CoreumFoundation/coreum/pkg/config/constant"
	"github.com/CoreumFoundation/faucet/pkg/logger"
)

type Config struct {
	BeforeDateTime          time.Time
	AfterDateTime           time.Time
	Denom                   string
	CoreumAccount           string
	CoreumFoundationAccount string
	CoreumRPCURL            string
	XrplRPCAPIURL           string
	XrplScanAPIURL          string
	XrplHistoricalAPIURL    string
	XrplAccount             string
	XrplCurrency            string
	XrplIssuer              string
	BridgeChainIndex        string
	OutputDocument          string
	FeeConfigs              []FeeConfig
	IncludeAll              bool
	MultichainRescanAPIURL  string
}

// FeeConfig the settings used for the calculation of the final amount which includes fee.
type FeeConfig struct {
	StartTime time.Time
	FeeRatio  *big.Int // to get percent you need to div it by 1000, we need it to make the calculation without floats
	MinFee    *big.Int
	MaxFee    *big.Int
	MinAmount *big.Int
	MaxAmount *big.Int
}

func Setup(cmd *cobra.Command) (Config, context.Context, *zap.Logger, error) {
	loggerConfig, _ := logger.ConfigureWithCLI(logger.ToolDefaultConfig)
	log := logger.New(loggerConfig)
	ctx := logger.WithLogger(context.Background(), log)

	сfg, err := getConfig(cmd)
	if err != nil {
		return Config{}, nil, nil, err
	}

	return сfg, ctx, log, nil
}

func getConfig(cmd *cobra.Command) (Config, error) {
	beforeDateTimeString, err := cmd.Flags().GetString(beforeDateTimeFlag)
	if err != nil {
		return Config{}, err
	}
	beforeDateTime, err := time.Parse(time.DateTime, beforeDateTimeString)
	if err != nil {
		return Config{}, errors.Errorf("error parsing %s the expected format is %s", beforeDateTimeFlag, time.DateTime)
	}

	afterDateTimeString, err := cmd.Flags().GetString(afterDateTimeFlag)
	if err != nil {
		return Config{}, err
	}
	afterDateTime, err := time.Parse(time.DateTime, afterDateTimeString)
	if err != nil {
		return Config{}, errors.Errorf("error parsing %s the expected format is %s", afterDateTimeFlag, time.DateTime)
	}

	network, err := config.NetworkByChainID(constant.ChainIDMain)
	if err != nil {
		return Config{}, err
	}

	network.SetSDKConfig()

	coreumRPCAddress, err := cmd.Flags().GetString(coreumNodeFlag)
	if err != nil {
		return Config{}, err
	}

	coreumAccount, err := cmd.Flags().GetString(coreumAccountFlag)
	if err != nil {
		return Config{}, err
	}

	coreumFoundationAccount, err := cmd.Flags().GetString(coreumFoundationAccountFlag)
	if err != nil {
		return Config{}, err
	}

	xrplRPCAPIURL, err := cmd.Flags().GetString(xrplRPCAPIURLFlag)
	if err != nil {
		return Config{}, err
	}

	xrplHistoricalAPIURL, err := cmd.Flags().GetString(xrplHistoricalAPIURLFlag)
	if err != nil {
		return Config{}, err
	}

	xrplScanAPIURL, err := cmd.Flags().GetString(xrplScanAPIURLFlag)
	if err != nil {
		return Config{}, err
	}

	xrplAccount, err := cmd.Flags().GetString(xrplAccountFlag)
	if err != nil {
		return Config{}, err
	}

	xrplCurrency, err := cmd.Flags().GetString(xrplCurrencyFlag)
	if err != nil {
		return Config{}, err
	}

	xrplIssuer, err := cmd.Flags().GetString(xrplIssuerFlag)
	if err != nil {
		return Config{}, err
	}

	bridgeChainIndex, err := cmd.Flags().GetString(bridgeChainIndexFlag)
	if err != nil {
		return Config{}, err
	}

	outputDocument := ""
	if cmd.Flags().Lookup(outputDocumentFlag) != nil {
		outputDocument, err = cmd.Flags().GetString(outputDocumentFlag)
		if err != nil {
			return Config{}, err
		}
	}

	includeAll := false
	if cmd.Flags().Lookup(includeAllFlag) != nil {
		includeAll, err = cmd.Flags().GetBool(includeAllFlag)
		if err != nil {
			return Config{}, err
		}
	}

	multichainRescanAPIURL := ""
	if cmd.Flags().Lookup(multichainRescanAPIURLFlag) != nil {
		multichainRescanAPIURL, err = cmd.Flags().GetString(multichainRescanAPIURLFlag)
		if err != nil {
			return Config{}, err
		}
	}

	// the feeConfigs are fixed, and can be modified in the code only
	// we use the list of the configs since the fees have been modified during the bridge life, and
	// each time period uses different fee configs.
	feeConfigs := []FeeConfig{
		{
			StartTime: time.Date(2023, time.Month(3), 24, 17, 0, 0, 0, time.UTC),
			FeeRatio:  big.NewInt(1),                // 0.1%
			MinFee:    big.NewInt(2_400000),         // 2.4 CORE
			MaxFee:    big.NewInt(477_000000),       // 477 CORE
			MinAmount: big.NewInt(4_800000),         // 4.8 CORE
			MaxAmount: big.NewInt(2_400_000_000000), // 2.400.000 CORE
		},
		{
			StartTime: time.Date(2023, time.Month(3), 17, 13, 0, 0, 0, time.UTC),
			FeeRatio:  big.NewInt(1),          // 0.1%
			MinFee:    big.NewInt(7000),       // 0.007 CORE
			MaxFee:    big.NewInt(50000),      // 0.05 CORE
			MinAmount: big.NewInt(8000),       // 0.008 CORE
			MaxAmount: big.NewInt(100_000000), // 100 CORE
		},
		{
			StartTime: time.Date(2023, time.Month(1), 1, 0, 0, 0, 0, time.UTC),
			FeeRatio:  big.NewInt(1),                // 0.1%
			MinFee:    big.NewInt(2_400000),         // 2.4 CORE
			MaxFee:    big.NewInt(477_000000),       // 477 CORE
			MinAmount: big.NewInt(4_800000),         // 4.8 CORE
			MaxAmount: big.NewInt(2_400_000_000000), // 2.400.000 CORE
		},
	}

	return Config{
		BeforeDateTime:          beforeDateTime.UTC(),
		AfterDateTime:           afterDateTime.UTC(),
		Denom:                   network.Denom(),
		CoreumAccount:           coreumAccount,
		CoreumFoundationAccount: coreumFoundationAccount,
		CoreumRPCURL:            coreumRPCAddress,
		XrplRPCAPIURL:           xrplRPCAPIURL,
		XrplScanAPIURL:          xrplScanAPIURL,
		XrplHistoricalAPIURL:    xrplHistoricalAPIURL,
		XrplAccount:             xrplAccount,
		XrplCurrency:            xrplCurrency,
		XrplIssuer:              xrplIssuer,
		BridgeChainIndex:        bridgeChainIndex,
		OutputDocument:          outputDocument,
		FeeConfigs:              feeConfigs,
		IncludeAll:              includeAll,
		MultichainRescanAPIURL:  multichainRescanAPIURL,
	}, nil
}
