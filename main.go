package main

import (
	"context"
	"log"
	"runtime"
	"sync"
	"time"

	sendinblue "github.com/sendinblue/APIv3-go-library/v2/lib"
)

type resultError struct {
	api        string
	pID        int64
	statusCode int
	err        error
}

var count int64

func main() {
	ctx := context.Background()
	apiKye := make([]string, 0)
	// apikey will be around 100 (100 clients)
	apiKye = append(apiKye, "")
	filesURl := make([]string, 0)
	// files can be of size 6k to 2M
	filesURl = append(filesURl, "https://raw.githubusercontent.com/amitsendinblue/Hello/master/ImportLoadTestFiles/20Lakh.csv", "https://raw.githubusercontent.com/amitsendinblue/Hello/master/ImportLoadTestFiles/10Lakh.csv", "https://raw.githubusercontent.com/amitsendinblue/Hello/master/ImportLoadTestFiles/5Lakh.csv", "https://raw.githubusercontent.com/amitsendinblue/Hello/master/ImportLoadTestFiles/1Lakh.csv", "https://raw.githubusercontent.com/amitsendinblue/Hello/master/ImportLoadTestFiles/50000.csv")
	maxImportLimit := 15000
	startTime := time.Now()
	result := make(chan resultError, maxImportLimit)
	max := runtime.GOMAXPROCS(0)
	g := make(chan bool, max)
	log.Println(len(apiKye), len(filesURl), maxImportLimit/(len(apiKye)*len(filesURl)))
	for j := 0; j < maxImportLimit/(len(apiKye)*len(filesURl)); j++ {
		for i, api := range apiKye {
			go makeHTTPRequest(ctx, filesURl, result, i, api, g)
		}
	}
	for i := 0; i < maxImportLimit; i++ {
		r := <-result
		log.Println(r)
	}
	endTime := time.Now()
	log.Println(int64(endTime.Sub(startTime) / time.Second))
	close(result)
	log.Println(count)
}

func makeHTTPRequest(ctx context.Context, filesURL []string, result chan<- resultError, w int, api string, l chan bool) {
	var m sync.Mutex
	cfg := sendinblue.NewConfiguration()
	cfg.AddDefaultHeader("api-key", api)
	sib := sendinblue.NewAPIClient(cfg)
	l <- true
	for _, url := range filesURL {
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
			api:        api,
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
	<-l
}
