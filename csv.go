package main

import (
	"encoding/csv"
	"io/fs"
	"math/big"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

// WriteAuditTxsToCSV create and writes AuditTx CSV file.
func WriteAuditTxsToCSV(txs []AuditTx, path string) error {
	file, err := createFile(path)
	if err != nil {
		return err
	}

	writer := csv.NewWriter(file)
	defer func() {
		writer.Flush()
		file.Close()
	}()

	// write header
	if err := writer.Write([]string{
		"Hash",
		"FromAddress",
		"ToAddress",
		"Amount",
		"Memo",
		"Timestamp",
	}); err != nil {
		return err
	}

	for _, tx := range txs {
		err := writer.Write([]string{
			tx.Hash,
			tx.FromAddress,
			tx.ToAddress,
			tx.Amount.String(),
			tx.Memo,
			tx.Timestamp.String(),
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// WriteTxsDiscrepancyToCSV create and writes TxDiscrepancy CSV file.
func WriteTxsDiscrepancyToCSV(discrepancies []TxDiscrepancy, path string) error {
	file, err := createFile(path)
	if err != nil {
		return err
	}

	writer := csv.NewWriter(file)
	defer func() {
		writer.Flush()
		file.Close()
	}()

	// write header
	if err := writer.Write([]string{
		"XrplHash",
		"XrplAmount",
		"XrplTargetAddress",
		"XrplMemo",
		"XrplTimestamp",
		"CoreumHash",
		"CoreumAmount",
		"ExpectedAmount",
		"CoreumTargetAddress",
		"CoreumMemo",
		"CoreumTimestamp",
		"BridgingTime",
		"Discrepancy",
	}); err != nil {
		return err
	}

	for _, discrepancy := range discrepancies {
		err := writer.Write([]string{
			discrepancy.XrplTx.Hash,
			discrepancy.XrplTx.Amount.String(),
			discrepancy.XrplTx.TargetAddress,
			discrepancy.XrplTx.Memo,
			discrepancy.XrplTx.Timestamp.String(),
			discrepancy.CoreumTx.Hash,
			discrepancy.CoreumTx.Amount.String(),
			discrepancy.ExpectedAmount.String(),
			discrepancy.CoreumTx.TargetAddress,
			discrepancy.CoreumTx.Memo,
			discrepancy.CoreumTx.Timestamp.String(),
			discrepancy.BridgingTime.String(),
			discrepancy.Discrepancy,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func createFile(path string) (*os.File, error) {
	if err := os.MkdirAll(filepath.Dir(path), fs.ModePerm); err != nil {
		return nil, errors.Errorf("can't create dir, path:%s, err: %s", path, err)
	}

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, fs.ModePerm)
	if err != nil {
		return nil, err
	}
	return file, nil
}

func convertBigIntSliceToString(slice []*big.Int) string {
	stringSlice := make([]string, 0)
	for _, val := range slice {
		stringSlice = append(stringSlice, val.String())
	}

	return strings.Join(stringSlice, ";")
}
