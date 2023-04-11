package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gammazero/workerpool"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
)

var (
	xrplRequestTimeout          = 10 * time.Second
	xrplHistoricalDataPageLimit = 1000 // this limit is maximum for the historical API
	xrplReceivedTxType          = "received"
	oneMillionFloat             = big.NewFloat(1_000_000)
	xrplResStatusSuccess        = "success"
	xrplGetTxRetries            = 10
	xrplGetTxRetryTimout        = 500 * time.Millisecond
)

// Historical models

type xrplAccountTransactionPayment struct {
	TxHash string `json:"tx_hash"`
}

type xrplAccountTransactionsResp struct {
	Result   string                          `json:"result"`
	Marker   string                          `json:"marker"`
	Payments []xrplAccountTransactionPayment `json:"payments"`
}

// RPC API models

type xrplTransactionRequestParams struct {
	Transaction string `json:"transaction"`
	Binary      bool   `json:"binary"`
}

type xrplTransactionRequest struct {
	Method string                         `json:"method"`
	Params []xrplTransactionRequestParams `json:"params"`
}

type xrplMetaDeliveredAmount struct {
	Currency string     `json:"currency"`
	Issuer   string     `json:"issuer"`
	Value    *big.Float `json:"value,string"`
}

type xrplMeta struct {
	DeliveredAmount xrplMetaDeliveredAmount `json:"delivered_amount"`
}

type xrplMemoItem struct {
	MemoData string `json:"MemoData"` // hex string
	MemoType string `json:"MemoType"` // hex string
}

type xrplMemo struct {
	Memo xrplMemoItem `json:"Memo"`
}

type xrplTransaction struct {
	Account         string     `json:"Account"`
	Destination     string     `json:"Destination"`
	Meta            xrplMeta   `json:"meta"`
	Memos           []xrplMemo `json:"Memos"`
	Hash            string     `json:"hash"`
	TransactionType string     `json:"TransactionType"`
	Status          string     `json:"status"`
	Date            int        `json:"date"`
}

type xrplTransactionResp struct {
	Result xrplTransaction `json:"result"`
}

type xrplCurrencySupply struct {
	Currency string     `json:"currency"`
	Value    *big.Float `json:"value,string"`
}

// GetXRPLAuditTransactions returns the list of the valid xrpl bridge transaction converted to the audit model.
func GetXRPLAuditTransactions(
	ctx context.Context,
	fetcherPoolSize int,
	rpcAPIURL, historicalAPIURL, account, currency, issuer, bridgeChainIndex string,
	beforeDateTime, afterDateTime time.Time,
) ([]AuditTx, error) {
	txs, err := getXRPLPaymentTransactions(
		ctx, fetcherPoolSize, rpcAPIURL, historicalAPIURL, account, currency, issuer, beforeDateTime, afterDateTime,
	)
	if err != nil {
		return nil, err
	}

	filteredTxs := filterXRPLBridgeTransactionsAndConvertToTxAudit(bridgeChainIndex, txs)
	logger.Get(ctx).Info(fmt.Sprintf("Found xrpl txs total after bridge related filtration: %d", len(filteredTxs)))

	return filteredTxs, nil
}

// GetXrplCurrencySupply returns the supply of the currency on xrpl.
func GetXrplCurrencySupply(ctx context.Context, baseURL, issuer, currency string) (*big.Int, error) {
	url := fmt.Sprintf("%s/api/v1/account/%s/obligations", baseURL, issuer)
	reqCtx, reqCtxCancel := context.WithTimeout(ctx, xrplRequestTimeout)
	defer reqCtxCancel()
	var resBody []xrplCurrencySupply
	err := DoJSON(reqCtx, http.MethodGet, url, nil, &resBody)
	if err != nil {
		return nil, err
	}

	for _, issuerCurrency := range resBody {
		if issuerCurrency.Currency == currency {
			return convertFloatToSixDecimalsInt(issuerCurrency.Value), nil
		}
	}

	return nil, errors.Errorf("currency %s not found for %s", currency, issuer)
}

// filterXRPLBridgeTransactionsAndConvertToTxAudit filters the list of the xrpl transactions to leave the bridge only
// and converts them all to tx audit transactions.
func filterXRPLBridgeTransactionsAndConvertToTxAudit(bridgeChainIndex string, txs []xrplTransaction) []AuditTx {
	filteredTxs := make([]AuditTx, 0)
	for _, tx := range txs {
		var (
			memo    string
			address string
			ok      bool
		)
		for _, memoItem := range tx.Memos {
			address, memo, ok = decodeXRPLBridgeMemo(memoItem.Memo.MemoData, bridgeChainIndex)
			if ok {
				break
			}
		}
		if !ok {
			continue
		}
		amount := convertFloatToSixDecimalsInt(tx.Meta.DeliveredAmount.Value)
		if amount.Cmp(big.NewInt(0)) != 1 {
			continue
		}

		timestamp := convertXRPLDateToTime(tx.Date)

		filteredTxs = append(filteredTxs, AuditTx{
			Hash:          tx.Hash,
			FromAddress:   tx.Account,
			ToAddress:     tx.Destination,
			TargetAddress: address,
			Amount:        amount,
			Memo:          memo,
			Timestamp:     timestamp,
		})
	}

	sort.Slice(filteredTxs, func(i, j int) bool {
		return filteredTxs[i].Timestamp.After(filteredTxs[j].Timestamp)
	})

	return filteredTxs
}

