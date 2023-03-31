package main

import (
	"fmt"
	"math/big"
)

// Summary represents the summary report data.
type Summary struct {
	CoreumIncomeAmount           *big.Int
	CoreumOutcomeAmount          *big.Int
	CoreumBalance                *big.Int
	XrplBurntAmount              *big.Int
	XrplOrphanTxCount            int
	XrplOrphanTxAmount           *big.Int
	XrplSupply                   *big.Int
	FeesAmount                   *big.Int
	NoneOrphanDiscrepanciesCount int
}

func (r Summary) String() string {
	return fmt.Sprintf(
		"Coreum [IncomeAmount:%s, OutcomeAmount:%s, Balance:%s] \n"+
			"Xrpl   [Burnt:%s, Supply:%s, OrphanTxs:%d, OrphanTxAmount:%s] \n"+
			"Fees: %s \n"+
			"NoneOrphanDiscrepancies: %d",
		r.CoreumIncomeAmount, r.CoreumOutcomeAmount, r.CoreumBalance,
		r.XrplBurntAmount, r.XrplSupply, r.XrplOrphanTxCount, r.XrplOrphanTxAmount,
		r.FeesAmount,
		r.NoneOrphanDiscrepanciesCount,
	)
}

func BuildSummary(
	discrepancies []TxDiscrepancy,
	coreumIncomingTxs []AuditTx,
	coreumBalance,
	xrplSupply *big.Int,
) Summary {
	coreumIncomeAmount := big.NewInt(0)
	for _, coreumInTx := range coreumIncomingTxs {
		coreumIncomeAmount = big.NewInt(0).Add(coreumIncomeAmount, coreumInTx.Amount)
	}

	coreumOutcomeAmount := big.NewInt(0)
	xrplOrphanTxCount := 0
	xrplOrphanTxAmount := big.NewInt(0)
	xrplBurntAmount := big.NewInt(0)
	feesAmount := big.NewInt(0)
	noneOrphanDiscrepanciesCount := 0
	for _, discrepancy := range discrepancies {
		if discrepancy.Discrepancy == "" {
			xrplBurntAmount = big.NewInt(0).Add(xrplBurntAmount, discrepancy.XrplTx.Amount)
			coreumOutcomeAmount = big.NewInt(0).Add(coreumOutcomeAmount, discrepancy.CoreumTx.Amount)
			feesAmount = big.NewInt(0).Add(feesAmount, big.NewInt(0).Sub(discrepancy.XrplTx.Amount, discrepancy.CoreumTx.Amount))
			continue
		}
		if discrepancy.Discrepancy == DiscrepancyOrphanXrplTx {
			xrplBurntAmount = big.NewInt(0).Add(xrplBurntAmount, discrepancy.XrplTx.Amount)
			xrplOrphanTxCount++
			xrplOrphanTxAmount = big.NewInt(0).Add(xrplOrphanTxAmount, discrepancy.XrplTx.Amount)
			continue
		}
		noneOrphanDiscrepanciesCount++
	}

	return Summary{
		CoreumIncomeAmount:           coreumIncomeAmount,
		CoreumOutcomeAmount:          coreumOutcomeAmount,
		CoreumBalance:                coreumBalance,
		XrplBurntAmount:              xrplBurntAmount,
		XrplSupply:                   xrplSupply,
		XrplOrphanTxCount:            xrplOrphanTxCount,
		XrplOrphanTxAmount:           xrplOrphanTxAmount,
		FeesAmount:                   feesAmount,
		NoneOrphanDiscrepanciesCount: noneOrphanDiscrepanciesCount,
	}
}
