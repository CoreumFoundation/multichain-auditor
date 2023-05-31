package main

import (
	"encoding/json"
	"io"
	"math/big"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/pkg/errors"
)

const (
	DiscrepancyInvalidMemoOnCoreum                     = "invalid memo on coreum"
	DiscrepancyDuplicatedXrplTxHashInMemoOnCoreum      = "duplicated xrpl tx hash in memo on coreum"
	DiscrepancyOrphanXrplTx                            = "orphan xrpl tx"
	DiscrepancyDifferentTargetAddressesOnXrplAndCoreum = "different target addresses on xrpl and coreum"
	DiscrepancyDifferentAmountOnXrplAndCoreum          = "different amount on xrpl and coreum"
	DiscrepancyOrphanCoreumTx                          = "orphan coreum tx"

	InfoAmountOutOfRange = "not a discrepancy: amount out of range"
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
	XrplTx   AuditTx
	CoreumTx AuditTx

	ExpectedAmount *big.Int
	BridgingTime   time.Duration
	Discrepancy    string
}

// FindAuditTxDiscrepancies find the discrepancies between coreum and XRPL transactions.
func FindAuditTxDiscrepancies(
	xrplTxs, coreumTxs []AuditTx,
	feeConfigs []FeeConfig,
	includeAll bool,
	beforeDateTime, afterDateTime time.Time,
) []TxDiscrepancy {
	allTxs := make([]AuditTx, 0, len(xrplTxs)+len(coreumTxs))
	allTxs = append(allTxs, xrplTxs...)
	allTxs = append(allTxs, coreumTxs...)

	sort.Slice(allTxs, func(i, j int) bool {
		return allTxs[i].Timestamp.Before(allTxs[j].Timestamp)
	})

	collectResults(allTxs)
	computeMissingBalances()

	return nil
}

func collectResults(txs []AuditTx) {
	f, err := os.Create("txs.json")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	for _, tx := range txs {
		if err := enc.Encode(tx); err != nil {
			panic(err)
		}
	}
}

func computeMissingBalances() {
	feeConfig := FeeConfig{
		StartTime: time.Date(2023, time.Month(3), 24, 17, 0, 0, 0, time.UTC),
		FeeRatio:  big.NewInt(1),                // 0.1%
		MinFee:    big.NewInt(2_400000),         // 2.4 CORE
		MaxFee:    big.NewInt(477_000000),       // 477 CORE
		MinAmount: big.NewInt(4_800000),         // 4.8 CORE
		MaxAmount: big.NewInt(2_400_000_000000), // 2.400.000 CORE
	}

	f, err := os.Open("txs.json")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	// FIXME (wojtek): this should use sdk.Int!!!!
	toSend := map[string]*big.Int{}
	zero := big.NewInt(0)
	dec := json.NewDecoder(f)
	for {
		var tx AuditTx
		err := dec.Decode(&tx)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			panic(err)
		}

		if toSend[tx.TargetAddress] == nil {
			toSend[tx.TargetAddress] = zero
		}

		if strings.HasPrefix(tx.ToAddress, "core") {
			toSend[tx.TargetAddress] = new(big.Int).Add(toSend[tx.TargetAddress], new(big.Int).Neg(tx.Amount))
		} else {
			toSend[tx.TargetAddress] = new(big.Int).Add(toSend[tx.TargetAddress], computeAmountWithoutFee(tx.Amount, feeConfig))
		}
	}

	addresses := make([]string, 0, len(toSend))
	for addr, amount := range toSend {
		if amount.Cmp(zero) != 0 {
			addresses = append(addresses, addr)
		}
	}
	sort.Strings(addresses)

	f2, err := os.Create("diff.txt")
	if err != nil {
		panic(err)
	}
	for _, addr := range addresses {

		_, err := f2.Write([]byte(addr + ": " + toSend[addr].String() + "\n"))
		if err != nil {
			panic(err)
		}
	}
}

func fillDiscrepancy(xrplTx, coreumTx AuditTx, discrepancy string, expectedAmount *big.Int) TxDiscrepancy {
	bridgingTime := time.Duration(0)
	if !xrplTx.Timestamp.IsZero() && !coreumTx.Timestamp.IsZero() {
		bridgingTime = coreumTx.Timestamp.Sub(xrplTx.Timestamp)
	}

	return TxDiscrepancy{
		XrplTx:         xrplTx,
		CoreumTx:       coreumTx,
		ExpectedAmount: expectedAmount,
		BridgingTime:   bridgingTime,
		Discrepancy:    discrepancy,
	}
}

func decodeXrplTxHashFromCoreumMemo(memo string) string {
	memoFragments := strings.Split(memo, ":")
	if len(memoFragments) != 3 {
		return ""
	}

	return strings.ToUpper(strings.ReplaceAll(memoFragments[1], "0x", ""))
}

// computeAmountWithoutFee computes the correct fee based on the fee config
// fee = amount * feeRatio/1000
// if fee <= minFee, fee = minFee
// if fee >= maxFee, fee = maxFee
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
