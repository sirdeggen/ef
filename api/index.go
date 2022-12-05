package handler

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/GorillaPool/go-junglebus"
	"github.com/GorillaPool/go-junglebus/models"
	"github.com/libsv/go-bt/v2"
)

type EfResponse struct {
	Ef    string `json:"ef,omitempty"`
	Error string `json:"error,omitempty"`
}

type EfRequest struct {
	RawTx string `json:"rawtx,omitempty"`
}

func doResponse(w http.ResponseWriter, response EfResponse) {
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Printf("error encoding json: %s", err.Error())
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, "error encoding json: %s", err.Error())
	}
}

func Handler(w http.ResponseWriter, r *http.Request) {
	var response EfResponse

	// json api
	w.Header().Set("Content-Type", "application/json")

	// get the tx inputs from request
	var req EfRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		doResponse(w, EfResponse{Error: err.Error()})
		return
	}

	// go lookup the input transactions on JungleBus
	junglebusClient, err := junglebus.New(
		junglebus.WithHTTP("https://junglebus.gorillapool.io"),
	)
	if err != nil {
		log.Fatalln(err.Error())
	}

	transaction, err := bt.NewTxFromString(req.RawTx)
	if err != nil {
		log.Fatalln(err.Error())
	}

	for _, input := range transaction.Inputs {
		txid := input.PreviousTxIDStr()
		log.Printf("Looking up transaction %s", txid)
		var tx *models.Transaction
		if tx, err = junglebusClient.GetTransaction(context.Background(), txid); err != nil {
			log.Printf("ERROR: failed getting transaction %s\n", err.Error())
		} else {
			actualTx, err := bt.NewTxFromBytes(tx.Transaction)
			if err != nil {
				log.Println(err.Error())
			}
			o := actualTx.Outputs[input.PreviousTxOutIndex]
			log.Println(o.LockingScript.String())
			input.PreviousTxScript = o.LockingScript
			input.PreviousTxSatoshis = o.Satoshis
		}
	}

	response.Ef = hex.EncodeToString(transaction.ExtendedBytes())

	// respond with the extended transaction
	doResponse(w, response)
}
