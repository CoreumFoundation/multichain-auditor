package main

import (
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestFindAuditTxDiscrepancies(t *testing.T) {
	const bridgeChainIndex = "1111"
	zerFeeConfig := FeeConfig{
		FeeRatio: big.NewInt(0),
		MinFee:   big.NewInt(0),
		MaxFee:   big.NewInt(0),
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
				feeConfigs: []FeeConfig{zerFeeConfig},
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
					CoreumHash:          "coreHash1",
					CoreumAmount:        big.NewInt(123),
					CoreumTargetAddress: "core1",
					CoreumMemo:          "invalid-memo",
					CoreumTimestamp:     time.Date(2023, time.Month(1), 1, 0, 0, 0, 0, time.UTC),
					Discrepancy:         DiscrepancyInvalidMemoOnCoreum,
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
				feeConfigs: []FeeConfig{zerFeeConfig},
			},
			want: []TxDiscrepancy{
				{
					CoreumHash:          "coreHash2",
					CoreumAmount:        big.NewInt(123),
					CoreumTargetAddress: "core2",
					CoreumMemo:          bridgeChainIndex + ":" + "xrplHash1" + ":0",
					CoreumTimestamp:     time.Date(2023, time.Month(1), 1, 0, 0, 0, 0, time.UTC),
					Discrepancy:         DiscrepancyDuplicatedXrplTxHashInMemoOnCoreum,
				},
			},
		},
		{
			name: "negative_missing_xrpl_tx_on _oreum",
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
				coreumTxs: []AuditTx{},
			},
			want: []TxDiscrepancy{
				{
					XrplHash:          "xrplHash1",
					XrplAmount:        big.NewInt(123),
					XrplTargetAddress: "core1",
					XrplMemo:          "core1:" + bridgeChainIndex,
					XrplTimestamp:     time.Date(2023, time.Month(1), 1, 0, 0, 0, 0, time.UTC),
					Discrepancy:       DiscrepancyOrphanXrplTx,
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
			},
			want: []TxDiscrepancy{
				{
					XrplHash:          "xrplHash1",
					XrplAmount:        big.NewInt(123),
					XrplTargetAddress: "core2",
					XrplMemo:          "core2:" + bridgeChainIndex,
					XrplTimestamp:     time.Date(2023, time.Month(1), 1, 0, 0, 0, 0, time.UTC),

					CoreumHash:          "coreHash1",
					CoreumAmount:        big.NewInt(123),
					CoreumTargetAddress: "core1",
					CoreumMemo:          bridgeChainIndex + ":" + "xrplHash1" + ":0",
					CoreumTimestamp:     time.Date(2023, time.Month(1), 1, 0, 0, 0, 0, time.UTC),

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
				feeConfigs: []FeeConfig{zerFeeConfig},
			},
			want: []TxDiscrepancy{
				{
					XrplHash:          "xrplHash1",
					XrplAmount:        big.NewInt(123),
					XrplTargetAddress: "core1",
					XrplMemo:          "core1:" + bridgeChainIndex,
					XrplTimestamp:     time.Date(2023, time.Month(1), 1, 0, 0, 0, 0, time.UTC),

					CoreumHash:          "coreHash1",
					CoreumAmount:        big.NewInt(124),
					AmountsWithoutFee:   []*big.Int{big.NewInt(123)},
					CoreumTargetAddress: "core1",
					CoreumMemo:          bridgeChainIndex + ":" + "xrplHash1" + ":0",
					CoreumTimestamp:     time.Date(2023, time.Month(1), 1, 0, 0, 0, 0, time.UTC),

					Discrepancy: DiscrepancyDifferentAmountOnXrplAndCoreum,
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
			},
			want: []TxDiscrepancy{
				{
					CoreumHash:          "coreHash1",
					CoreumAmount:        big.NewInt(123),
					CoreumTargetAddress: "core1",
					CoreumMemo:          bridgeChainIndex + ":" + "xrplHash1" + ":0",
					CoreumTimestamp:     time.Date(2023, time.Month(1), 1, 0, 0, 0, 0, time.UTC),

					Discrepancy: DiscrepancyOrphanCoreumTx,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FindAuditTxDiscrepancies(tt.args.xrplTxs, tt.args.coreumTxs, tt.args.feeConfigs, false)
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
