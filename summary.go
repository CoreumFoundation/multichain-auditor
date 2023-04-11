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
		convertFloatToSixDecimalsFloatText(r.CoreumIncomeAmount), convertFloatToSixDecimalsFloatText(r.CoreumOutcomeAmount), convertFloatToSixDecimalsFloatText(r.CoreumBalance),
		convertFloatToSixDecimalsFloatText(r.XrplBurntAmount), convertFloatToSixDecimalsFloatText(r.XrplSupply), r.XrplOrphanTxCount, convertFloatToSixDecimalsFloatText(r.XrplOrphanTxAmount),
		convertFloatToSixDecimalsFloatText(r.FeesAmount),
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
		switch discrepancy.Discrepancy {
		case "":
			xrplBurntAmount = big.NewInt(0).Add(xrplBurntAmount, discrepancy.XrplTx.Amount)
			coreumOutcomeAmount = big.NewInt(0).Add(coreumOutcomeAmount, discrepancy.CoreumTx.Amount)
			feesAmount = big.NewInt(0).Add(feesAmount, big.NewInt(0).Sub(discrepancy.XrplTx.Amount, discrepancy.CoreumTx.Amount))
			continue
		case InfoAmountOutOfRange:
			xrplBurntAmount = big.NewInt(0).Add(xrplBurntAmount, discrepancy.XrplTx.Amount)
			continue
		case DiscrepancyOrphanXrplTx:
			xrplBurntAmount = big.NewInt(0).Add(xrplBurntAmount, discrepancy.XrplTx.Amount)
			xrplOrphanTxCount++
			xrplOrphanTxAmount = big.NewInt(0).Add(xrplOrphanTxAmount, discrepancy.XrplTx.Amount)
			continue
		default:
			noneOrphanDiscrepanciesCount++
		}
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
