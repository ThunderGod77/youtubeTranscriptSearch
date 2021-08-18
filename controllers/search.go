package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/olivere/elastic/v7"
	"io/ioutil"
	"log"
	"net/http"
	"rssFeedSearch/global"
)

type query struct {
	Query string `json:"query"`
}

func SearchByContent(w http.ResponseWriter, r *http.Request) {
	//reading the body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		newWebError(w, err, http.StatusInternalServerError)
		return
	}
	var q query

	//parsing the json
	err = json.Unmarshal(body, &q)
	if err != nil {
		newWebError(w, err, http.StatusInternalServerError)
		return
	}

	result, err := search(q.Query)
	if err != nil {
		newWebError(w, err, http.StatusInternalServerError)
		return
	}

	jsonResp, err := json.Marshal(result)
	if err != nil {
		newWebError(w, err, http.StatusInternalServerError)
		return
	}
	sendResp(w, http.StatusOK, jsonResp)
	return

}

func search(q string) ([]elasticDoc, error) {
	mPQ := elastic.NewMatchQuery("text", q).Fuzziness("AUTO")
	searchQuery := global.ElasticClient.Search().Index(indexName).Query(mPQ).From(0).Size(15).Pretty(true)
	searchResp, err := searchQuery.Do(context.Background())
	if err != nil {
		return nil, err
	}
	var result []elasticDoc
	if searchResp.Hits.TotalHits.Value > 0 {
		log.Printf("Found a total of %d results\n", searchResp.Hits.TotalHits.Value)
		for _, hit := range searchResp.Hits.Hits {
			var ed elasticDoc
			err := json.Unmarshal(hit.Source, &ed)
			if err != nil {
				log.Println(err)
				continue
			}
			result = append(result, ed)
		}
	} else {
		return nil, errors.New("no search results found")
	}
	return result, nil

}
