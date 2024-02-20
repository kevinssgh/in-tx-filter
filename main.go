package main

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

var rootCmd = &cobra.Command{
	Use:   "filterdeposit",
	Short: "filter missing inbound deposits",
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

type deposit struct {
	txId   string
	amount float64
}

func CheckForCCTX(list []deposit) {
	zetaUrl := "http://46.4.15.110:1317/zeta-chain/crosschain/in_tx_hash_to_cctx_data/"
	var missedList []deposit

	fmt.Println("Going through list, num of transactions: ", len(list))
	for _, entry := range list {
		url := zetaUrl + entry.txId
		res, getErr := http.Get(url)
		if getErr != nil {
			log.Fatal(getErr)
		}

		data, readErr := ioutil.ReadAll(res.Body)
		if readErr != nil {
			log.Fatal(readErr)
		}

		var cctx map[string]interface{}
		err := json.Unmarshal(data, &cctx)
		if err != nil {
			fmt.Println("error unmarshalling: ", err.Error())
		}

		if _, ok := cctx["code"]; ok {
			missedList = append(missedList, entry)
			//fmt.Println("appending to missed list: ", entry)
		}
	}

	for _, entry := range missedList {
		fmt.Printf("%s, amount: %d\n", entry.txId, int64(entry.amount))
	}
}
