package main

import (
	"math/big"
	"sort"
	"strings"
	"time"
)

const (
	DiscrepancyInvalidMemoOnCoreum                     = "invalid memo on coreum"
	DiscrepancyDuplicatedXrplTxHashInMemoOnCoreum      = "duplicated xrpl tx hash in memo on coreum"
	DiscrepancyOrphanXrplTx                            = "orphan xrpl tx"
	DiscrepancyDifferentTargetAddressesOnXrplAndCoreum = "different target addresses on xrpl and coreum"
	DiscrepancyDifferentAmountOnXrplAndCoreum          = "different amount on xrpl and coreum"
	DiscrepancyOrphanCoreumTx                          = "orphan coreum tx"
)

var thousandInt = big.NewInt(1000)

// AuditTx represents chain agnostic unified format of the bridge transaction.
type AuditTx struct {
	Hash          string
	FromAddress   string
	ToAddress     string
	TargetAddress string
	Amount        *big.Int
	Memo          string
	Timestamp     time.Time
}

// TxDiscrepancy represent discrepancy of the xrpl and coreum transactions.
type TxDiscrepancy struct {
	XrplHash          string
	XrplAmount        *big.Int
	XrplTargetAddress string
	XrplMemo          string
	XrplTimestamp     time.Time

	CoreumHash          string
	CoreumAmount        *big.Int
	AmountsWithoutFee   []*big.Int
	CoreumTargetAddress string
	CoreumMemo          string
	CoreumTimestamp     time.Time

	Discrepancy string
}

// FindAuditTxDiscrepancies find the discrepancies between coreum and XRPL transactions.
func FindAuditTxDiscrepancies(xrplTxs, coreumTxs []AuditTx, feeConfigs []FeeConfig, includeAll bool) []TxDiscrepancy {
	discrepancies := make([]TxDiscrepancy, 0)
	xrplTxsMap := make(map[string]AuditTx)
	for _, xrplTx := range xrplTxs {
		xrplTxsMap[strings.ToUpper(xrplTx.Hash)] = xrplTx
	}

	xrplTxHashToCoreumTxMap := make(map[string]AuditTx)
	for _, coreumTx := range coreumTxs {
		xrplTxHash := decodeXrplTxHashFromCoreumMemo(coreumTx.Memo)
		if xrplTxHash == "" {
			discrepancies = append(discrepancies, fillDiscrepancy(AuditTx{}, coreumTx, DiscrepancyInvalidMemoOnCoreum, nil))
			continue
		}

		if _, ok := xrplTxHashToCoreumTxMap[xrplTxHash]; ok {
			discrepancies = append(discrepancies, fillDiscrepancy(AuditTx{}, coreumTx, DiscrepancyDuplicatedXrplTxHashInMemoOnCoreum, nil))
			continue
		}
		xrplTxHashToCoreumTxMap[xrplTxHash] = coreumTx
	}

	for xrplTxHash, xrplTx := range xrplTxsMap {
		coreumTx, ok := xrplTxHashToCoreumTxMap[xrplTxHash]
		if !ok {
			discrepancies = append(discrepancies, fillDiscrepancy(xrplTx, AuditTx{}, DiscrepancyOrphanXrplTx, nil))
			delete(xrplTxsMap, xrplTxHash)
			continue
		}
		if xrplTx.TargetAddress != coreumTx.TargetAddress {
			discrepancies = append(discrepancies, fillDiscrepancy(xrplTx, coreumTx, DiscrepancyDifferentTargetAddressesOnXrplAndCoreum, nil))
			delete(xrplTxsMap, xrplTxHash)
			delete(xrplTxHashToCoreumTxMap, xrplTxHash)
			continue
		}

		amountMatches := false
		amountsWithoutFee := make([]*big.Int, 0)
		for _, feeConfig := range feeConfigs {
			amountWithoutFee := computeAmountWithoutFee(xrplTx.Amount, feeConfig)
			if amountWithoutFee.Cmp(coreumTx.Amount) == 0 {
				amountMatches = true
				break
			}
			amountsWithoutFee = append(amountsWithoutFee, amountWithoutFee)
		}
		if !amountMatches {
			discrepancies = append(discrepancies, fillDiscrepancy(xrplTx, coreumTx, DiscrepancyDifferentAmountOnXrplAndCoreum, amountsWithoutFee))
			delete(xrplTxsMap, xrplTxHash)
			delete(xrplTxHashToCoreumTxMap, xrplTxHash)
			continue
		}

		if includeAll {
			discrepancies = append(discrepancies, fillDiscrepancy(xrplTx, coreumTx, "", nil))
		}

		delete(xrplTxsMap, xrplTxHash)
		delete(xrplTxHashToCoreumTxMap, xrplTxHash)
	}

	for _, coreumTx := range xrplTxHashToCoreumTxMap {
		discrepancies = append(discrepancies, fillDiscrepancy(AuditTx{}, coreumTx, DiscrepancyOrphanCoreumTx, nil))
	}

	sort.Slice(discrepancies, func(i, j int) bool {
		return discrepancies[i].XrplTimestamp.After(discrepancies[j].XrplTimestamp)
	})

	return discrepancies
}

func fillDiscrepancy(xrplTx, coreumTx AuditTx, discrepancy string, amountsWithoutFee []*big.Int) TxDiscrepancy {
	return TxDiscrepancy{
		XrplHash:          xrplTx.Hash,
		XrplAmount:        xrplTx.Amount,
		XrplTargetAddress: xrplTx.TargetAddress,
		XrplMemo:          xrplTx.Memo,
		XrplTimestamp:     xrplTx.Timestamp,

		CoreumHash:          coreumTx.Hash,
		CoreumAmount:        coreumTx.Amount,
		AmountsWithoutFee:   amountsWithoutFee,
		CoreumTargetAddress: coreumTx.TargetAddress,
		CoreumMemo:          coreumTx.Memo,
		CoreumTimestamp:     coreumTx.Timestamp,

		Discrepancy: discrepancy,
	}
}

func decodeXrplTxHashFromCoreumMemo(memo string) string {
	memoFragments := strings.Split(memo, ":")
	if len(memoFragments) != 3 {
		return ""
	}

	return strings.ToUpper(strings.ReplaceAll(memoFragments[1], "0x", ""))
}

func computeAmountWithoutFee(amount *big.Int, config FeeConfig) *big.Int {
	fee := big.NewInt(0).Div(big.NewInt(0).Mul(amount, config.FeeRatio), thousandInt)
	if fee.Cmp(config.MinFee) == -1 {
		fee = config.MinFee
	}
	if fee.Cmp(config.MaxFee) == 1 {
		fee = config.MaxFee
	}

	return big.NewInt(0).Sub(amount, fee)
}
