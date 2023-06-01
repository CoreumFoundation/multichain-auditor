package main

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	sdktx "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/gammazero/workerpool"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
	"github.com/CoreumFoundation/coreum/pkg/config"
	"github.com/CoreumFoundation/coreum/pkg/config/constant"
	"github.com/CoreumFoundation/coreum/x/wbank"
)

const (
	coreumTxFetcherPoolSize = 100
)

type bankSendWithMemo struct {
	Hash string
	*banktypes.MsgSend
	Memo      string
	Timestamp time.Time
}

// GetCoreumAuditTransactions returns the list of the valid coreum bridge transaction converted to the audit model.
func GetCoreumAuditTransactions(
	ctx context.Context,
	clientCtx client.Context,
	event, denom string,
	beforeDateTime, afterDateTime time.Time,
) ([]AuditTx, error) {
	bankTxs, err := getTxsWithSingleBankSend(ctx, clientCtx, event, beforeDateTime, afterDateTime)
	if err != nil {
		return nil, err
	}

	return convertBankTxsToAuditTxs(bankTxs, denom), nil
}

// GetCoreumAccountBalance returns the coreum account balance.
func GetCoreumAccountBalance(ctx context.Context, clientCtx client.Context, account, denom string) (*big.Int, error) {
	bankClient := banktypes.NewQueryClient(clientCtx)
	res, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: account,
		Denom:   denom,
	})
	if err != nil {
		return nil, errors.Errorf("can't get account %s balance, err: %s", account, err)
	}

	return res.Balance.Amount.BigInt(), nil
}

func createClientContext(cfg Config) client.Context {
	// List required modules.
	// If you need types from any other module import them and add here.
	modules := module.NewBasicManager(
		auth.AppModuleBasic{},
		wbank.AppModuleBasic{},
	)

	rpcClient, err := client.NewClientFromNode(cfg.CoreumRPCURL)
	if err != nil {
		panic(err)
	}

	encodingConfig := config.NewEncodingConfig(modules)
	clientCtx := client.Context{}.
		WithChainID(string(constant.ChainIDMain)).
		WithClient(rpcClient).
		WithCodec(encodingConfig.Codec).
		WithInterfaceRegistry(encodingConfig.InterfaceRegistry).
		WithTxConfig(encodingConfig.TxConfig).
		WithLegacyAmino(encodingConfig.Amino)

	return clientCtx
}

// getTxsWithSingleBankSend returns transactions filtered by the provided event and time.
// It assumes that all the transactions contain only a single bank send, and errors out
// if this is not true. We can start with this assumption and write more complicated type assertion
// if we face errors.
func getTxsWithSingleBankSend(
	ctx context.Context,
	clientCtx client.Context,
	event string,
	beforeDateTime, afterDateTime time.Time,
) ([]bankSendWithMemo, error) {
	log := logger.Get(ctx)
	log.Info(fmt.Sprintf("Fetching coreum txs before: %s, after: %s ...", beforeDateTime.Format(time.DateTime), afterDateTime.Format(time.DateTime)))

	tmEvents := []string{event}

	limit := 100 // 100 is the max limit
	var bankSendMessages []bankSendWithMemo

	// allocate limited pool to fetch tx in parallel
	workerPool := workerpool.New(coreumTxFetcherPoolSize)
	defer workerPool.Stop()

	page := uint64(1)
	res, err := authtx.QueryTxsByEvents(clientCtx, tmEvents, int(page), limit, "")
	if err != nil {
		return nil, err
	}
	log.Info("Fetching", zap.String("Page", fmt.Sprintf("%d/%d", page, res.PageTotal)))
	page++

	txs := make([]*sdk.TxResponse, 0)
	txs = append(txs, res.Txs...)

	mu := sync.Mutex{}
	wg := sync.WaitGroup{}
	var fetchError error
	wg.Add(int(res.PageTotal) - 1) // we already fetched first
	for ; page <= res.PageTotal; page++ {
		pageToFetch := page
		workerPool.Submit(func() {
			defer wg.Done()
			log.Info("Fetching", zap.String("Page", fmt.Sprintf("%d/%d", pageToFetch, res.PageTotal)))
			res, err = authtx.QueryTxsByEvents(clientCtx, tmEvents, int(pageToFetch), limit, "")
			if err != nil {
				fetchError = err
				log.Error("Can't fetch page", zap.String("Page", fmt.Sprintf("%d", pageToFetch)), zap.Error(err))
				return
			}
			mu.Lock()
			defer mu.Unlock()
			txs = append(txs, res.Txs...)
		})
	}
	wg.Wait()
	if fetchError != nil {
		return nil, fetchError
	}

	for _, txAny := range txs {
		tx, ok := txAny.Tx.GetCachedValue().(*sdktx.Tx)
		if !ok {
			return nil, errors.New("tx does not implement sdk.Tx interface")
		}

		messages := tx.GetMsgs()
		if len(messages) != 1 {
			return nil, errors.New("there should be only 1 message in the transaction")
		}

		msg := messages[0]
		bankSend, ok := msg.(*banktypes.MsgSend)
		if !ok {
			return nil, errors.New("message is not bank MsgSend type")
		}
		timestamp, err := time.Parse(time.RFC3339, txAny.Timestamp)

		if err != nil {
			return nil, errors.Errorf("can't parse time: %s with format %s", txAny.Timestamp, time.RFC3339)
		}
		bankSendMessages = append(bankSendMessages, bankSendWithMemo{
			Hash:      txAny.TxHash,
			MsgSend:   bankSend,
			Memo:      tx.Body.Memo,
			Timestamp: timestamp,
		})
	}

	log.Info(fmt.Sprintf("Found coreum txs total: %d", len(bankSendMessages)))

	return bankSendMessages, nil
}

func convertBankTxsToAuditTxs(coreumTxs []bankSendWithMemo, denom string) []AuditTx {
	txs := make([]AuditTx, 0, len(coreumTxs))
	for _, coreumTx := range coreumTxs {
		txs = append(txs, AuditTx{
			Hash:          coreumTx.Hash,
			FromAddress:   coreumTx.FromAddress,
			ToAddress:     coreumTx.ToAddress,
			TargetAddress: coreumTx.ToAddress,
			Amount:        coreumTx.Amount.AmountOf(denom).BigInt(),
			Memo:          coreumTx.Memo,
			Timestamp:     coreumTx.Timestamp,
		})
	}

	return txs
}
