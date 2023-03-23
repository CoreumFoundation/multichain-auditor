package main

import (
	"context"
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/types/module"
	sdktx "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
	"github.com/CoreumFoundation/coreum/pkg/config"
	"github.com/CoreumFoundation/coreum/x/wbank"
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
	fromDateTime, toDateTime time.Time,
) ([]AuditTx, error) {
	bankTxs, err := getTxsWithSingleBankSend(ctx, clientCtx, event, fromDateTime, toDateTime)
	if err != nil {
		return nil, err
	}

	return convertBankTxsToAuditTxs(bankTxs, denom), nil
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
		WithChainID(cfg.ChainID).
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
	fromDateTime, toDateTime time.Time,
) ([]bankSendWithMemo, error) {
	log := logger.Get(ctx)
	log.Info(fmt.Sprintf("Fetching coreum txs from: %s, to: %s ...", fromDateTime.Format(time.DateTime), toDateTime.Format(time.DateTime)))

	tmEvents := []string{event}
	page := 0
	limit := 100 // 100 is the max limit
	var bankSendMessages []bankSendWithMemo
	getMore := true
	for getMore {
		page++
		res, err := authtx.QueryTxsByEvents(clientCtx, tmEvents, page, limit, "")
		if err != nil {
			return nil, err
		}
		if page == 1 {
			log.Info("Fetching coreum txs ...",
				zap.Uint64("Total Items", res.TotalCount),
				zap.Int("PerPage Items", limit),
				zap.Uint64("Total Page", res.PageTotal),
			)
		}

		log.Info("Fetching ...", zap.String("Page", fmt.Sprintf("%d/%d", page, res.PageTotal)))
		if page == int(res.PageTotal) || res.PageTotal == 0 {
			getMore = false
		}
		for _, txAny := range res.Txs {
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
			if timestamp.After(fromDateTime) {
				continue
			}
			if timestamp.Before(toDateTime) {
				getMore = false
				break
			}

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
