package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

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
				zap.Uint64("Total Page", res.PageTotal),
			)
		}

		log.Info("Fetching ...", zap.String("page", fmt.Sprintf("%d/%d", page, res.PageTotal)))
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
			response = append(response, BankSendWithMemo{
				MsgSend:   bankSend,
				Memo:      tx.Body.Memo,
				Timestamp: txAny.Timestamp,
			})
		}
	}

	return response, nil
}

type BankSendWithMemo struct {
	*banktypes.MsgSend
	Memo      string
	Timestamp string
}

func ensureDir(dirPath string, perm fs.FileMode) {
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		os.MkdirAll(dirPath, fs.FileMode(perm))
	}
}

func writeCoreumTxsToCSV(list []BankSendWithMemo, denom string, path string) error {
	permission := fs.FileMode(0777)
	ensureDir(filepath.Dir(path), permission)

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, permission)
	if err != nil {
		return errors.WithStack(err)
	}

	writer := csv.NewWriter(file)
	defer func() {
		writer.Flush()
		file.Close()
	}()

	// write header
	if err := writer.Write([]string{
		"FromAddress",
		"ToAddress",
		"Amount",
		"Memo",
		"Timestamp",
	}); err != nil {
		return errors.WithStack(err)
	}

	for _, elem := range list {
		err := writer.Write([]string{
			elem.FromAddress,
			elem.ToAddress,
			elem.Amount.AmountOf(denom).String(),
			elem.Memo,
			elem.Timestamp,
		})
		if err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}
