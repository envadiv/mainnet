package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/envadiv/Passage3D/app"
)

type Record struct {
	Address                sdk.AccAddress
	TotalAmount            Dec
	StartTime              time.Time
	NumWeeklyDistributions int
}

const (
	// Based on 365.2425 days/year
	SecondsPerYear  = 31556952
	SecondsPerMonth = 2629746
	SecondsPerWeek  = 604800
)

var (
	OneYear, OneMonth, OneWeek time.Duration
)

func init() {
	var err error

	OneYear, err = time.ParseDuration(fmt.Sprintf("%ds", SecondsPerYear))
	if err != nil {
		panic(err)
	}

	OneMonth, err = time.ParseDuration(fmt.Sprintf("%ds", SecondsPerMonth))
	if err != nil {
		panic(err)
	}

	OneWeek, err = time.ParseDuration(fmt.Sprintf("%ds", SecondsPerWeek))
	if err != nil {
		panic(err)
	}
}

func ParseAccountsCsv(rdr io.Reader, genesisTime time.Time, errorsAsWarnings bool) ([]Record, error) {
	csvRdr := csv.NewReader(rdr)
	lines, err := csvRdr.ReadAll()
	if err != nil {
		return nil, err
	}

	records := make([]Record, 0, len(lines))
	for i, line := range lines {
		record, err := parseLine(line, genesisTime)
		if err != nil {
			if errorsAsWarnings {
				fmt.Printf("WARNING: Error on line %d: %v\n", i, err)
				continue
			} else {
				return nil, err
			}
		}

		records = append(records, record)
	}

	return records, nil
}

func parseLine(line []string, genesisTime time.Time) (Record, error) {
	addr, err := sdk.GetFromBech32(line[0], app.Bech32PrefixAccAddr)
	if err != nil {
		fmt.Println("bech32 error", line[0])
		return Record{}, err
	}

	amount, err := NewDecFromString(line[1])
	if err != nil {
		return Record{}, err
	}

	var startTime time.Time
	startTimeStr := strings.TrimSpace(line[2])
	switch startTimeStr {
	case "0":
	case "MAINNET+1YEAR":
		startTime = genesisTime.Add(OneYear)
	default:
		numWeeks, err := strconv.Atoi(startTimeStr)
		if err != nil {
			return Record{}, err
		}

		fmt.Println("numWeeks", numWeeks)

		weeksInSeconds, err := time.ParseDuration(fmt.Sprintf("%ds", SecondsPerWeek*numWeeks))
		if err != nil {
			return Record{}, err
		}

		startTime = genesisTime.Add(weeksInSeconds)
	}

	numDist, err := strconv.Atoi(line[3])
	if err != nil {
		return Record{}, err
	}

	// if numDist < 1 {
	// 	return Record{}, fmt.Errorf("expected a positive integer, got %d", numDist)
	// }

	return Record{
		Address:                addr,
		TotalAmount:            amount,
		StartTime:              startTime,
		NumWeeklyDistributions: numDist,
	}, nil

}

func (r Record) Equal(o Record) bool {
	return r.StartTime.Equal(o.StartTime) &&
		r.TotalAmount.IsEqual(o.TotalAmount) &&
		r.Address.Equals(o.Address) &&
		r.NumWeeklyDistributions == o.NumWeeklyDistributions
}
