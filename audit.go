package main

import (
	"fmt"
	"io"
	"sort"
	"time"
)

func SortAccounts(accMap map[string]Account) []Account {
	var addrs []string

	for addr := range accMap {
		addrs = append(addrs, addr)
	}

	sort.Strings(addrs)

	res := make([]Account, 0, len(accMap))
	for _, addr := range addrs {
		res = append(res, accMap[addr])
	}

	return res
}

func PrintAccountAudit(accounts []Account, genesisTime time.Time, writer io.Writer) {
	for _, acc := range accounts {
		_, err := fmt.Fprintf(writer, "%s\t%s\t%d\n", acc.Address, acc.TotalPassage, len(acc.Distributions))
		if err != nil {
			panic(err)
		}
		for _, dist := range acc.Distributions {
			var timeStr string
			if dist.Time.Equal(genesisTime) {
				timeStr = "MAINNET"
			} else {
				timeStr = dist.Time.Format("2006-01-02 15:04:05")
			}
			_, err = fmt.Fprintf(writer, "\t%s\t%s\n", dist.Passage, timeStr)
			if err != nil {
				panic(err)
			}
		}
	}
}
