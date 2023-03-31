package main

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildSummary(t *testing.T) {
	discrepancies := []TxDiscrepancy{
		// no discrepancy
		{
			XrplTx: AuditTx{
				Amount: big.NewInt(100),
			},
			CoreumTx: AuditTx{
				Amount: big.NewInt(90),
			},
		},
		// no discrepancy
		{
			XrplTx: AuditTx{
				Amount: big.NewInt(90),
			},
			CoreumTx: AuditTx{
				Amount: big.NewInt(80),
			},
		},
		// orphan discrepancy
		{
			XrplTx: AuditTx{
				Amount: big.NewInt(25),
			},
			Discrepancy: DiscrepancyOrphanXrplTx,
		},
		// orphan discrepancy
		{
			XrplTx: AuditTx{
				Amount: big.NewInt(10),
			},
			Discrepancy: DiscrepancyOrphanXrplTx,
		},
		// none orphan discrepancy
		{
			XrplTx: AuditTx{
				Amount: big.NewInt(10),
			},
			Discrepancy: DiscrepancyDifferentAmountOnXrplAndCoreum,
		},
	}
	coreumIncomingTxs := []AuditTx{
		{
			Amount: big.NewInt(350),
		},
		{
			Amount: big.NewInt(20),
		},
	}

	coreumBalance := big.NewInt(333)
	xrplSupply := big.NewInt(555)

	got := BuildSummary(discrepancies, coreumIncomingTxs, coreumBalance, xrplSupply)
	want := Summary{
		CoreumIncomeAmount:           big.NewInt(370),
		CoreumOutcomeAmount:          big.NewInt(170),
		CoreumBalance:                big.NewInt(333),
		XrplBurntAmount:              big.NewInt(225),
		XrplSupply:                   big.NewInt(555),
		XrplOrphanTxCount:            2,
		XrplOrphanTxAmount:           big.NewInt(35),
		FeesAmount:                   big.NewInt(20),
		NoneOrphanDiscrepanciesCount: 1,
	}

	require.Equal(t, want, got)
}
