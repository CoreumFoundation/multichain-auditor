package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/pkg/errors"

	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
)

const (
	multichainRescanRequestTimeout   = 10 * time.Second
	multichainRescanResStatusSuccess = "Success"
	multichainRescanRetries          = 10
	multichainRescanRetryTimout      = time.Second
)

type multichainRescanResp struct {
	Msg   string `json:"msg"`
	Error string `json:"error"`
}

func RescanMultichainTxs(ctx context.Context, baseURL string, txHashes []string) error {
	log := logger.Get(ctx)
	log.Info(fmt.Sprintf("Rescanning %d xrpl txs", len(txHashes)))
	for _, txHash := range txHashes {
		log.Info(fmt.Sprintf("Rescanning %q xrpl tx", txHash))
		err := rescanMultichainTxWithRetry(ctx, baseURL, txHash)
		if err != nil {
			return err
		}
	}

	return nil
}

func rescanMultichainTxWithRetry(ctx context.Context, baseURL, txHash string) error {
	url := fmt.Sprintf("%s/v2/reswaptxns?hash=%s&srcChainID=XRP&destChainID=ATOM_DCORE", baseURL, txHash)

	var (
		resBody multichainRescanResp
		err     error
	)
	for i := 0; i < multichainRescanRetries; i++ {
		reqCtx, reqCtxCancel := context.WithTimeout(ctx, multichainRescanRequestTimeout)
		err = DoJSON(reqCtx, http.MethodGet, url, nil, &resBody)
		reqCtxCancel()
		if err == nil && resBody.Msg == multichainRescanResStatusSuccess {
			return nil
		}
		if i != multichainRescanRetries-1 {
			<-time.After(multichainRescanRetryTimout)
		}
	}

	return errors.Errorf("can't send %s tx to rescan with %d retries and timeout %s, last response: %v, last err: %s",
		txHash, multichainRescanRetries, multichainRescanRetryTimout, resBody, err)
}