// getXRPLPaymentTransactions fetches all payment transactions from xrpl for the specified account and fill them with
// the full set of required attributes.
func getXRPLPaymentTransactions(
	ctx context.Context,
	fetcherPoolSize int,
	rpcAPIURL, historicalAPIURL, account, currency, issuer string,
	beforeDateTime, afterDateTime time.Time,
) ([]xrplTransaction, error) {
	log := logger.Get(ctx)
	log.Info(fmt.Sprintf("Fetching xrpl txs before: %s, after: %s ...", beforeDateTime.Format(time.DateTime), afterDateTime.Format(time.DateTime)))

	wg := sync.WaitGroup{}
	mu := sync.Mutex{}
	txs := make([]xrplTransaction, 0)

	// allocate limited pool to fetch tx in parallel
	workerPool := workerpool.New(fetcherPoolSize)
	defer workerPool.Stop()

	marker := "" // empty marker indicates that we fetch from latest
	page := 1
	for {
		var (
			txHashes []string
			err      error
		)

		log.Info("Fetching", zap.String("Page", fmt.Sprintf("%d", page)))
		txHashes, marker, err = getXRPLHistoricalPaymentTxHashes(
			ctx, historicalAPIURL, account, currency, issuer, marker, beforeDateTime, afterDateTime,
		)
		if err != nil {
			return nil, err
		}
		page++
		wg.Add(len(txHashes))
		for _, txHash := range txHashes {
			txHashCopy := txHash
			workerPool.Submit(
				func() {
					defer wg.Done()
					var tx xrplTransaction
					tx, err = getXRPLTxWithRetry(ctx, rpcAPIURL, txHashCopy)
					if err != nil {
						return
					}
					mu.Lock()
					defer mu.Unlock()
					txs = append(txs, tx)
				},
			)
		}
		wg.Wait()
		if err != nil {
			return nil, err
		}
		// if marker is empty no pages are left
		if marker == "" {
			break
		}
	}

	log.Info(fmt.Sprintf("Found xrpl txs total: %d", len(txs)))

	return txs, nil
}

func getXRPLHistoricalPaymentTxHashes(
	ctx context.Context,
	baseURL, account, currency, issuer, marker string, beforeDateTime, afterDateTime time.Time,
) ([]string, string, error) {
	url := fmt.Sprintf("%s/v2/accounts/%s/payments/?type=%s&currency=%s&issuer=%s&marker=%s&limit=%d&end=%s&start=%s",
		baseURL, account, xrplReceivedTxType, currency, issuer, marker, xrplHistoricalDataPageLimit, beforeDateTime.Format(time.RFC3339), afterDateTime.Format(time.RFC3339))
	reqCtx, reqCtxCancel := context.WithTimeout(ctx, xrplRequestTimeout)
	defer reqCtxCancel()
	var resBody xrplAccountTransactionsResp
	err := DoJSON(reqCtx, http.MethodGet, url, nil, &resBody)
	if err != nil {
		return nil, "", err
	}
	if resBody.Result != xrplResStatusSuccess {
		return nil, "", errors.Errorf("receive unexpected result status: %s", resBody.Result)
	}
	txs := make([]string, 0, len(resBody.Payments))
	for _, txPayment := range resBody.Payments {
		txs = append(txs, txPayment.TxHash)
	}

	return txs, resBody.Marker, nil
}

func getXRPLTxWithRetry(ctx context.Context, baseURL, txHash string) (xrplTransaction, error) {
	reqBody := xrplTransactionRequest{
		Method: "tx",
		Params: []xrplTransactionRequestParams{
			{
				Transaction: txHash,
				Binary:      false,
			},
		},
	}

	var (
		resBody xrplTransactionResp
		err     error
	)
	for i := 0; i < xrplGetTxRetries; i++ {
		reqCtx, reqCtxCancel := context.WithTimeout(ctx, xrplRequestTimeout)
		err := DoJSON(reqCtx, http.MethodPost, baseURL, reqBody, &resBody)
		reqCtxCancel()
		if err == nil {
			return resBody.Result, nil
		}
		if i != xrplGetTxRetries-1 {
			<-time.After(xrplGetTxRetryTimout)
		}
	}

	return xrplTransaction{}, errors.Errorf("can't get xrpl tx %s by hash with %d retries and timeout %s, last response: %v, last err: %s",
		txHash, xrplGetTxRetries, xrplGetTxRetryTimout, resBody, err)
}

func decodeXRPLBridgeMemo(hexMemo, bridgeChainIndex string) (string, string, bool) {
	memo, err := hex.DecodeString(hexMemo)
	if err != nil {
		return "", "", false
	}
	memoFragments := strings.Split(string(memo), ":")
	if len(memoFragments) != 2 {
		return "", "", false
	}

	if memoFragments[1] != bridgeChainIndex {
		return "", "", false
	}

	return memoFragments[0], string(memo), true
}

func convertFloatToSixDecimalsInt(amount *big.Float) *big.Int {
	if amount == nil {
		return big.NewInt(0)
	}
	convertedAmount, _ := big.NewFloat(0).Mul(amount, oneMillionFloat).Int(nil)
	return convertedAmount
}

func convertXRPLDateToTime(xrplDate int) time.Time {
	txTime := time.Date(2000, time.Month(1), 1, 0, 0, 0, 0, time.UTC)
	return txTime.Add(time.Duration(xrplDate) * time.Second)
}
