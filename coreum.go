package main

import (
	"context"
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
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

type BankSendWithMemo struct {
	Hash string
	*banktypes.MsgSend
	Memo      string
	Timestamp time.Time
}

func createClientContext(cfg Config) client.Context {
	// List required modules.
	// If you need types from any other module import them and add here.
	modules := module.NewBasicManager(
		auth.AppModuleBasic{},
		wbank.AppModuleBasic{},
	)

	rpcClient, err := client.NewClientFromNode(cfg.coreumRPCAddress)
	if err != nil {
		panic(err)
	}

	encodingConfig := config.NewEncodingConfig(modules)
	clientCtx := client.Context{}.
		WithChainID(string(cfg.chainID)).
		WithClient(rpcClient).
		WithKeyring(keyring.NewInMemory()).
		WithBroadcastMode(flags.BroadcastBlock).
		WithCodec(encodingConfig.Codec).
		WithInterfaceRegistry(encodingConfig.InterfaceRegistry).
		WithTxConfig(encodingConfig.TxConfig).
		WithLegacyAmino(encodingConfig.Amino)

	return clientCtx
}

// findTxsWithSingleBankSend returns transactions filtered by the provided event.
// It assumes that all the transactions contain only a single bank send, and errors out
// if this is not true. We can start with this assumption and write more complicated type assertion
// if we face errors.
func findTxsWithSingleBankSend(ctx context.Context, clientCtx client.Context, event string) ([]BankSendWithMemo, error) {
	log := logger.Get(ctx)
	tmEvents := []string{event}
	page := 0
	limit := 30
	var response []BankSendWithMemo
	getMore := true
	for getMore {
		page++
		res, err := authtx.QueryTxsByEvents(clientCtx, tmEvents, page, limit, "")
		if err != nil {
			return nil, err
		}
		if page == 1 {
			log.Info("Fetching txs..",
				zap.Uint64("Total Items", res.TotalCount),
				zap.Int("PerPage Items", limit),
			)
		}

		log.Info("Fetching ...", zap.String("page", fmt.Sprintf("%d/%d", page, res.PageTotal)))
		if page == int(res.PageTotal) {
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
			if err != nil {
				return nil, errors.Errorf("can't parse time: %s with format %s", txAny.Timestamp, time.RFC3339)
			}
			response = append(response, BankSendWithMemo{
				Hash:      txAny.TxHash,
				MsgSend:   bankSend,
				Memo:      tx.Body.Memo,
				Timestamp: timestamp,
			})
		}
	}

	return response, nil
}

func writeCoreumTxsToCSV(coreumTxs []BankSendWithMemo, denom, path string) error {
	txs := make([]txExportItem, 0, len(coreumTxs))
	for _, coreumTx := range coreumTxs {
		txs = append(txs, txExportItem{
			Hash:          coreumTx.Hash,
			FromAddress:   coreumTx.FromAddress,
			ToAddress:     coreumTx.ToAddress,
			TargetAddress: coreumTx.ToAddress,
			Amount:        coreumTx.Amount.AmountOf(denom).BigInt(),
			Memo:          coreumTx.Memo,
			Timestamp:     coreumTx.Timestamp,
		})
	}

	return writeTxsToCSV(txs, path)
}
