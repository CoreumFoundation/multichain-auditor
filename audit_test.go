package main

import (
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestFindAuditTxDiscrepancies(t *testing.T) {
	const bridgeChainIndex = "1111"
	onePercentFeeConfigWithMinAndMaxLimits := FeeConfig{
		StartTime: time.Date(2022, time.Month(6), 1, 0, 0, 0, 0, time.UTC),
		FeeRatio:  big.NewInt(0),
		MinFee:    big.NewInt(1000),
		MaxFee:    big.NewInt(0),
		MinAmount: big.NewInt(1000),
		MaxAmount: big.NewInt(10000),
	}
	zeroFeeConfig := FeeConfig{
		StartTime: time.Date(2022, time.Month(0), 1, 0, 0, 0, 0, time.UTC),
		FeeRatio:  big.NewInt(0),
		MinFee:    big.NewInt(0),
		MaxFee:    big.NewInt(0),
		MinAmount: big.NewInt(0),
		MaxAmount: big.NewInt(1_000_000),
	}

	type args struct {
		xrplTxs    []AuditTx
		coreumTxs  []AuditTx
		feeConfigs []FeeConfig
	}
	tests := []struct {
		name string
		args args
		want []TxDiscrepancy
	}{
		{
			name: "positive_all_matches",
			args: args{
				xrplTxs: []AuditTx{
					{
						Hash:          "xrplHash1",
						TargetAddress: "core1",
						Amount:        big.NewInt(123),
						Memo:          "core1:" + bridgeChainIndex,
						Timestamp:     time.Date(2023, time.Month(1), 1, 0, 0, 0, 0, time.UTC),
					},
					{
						Hash:          "xrplHash2",
						TargetAddress: "core2",
						Amount:        big.NewInt(123),             // same amount as prev
						Memo:          "core2:" + bridgeChainIndex, // same address as prev
						Timestamp:     time.Date(2023, time.Month(1), 1, 0, 0, 0, 0, time.UTC),
					},
				},
				coreumTxs: []AuditTx{
					{
						Hash:          "coreHash1",
						TargetAddress: "core1",
						Amount:        big.NewInt(123),
						Memo:          bridgeChainIndex + ":" + "xrplHash1" + ":0",
						Timestamp:     time.Date(2023, time.Month(1), 1, 0, 0, 0, 0, time.UTC),
					},
					{
						Hash:          "coreHash2",
						TargetAddress: "core2",
						Amount:        big.NewInt(123),
						Memo:          bridgeChainIndex + ":" + "xrplHash2" + ":0",
						Timestamp:     time.Date(2023, time.Month(1), 1, 0, 0, 0, 0, time.UTC),
					},
				},
				feeConfigs: []FeeConfig{zeroFeeConfig},
			},
			want: []TxDiscrepancy{},
		},
		{
			name: "positive_skipping_amounts",
			args: args{
				xrplTxs: []AuditTx{
					{
						Hash:          "xrplHash1",
						TargetAddress: "core1",
						Amount:        big.NewInt(1), // low
						Memo:          "core1:" + bridgeChainIndex,
						Timestamp:     time.Date(2023, time.Month(1), 1, 0, 0, 0, 0, time.UTC),
					},
					{
						Hash:          "xrplHash2",
						TargetAddress: "core1",
						Amount:        big.NewInt(2_000_000), // high
						Memo:          "core1:" + bridgeChainIndex,
						Timestamp:     time.Date(2023, time.Month(1), 1, 0, 0, 0, 0, time.UTC),
					},
					{
						Hash:          "xrplHash3", // the tx should get to different fee config
						TargetAddress: "core1",
						Amount:        big.NewInt(123),
						Memo:          "core1:" + bridgeChainIndex,
						Timestamp:     time.Date(2022, time.Month(1), 1, 0, 0, 0, 0, time.UTC),
					},
				},
				coreumTxs: []AuditTx{
					{
						Hash:          "coreHash1",
						TargetAddress: "core1",
						Amount:        big.NewInt(123),
						Memo:          bridgeChainIndex + ":" + "xrplHash3" + ":0",
						Timestamp:     time.Date(2022, time.Month(1), 1, 0, 0, 0, 0, time.UTC),
					},
				},
				// both txs should get the `onePercentFeeConfigWithMinAndMaxLimits`
				feeConfigs: []FeeConfig{zeroFeeConfig, onePercentFeeConfigWithMinAndMaxLimits},
			},
			want: []TxDiscrepancy{},
		},
		{
			name: "negative_invalid_memo_on_coreum",
			args: args{
				xrplTxs: []AuditTx{},
				coreumTxs: []AuditTx{
					{
						Hash:          "coreHash1",
						TargetAddress: "core1",
						Amount:        big.NewInt(123),
						Memo:          "invalid-memo",
						Timestamp:     time.Date(2023, time.Month(1), 1, 0, 0, 0, 0, time.UTC),
					},
				},
			},
			want: []TxDiscrepancy{
				{
					CoreumTx: AuditTx{
						Hash:          "coreHash1",
						Amount:        big.NewInt(123),
						TargetAddress: "core1",
						Memo:          "invalid-memo",
						Timestamp:     time.Date(2023, time.Month(1), 1, 0, 0, 0, 0, time.UTC),
					},

					Discrepancy: DiscrepancyInvalidMemoOnCoreum,
				},
			},
		},
		{
			name: "negative_duplicated_xrpl_tx_hash_in_memo_on_coreum",
			args: args{
				xrplTxs: []AuditTx{
					{
						Hash:          "xrplHash1",
						TargetAddress: "core1",
						Amount:        big.NewInt(123),
						Memo:          "core1:" + bridgeChainIndex,
						Timestamp:     time.Date(2023, time.Month(1), 1, 0, 0, 0, 0, time.UTC),
					},
				},
				coreumTxs: []AuditTx{
					{
						Hash:          "coreHash1",
						TargetAddress: "core1",
						Amount:        big.NewInt(123),
						Memo:          bridgeChainIndex + ":" + "xrplHash1" + ":0",
						Timestamp:     time.Date(2023, time.Month(1), 1, 0, 0, 0, 0, time.UTC),
					},
					{
						Hash:          "coreHash2",
						TargetAddress: "core2",
						Amount:        big.NewInt(123),
						Memo:          bridgeChainIndex + ":" + "xrplHash1" + ":0",
						Timestamp:     time.Date(2023, time.Month(1), 1, 0, 0, 0, 0, time.UTC),
					},
				},
				feeConfigs: []FeeConfig{zeroFeeConfig},
			},
			want: []TxDiscrepancy{
				{
					CoreumTx: AuditTx{
						Hash:          "coreHash2",
						Amount:        big.NewInt(123),
						TargetAddress: "core2",
						Memo:          bridgeChainIndex + ":" + "xrplHash1" + ":0",
						Timestamp:     time.Date(2023, time.Month(1), 1, 0, 0, 0, 0, time.UTC),
					},
					Discrepancy: DiscrepancyDuplicatedXrplTxHashInMemoOnCoreum,
				},
			},
		},
		{
			name: "negative_missing_xrpl_tx_on_coreum",
			args: args{
				xrplTxs: []AuditTx{
					{
						Hash:          "xrplHash1",
						TargetAddress: "core1",
						Amount:        big.NewInt(123),
						Memo:          "core1:" + bridgeChainIndex,
						Timestamp:     time.Date(2023, time.Month(1), 1, 0, 0, 0, 0, time.UTC),
					},
				},
				coreumTxs:  []AuditTx{},
				feeConfigs: []FeeConfig{zeroFeeConfig},
			},
			want: []TxDiscrepancy{
				{
					XrplTx: AuditTx{
						Hash:          "xrplHash1",
						Amount:        big.NewInt(123),
						TargetAddress: "core1",
						Memo:          "core1:" + bridgeChainIndex,
						Timestamp:     time.Date(2023, time.Month(1), 1, 0, 0, 0, 0, time.UTC),
					},
					Discrepancy: DiscrepancyOrphanXrplTx,
				},
			},
		},
		{
			name: "negative_different_target_addresses_on_xrpl_and_coreum",
			args: args{
				xrplTxs: []AuditTx{
					{
						Hash:          "xrplHash1",
						TargetAddress: "core2",
						Amount:        big.NewInt(123),
						Memo:          "core2:" + bridgeChainIndex,
						Timestamp:     time.Date(2023, time.Month(1), 1, 0, 0, 0, 0, time.UTC),
					},
				},
				coreumTxs: []AuditTx{
					{
						Hash:          "coreHash1",
						TargetAddress: "core1",
						Amount:        big.NewInt(123),
						Memo:          bridgeChainIndex + ":" + "xrplHash1" + ":0",
						Timestamp:     time.Date(2023, time.Month(1), 1, 0, 0, 0, 0, time.UTC),
					},
				},
				feeConfigs: []FeeConfig{zeroFeeConfig},
			},
			want: []TxDiscrepancy{
				{
					XrplTx: AuditTx{
						Hash:          "xrplHash1",
						Amount:        big.NewInt(123),
						TargetAddress: "core2",
						Memo:          "core2:" + bridgeChainIndex,
						Timestamp:     time.Date(2023, time.Month(1), 1, 0, 0, 0, 0, time.UTC),
					},
					CoreumTx: AuditTx{
						Hash:          "coreHash1",
						Amount:        big.NewInt(123),
						TargetAddress: "core1",
						Memo:          bridgeChainIndex + ":" + "xrplHash1" + ":0",
						Timestamp:     time.Date(2023, time.Month(1), 1, 0, 0, 0, 0, time.UTC),
					},
					Discrepancy: DiscrepancyDifferentTargetAddressesOnXrplAndCoreum,
				},
			},
		},
		{
			name: "negative_different_amount_on_xrpl_and_coreum",
			args: args{
				xrplTxs: []AuditTx{
					{
						Hash:          "xrplHash1",
						TargetAddress: "core1",
						Amount:        big.NewInt(123),
						Memo:          "core1:" + bridgeChainIndex,
						Timestamp:     time.Date(2023, time.Month(1), 1, 0, 0, 0, 0, time.UTC),
					},
				},
				coreumTxs: []AuditTx{
					{
						Hash:          "coreHash1",
						TargetAddress: "core1",
						Amount:        big.NewInt(124),
						Memo:          bridgeChainIndex + ":" + "xrplHash1" + ":0",
						Timestamp:     time.Date(2023, time.Month(1), 1, 0, 0, 0, 0, time.UTC),
					},
				},
				feeConfigs: []FeeConfig{zeroFeeConfig},
			},
			want: []TxDiscrepancy{
				{
					XrplTx: AuditTx{
						Hash:          "xrplHash1",
						Amount:        big.NewInt(123),
						TargetAddress: "core1",
						Memo:          "core1:" + bridgeChainIndex,
						Timestamp:     time.Date(2023, time.Month(1), 1, 0, 0, 0, 0, time.UTC),
					},
					CoreumTx: AuditTx{
						Hash:          "coreHash1",
						Amount:        big.NewInt(124),
						TargetAddress: "core1",
						Memo:          bridgeChainIndex + ":" + "xrplHash1" + ":0",
						Timestamp:     time.Date(2023, time.Month(1), 1, 0, 0, 0, 0, time.UTC),
					},
					ExpectedAmount: big.NewInt(123),
					Discrepancy:    DiscrepancyDifferentAmountOnXrplAndCoreum,
				},
			},
		},
		{
			name: "negative_orphan_coreum_tx",
			args: args{
				coreumTxs: []AuditTx{
					{
						Hash:          "coreHash1",
						TargetAddress: "core1",
						Amount:        big.NewInt(123),
						Memo:          bridgeChainIndex + ":" + "xrplHash1" + ":0",
						Timestamp:     time.Date(2023, time.Month(1), 1, 0, 0, 0, 0, time.UTC),
					},
				},
				feeConfigs: []FeeConfig{zeroFeeConfig},
			},
			want: []TxDiscrepancy{
				{
					CoreumTx: AuditTx{
						Hash:          "coreHash1",
						Amount:        big.NewInt(123),
						TargetAddress: "core1",
						Memo:          bridgeChainIndex + ":" + "xrplHash1" + ":0",
						Timestamp:     time.Date(2023, time.Month(1), 1, 0, 0, 0, 0, time.UTC),
					},

					Discrepancy: DiscrepancyOrphanCoreumTx,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			afterDateTime := time.Date(2030, time.Month(1), 1, 0, 0, 0, 0, time.UTC)
			beforeDateTime := time.Date(2020, time.Month(1), 1, 0, 0, 0, 0, time.UTC)
			got := FindAuditTxDiscrepancies(tt.args.xrplTxs, tt.args.coreumTxs, tt.args.feeConfigs, false, afterDateTime, beforeDateTime)
			require.Equal(t, tt.want, got)
		})
	}
}

func Test_computeAmountWithFees(t *testing.T) {
	feeConfig := FeeConfig{
		FeeRatio: big.NewInt(1),     // 0.1%
		MinFee:   big.NewInt(7000),  // 0.007 CORE
		MaxFee:   big.NewInt(50000), // 0.05 CORE
	}

	type args struct {
		amount *big.Int
		config FeeConfig
	}
	tests := []struct {
		name string
		args args
		want *big.Int
	}{
		{
			name: "min_fee",
			args: args{
				amount: big.NewInt(1_000_000),
				config: feeConfig,
			},
			want: big.NewInt(993000),
		},
		{
			name: "max_fee",
			args: args{
				amount: big.NewInt(100_000_000_000),
				config: feeConfig,
			},
			want: big.NewInt(99999950000),
		},
		{
			name: "fee_percent",
			args: args{
				amount: big.NewInt(10_000_000),
				config: feeConfig,
			},
			want: big.NewInt(9990000),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := computeAmountWithoutFee(tt.args.amount, tt.args.config)
			require.Equal(t, tt.want, got)
		})
	}
}
