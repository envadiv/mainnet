package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/btcsuite/btcutil/bech32"
	"github.com/spf13/cobra"
)

func AddressConverter() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "addr-converter [input-file] [output-file] [prefix]",
		Short: "Converting the bech32 address into prefix account address.",
		Args:  cobra.ExactArgs(3),
		Example: `
		go run . addr-converter input.csv output_address.csv pasg
		`,
		RunE: func(cmd *cobra.Command, args []string) error {
			input_file := args[0]
			output_file := args[1]
			prefix := args[2]
			if err := AddressConvert(input_file, output_file, prefix); err != nil {
				return err
			}
			_, _ = fmt.Println(fmt.Sprintf("Successfully prefix %s address conversion is done.", prefix))
			return nil
		},
	}

	return cmd
}

func AddressConvert(input_file, output_file, prefix string) error {
	var data [][]string
	data, err := readCsvFile(input_file)
	if err != nil {
		panic(err)
	}
	airdropRecords, err := parseRecords(prefix, data[1:])
	if err != nil {
		panic(err)
	}

	var airdropAccountRecords []Passage3DAirdropClaimRecord
	for _, airdropRecord := range airdropRecords {
		re := Passage3DAirdropClaimRecord{
			ClaimAmount:      strconv.Itoa(10 * 10e6),
			Passage3DAddress: airdropRecord,
		}
		airdropAccountRecords = append(airdropAccountRecords, re)
	}

	writeToCsvFile(output_file, airdropAccountRecords)
	return nil
}

type AirdropRecord struct {
	Name             string
	Passage3DAddress string
	CosmosAddress    string
}

type Passage3DAirdropClaimRecord struct {
	Passage3DAddress string
	ClaimAmount      string
}

var (
	Passage3dPrefix = "pasg"
)

func writeToCsvFile(output_file string, claimRecords []Passage3DAirdropClaimRecord) {

	csvFile, err := os.Create(output_file)

	if err != nil {
		log.Fatalf("failed creating file: %s", err)
	}

	csvwriter := csv.NewWriter(csvFile)

	for _, empRow := range claimRecords {
		_ = csvwriter.Write([]string{empRow.Passage3DAddress, empRow.ClaimAmount})
	}
	csvwriter.Flush()
	csvFile.Close()
}

func parseRecords(prefix string, records [][]string) ([]string, error) {
	var airdropAccounts []string
	for _, record := range records {
		// bech32 decode
		_, hrf, err := bech32.Decode(record[1])
		if err != nil {
			return nil, nil
		}
		p3dAccAddr, err := bech32.Encode(Passage3dPrefix, hrf)
		if err != nil {
			return nil, err
		}
		// encode with prefix
		airdropAccounts = append(airdropAccounts, p3dAccAddr)
	}
	return airdropAccounts, nil
}
