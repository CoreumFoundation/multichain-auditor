package main

import (
	"encoding/csv"
	"io/fs"
	"math/big"
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
)

type txExportItem struct {
	Hash          string
	FromAddress   string
	ToAddress     string
	TargetAddress string
	Amount        *big.Int
	Memo          string
	Timestamp     time.Time
}

func writeTxsToCSV(list []txExportItem, path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(path), fs.ModePerm); err != nil {
			return errors.Errorf("can't create dir, path:%s, err: %s", path, err)
		}
	}

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, fs.ModePerm)
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
		"TargetAddress",
		"Amount",
		"Memo",
		"Timestamp",
	}); err != nil {
		return err
	}

	for _, elem := range list {
		err := writer.Write([]string{
			elem.Hash,
			elem.FromAddress,
			elem.ToAddress,
			elem.TargetAddress,
			elem.Amount.String(),
			elem.Memo,
			elem.Timestamp.String(),
		})
		if err != nil {
			return err
		}
	}
	return nil
}
