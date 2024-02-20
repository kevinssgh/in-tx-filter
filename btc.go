package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
)

var btcCmd = &cobra.Command{
	Use:   "btc",
	Short: "Filter inbound btc deposits",
	Run:   FilterBTCTransactions,
}

func init() {
	rootCmd.AddCommand(btcCmd)
}

func FilterBTCTransactions(_ *cobra.Command, _ []string) {
	list := getHashList()
	CheckForCCTX(list)
}

func getHashList() []deposit {
	var list []deposit
	lastHash := ""

	url := "https://blockstream.info/api/address/bc1qm24wp577nk8aacckv8np465z3dvmu7ry45el6y/txs"

	for {
		nextQuery := url
		if lastHash != "" {
			path := fmt.Sprintf("/chain/%s", lastHash)
			nextQuery = url + path
		}
		res, getErr := http.Get(nextQuery)
		if getErr != nil {
			log.Fatal(getErr)
		}

		body, readErr := ioutil.ReadAll(res.Body)
		if readErr != nil {
			log.Fatal(readErr)
		}

		var txns []map[string]interface{}
		err := json.Unmarshal(body, &txns)
		if err != nil {
			fmt.Println("error unmarshalling: ", err.Error())
		}

		if len(txns) == 0 {
			break
		}

		fmt.Println("Length of txns: ", len(txns))

		for _, txn := range txns {
			hash := txn["txid"].(string)

			vout := txn["vout"].([]interface{})
			vout0 := vout[0].(map[string]interface{})
			var vout1 map[string]interface{}
			if len(vout) > 1 {
				vout1 = vout[1].(map[string]interface{})
			} else {
				continue
			}
			_, found := vout0["scriptpubkey"]
			scriptpubkey := ""
			if found {
				scriptpubkey = vout0["scriptpubkey"].(string)
			}
			_, found = vout0["scriptpubkey_address"]
			targetAddr := ""
			if found {
				targetAddr = vout0["scriptpubkey_address"].(string)
			}

			//Check if txn is confirmed
			status := txn["status"].(map[string]interface{})
			confirmed := status["confirmed"].(bool)
			if !confirmed {
				continue
			}

			//Filter out deposits less than min base fee
			if vout0["value"].(float64) < 1360 {
				continue
			}

			//Check if deposit is a donation
			scriptpubkey1 := vout1["scriptpubkey"].(string)
			if len(scriptpubkey1) >= 4 && scriptpubkey1[:2] == "6a" {
				memoSize, err := strconv.ParseInt(scriptpubkey1[2:4], 16, 32)
				if err != nil {
					continue
				}
				if int(memoSize) != (len(scriptpubkey1)-4)/2 {
					continue
				}
				memoBytes, err := hex.DecodeString(scriptpubkey1[4:])
				if err != nil {
					continue
				}
				if bytes.Equal(memoBytes, []byte(DonationMessage)) {
					continue
				}
			} else {
				continue
			}

			//Make sure deposit is sent to correct tss address
			if strings.Compare("0014", scriptpubkey[:4]) == 0 && targetAddr == "bc1qm24wp577nk8aacckv8np465z3dvmu7ry45el6y" {
				entry := deposit{
					hash,
					vout0["value"].(float64),
				}
				list = append(list, entry)
			}
		}

		lastTxn := txns[len(txns)-1]
		lastHash = lastTxn["txid"].(string)
		//fmt.Println("last hash: ", lastHash)
	}

	return list
}
