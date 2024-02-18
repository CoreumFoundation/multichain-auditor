package main

import (
	"context"
	"fmt"
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
	XrplFetchPoolSize       int
	BridgeChainIndex        string
	OutputDocument          string
	FeeConfigs              []FeeConfig
	IncludeAll              bool
	MultichainRescanAPIURL  string
	ManualBridgeTxSender    string
	NonProcessedAmounts     map[string]int
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

	xrplFetchPullSize, err := cmd.Flags().GetInt(xrplFetchPoolSizeFlag)
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

	manualBridgeTxSender := ""
	if cmd.Flags().Lookup(manualBridgeTxSenderFlag) != nil {
		manualBridgeTxSender, err = cmd.Flags().GetString(manualBridgeTxSenderFlag)
		if err != nil {
			return Config{}, err
		}
		fmt.Println("manualBridgeTxSender" + manualBridgeTxSender)
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

	nonProcessedAmounts := map[string]int{
		"core13ntfv565uvlp0x6gtkqkkt8s7jrr6l4dhjw5nk": 7247848871,
		"core1w93heglekxfpwud66ep92yjyaz6rfnh7jyduvf": 10000000 + 17695470, // 2 txs from this account.
		"core1fm9w57hcq5984d07ku9ewfwaxt89w64uh5cqu8": 2503000000,
		"core10t696zghgs55gszcng00kzkrqmgpr59k7t7md5": 101055856298,
		"core1emdkcuznuh2hrar22w4wmdv5j53lqpt883agft": 6000000,
		"core184cn6se86h4ghnfvcz8zjz8wsdas20jej89wqc": 10000000,
		"core163xhdn0u3ne736avtnen5gqv54kk4qhansjyc7": 2000000000,
		"core1qnvwd8vkryk53r6eaew8qzjafh3uyctpexjwlw": 5000000,
		"core1rspkvw33f2pek3h93g9rzzrmqhaxtql5fqrlsm": 216000000,
		"core1zeu60z4lwjf752kpjrxhc46yg4ukv0xzl680m2": 10000000000,
		"core1m02v6wt6klldkn4psau5zs44juzry4qn5zwaly": 700240293,
		"core1v5f05m2vrjp4lagw7830ahp227v0nrxnzuwc20": 300617536,
		"core1szdmerkyf7hdk6s3n78ajv439hvk43z8jrfjlm": 2021182735,
		"core1f8rjxs9qqy9vuq3amu0mdu96c6r54x6jgydssy": 555000000,
		"core1rravmcp3mhzwclvhp5kwdcqw838kdla4cyfpr6": 105000000,
		"core1v27fg3jjk4m9v673mq8n6hpsfzhhnkkex68f76": 2340000000,
		"core125stq4q8m554qj6f24e5w2cy8r054swz97uygc": 10000000,
		"core16w7tjgj2xj5cpqce4hprkkpamxs3pj6qsqdjc5": 1013000000,
		"core1evg36pmmdle5hxna3xxll420l6u5a72eq44suq": 242156300,
		"core127uq9ued37fzv4kvvywmecclsnqh6s6xer98zf": 27000000000,
		"core1nhvcg07wss35l30q9s6xhzse5dj6gksjxhfw04": 5000000,
		"core18ml0p4cgk4hehwqmkxadtjz4yjamlkhy8d4v8u": 10000000,
		"core103tvnrw7wc0uc9nf4dam73chhdf4hx6nscl6lz": 2291505364,
		"core1xtdms250hk5p3rksjf6yqs9vmqwm7qv82v6w70": 204708550,
		"core13tphl9xqlaazyvpc8vudgmvtwrrfmn072596ky": 134199062,
		"core1lh00c9fn7nauccz8uyxw6l4vc2ucsnaz356yhu": 60000000,
		"core1p27qk09x7vag6523rvxymhdnyul6wc6ahxrm6p": 141744708,
	}

	return Config{
		BeforeDateTime:          beforeDateTime.UTC(),
		AfterDateTime:           afterDateTime.UTC(),
		Denom:                   network.Denom(),
		CoreumAccount:           coreumAccount,
		CoreumFoundationAccount: coreumFoundationAccount,
		CoreumRPCURL:            coreumRPCAddress,
		XrplFetchPoolSize:       xrplFetchPullSize,
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
		ManualBridgeTxSender:    manualBridgeTxSender,
		NonProcessedAmounts:     nonProcessedAmounts,
	}, nil
}
