package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"

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

			fmt.Println(input_file, output_file, prefix)

			if err := AddressConvert(input_file, output_file, prefix); err != nil {
				return err
			}
			fmt.Println(fmt.Sprintf("Successfully prefix %s address conversion is done.", prefix))
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
			ClaimAmount:      airdropRecord.AirdropAmount,
			Passage3DAddress: airdropRecord.NewPrefixAddress,
		}
		airdropAccountRecords = append(airdropAccountRecords, re)
	}

	writeToCsvFile(output_file, airdropAccountRecords)
	return nil
}

type AirdropRecord struct {
	OldAddress       string
	NewPrefixAddress string
	AirdropAmount    string
}

type Passage3DAirdropClaimRecord struct {
	Passage3DAddress string
	ClaimAmount      string
}

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

func parseRecords(prefix string, records [][]string) ([]AirdropRecord, error) {
	var airdropAccounts []AirdropRecord
	var count = 0
	for _, record := range records {
		// bech32 decode
		_, hrf, err := bech32.Decode(record[0])
		if err != nil {
			return nil, err
		}
		aidropAmount := record[1]
		newAccAddr, err := bech32.Encode(prefix, hrf)
		if err != nil {
			return nil, err
		}
		// encode with prefix
		airdropAccounts = append(airdropAccounts, AirdropRecord{
			OldAddress:       record[0],
			NewPrefixAddress: newAccAddr,
			AirdropAmount:    aidropAmount,
		})

		count++
	}

	fmt.Println("total converted", count)
	return airdropAccounts, nil
}
