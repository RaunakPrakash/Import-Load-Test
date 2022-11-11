package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	sendinblue "github.com/sendinblue/APIv3-go-library/v2/lib"
)

type resultError struct {
	pID        int64
	statusCode int
	err        error
}

var count int64

func main() {
	ctx := context.Background()
	cfg := sendinblue.NewConfiguration()
	apiKye := make([]string, 0)
	// apikey will be around 100 (100 clients)
	apiKye = append(apiKye, "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")
	filesURl := make([]string, 0)
	// files can be of size 6k to 2M
	filesURl = append(filesURl, "https://raw.githubusercontent.com/amit0592/testFiles/main/1-1lakh.csv", "https://raw.githubusercontent.com/amit0592/testFiles/main/test3.csv", "https://raw.githubusercontent.com/amit0592/testFiles/main/test2.csv")
	maxImportLimit := 246
	startTime := time.Now()
	result := make(chan resultError, maxImportLimit)
	for j := 0; j < maxImportLimit/(len(apiKye)*len(filesURl)); j++ {
		for i, api := range apiKye {
			cfg.AddDefaultHeader("api-key", api)
			sib := sendinblue.NewAPIClient(cfg)
			go makeHTTPRequest(ctx, sib, filesURl, result, i)
		}
	}
	for i := 0; i < maxImportLimit; i++ {
		r := <-result
		fmt.Println(r)
	}
	endTime := time.Now()
	log.Println(int64(endTime.Sub(startTime) / time.Second))
	close(result)
	log.Println(count)
}

func makeHTTPRequest(ctx context.Context, sib *sendinblue.APIClient, filesURL []string, result chan<- resultError, w int) {
	var m sync.Mutex
	for _, url := range filesURL {
		fmt.Println(w)
		data := sendinblue.RequestContactImport{
			FileUrl:                 url,
			ListIds:                 []int64{2},
			EmailBlacklist:          false,
			SmsBlacklist:            false,
			UpdateExistingContacts:  true,
			EmptyContactsAttributes: false,
		}
		if ctx == nil {
			log.Println("time out")
			return
		}
		contacts, h, err := sib.ContactsApi.ImportContacts(ctx, data)
		r := resultError{
			pID:        contacts.ProcessId,
			statusCode: h.StatusCode,
			err:        err,
		}
		result <- r
		m.Lock()
		{
			count++
		}
		m.Unlock()
	}
	log.Println("done---------")
}
