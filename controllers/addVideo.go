package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/olivere/elastic/v7"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"rssFeedSearch/global"
	"strconv"
	"strings"
)

var indexName = "podcast"

type elasticDoc struct {
	Start   int    `json:"start"`
	End     int    `json:"end"`
	Text    string `json:"text"`
	Title   string `json:"title"`
	Channel string `json:"channel"`
	Url     string `json:"url"`
	Id      string `json:"id"`
}

type vid struct {
	Url string `json:"url"`
}

type browserResult struct {
	TimeStamps []string `json:"timeStamps"`
	Caption    []string `json:"caption"`
	VidTitle   string   `json:"vidTitle"`
	VidChannel string   `json:"vidChannel"`
}
type browserResp struct {
	Result browserResult `json:"result"`
	Err    string        `json:"err"`
}
type aResp struct {
	Err bool   `json:"err"`
	Id  string `json:"id"`
}

type combine struct {
	Start   int    `json:"start"`
	End     int    `json:"end"`
	Caption string `json:"caption"`
}

func AddVideos(w http.ResponseWriter, r *http.Request) {
	//reading the body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		newWebError(w, err, http.StatusInternalServerError)
		return
	}
	var v vid

	//parsing the json
	err = json.Unmarshal(body, &v)
	if err != nil {
		newWebError(w, err, http.StatusInternalServerError)
		return
	}

	//to validate if the url is a youtube video url
	val := validateURL(v.Url)
	if !val {
		newWebError(w, errors.New("invalid youtube url"), http.StatusBadRequest)
		return
	}

	//to get unique id for the process
	id := uuid.New().String()

	ar := aResp{
		Err: false,
		Id:  id,
	}
	respAr, err := json.Marshal(ar)
	if err != nil {
		newWebError(w, err, http.StatusInternalServerError)
		return
	}
	//to get the transcript of the video and add it to elastic search
	go getTranscript(v.Url, id)

	//sending response that the request has been accepted and is processing
	sendResp(w, http.StatusAccepted, respAr)
	return

}

func validateURL(url string) (ok bool) {
	match, _ := regexp.MatchString("^(http(s)?:\\/\\/)?((w){3}.)?youtu(be|.be)?(\\.com)?\\/.+", url)
	return match
}

func getTranscript(url, id string) {
	//registering the process in redis so user can see the status of the process
	err := registerProcess(id, url)
	if err != nil {
		log.Println(err)
		updateStatus(id, "Error", err.Error())
		return
	}
	//making the request to the node.js server to get the transcript of the video
	client := &http.Client{}
	var jsonStr = []byte(fmt.Sprintf("{\"url\":\"%s\"}", url))

	req, err := http.NewRequest("POST", "http://localhost:5000", bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		updateStatus(id, "Error", err.Error())
		return
	}
	defer resp.Body.Close()

	var br browserResp
	//reading the response from the node.js server
	body, err := ioutil.ReadAll(resp.Body)

	err = json.Unmarshal(body, &br)
	if err != nil {
		log.Println(err)
		updateStatus(id, "Error", err.Error())
		return
	}
	if br.Err != "" {
		log.Println(br.Err)
		updateStatus(id, "Error", br.Err)
		return
	}

	//processing the timestamps - converting them from strings to integers
	pTs, err := processTimeStamps(br.Result.TimeStamps)

	if err != nil {
		log.Println(err)
		updateStatus(id, "Error", err.Error())
		return
	}
	//processing captions  - removing escape characters and trimming space
	pC, err := processCaption(br.Result.Caption)
	if err != nil {
		log.Println(err)
		updateStatus(id, "Error", err.Error())
		return
	}

	//merging timestamps and captions into different sections of 300 seconds
	m, err := merging(pC, pTs)
	if err != nil {
		log.Println(err)
		updateStatus(id, "Error", err.Error())
		return
	}
	//updating status to indicate that data has been fetched
	updateStatus(id, "Obtained Data!", "")
	//indexing data to elastic search
	go addToElasticSearch(m, br.Result.VidTitle, br.Result.VidChannel, url, id)

}

func addToElasticSearch(m []combine, title, creator, url, id string) {
	bulk := global.ElasticClient.Bulk()




	for _, data := range m {

		eD := elasticDoc{
			Start:   data.Start,
			End:     data.End,
			Text:    data.Caption,
			Title:   title,
			Channel: creator,
			Url:     url,
			Id:      id,
		}
		req := elastic.NewBulkIndexRequest()
		req.OpType("index")
		req.Index(indexName)
		req.Doc(eD)
		bulk = bulk.Add(req)
	}
	bulkResp, err := bulk.Do(context.Background())
	if err != nil {
		log.Println(err)
		updateStatus(id, "Error", err.Error())
		return
	}
	indexed := bulkResp.Indexed()
	for _, info := range indexed {
		log.Println("nBulk response Index:", info)
	}
	updateStatus(id, "Indexed Successfully", "")
	return

}

func updateStatus(id, status, msg string) {
	conn := global.ConnPool.Get()
	defer conn.Close()

	_, err := conn.Do("HMSET", id, "status", status, "msg", msg)
	if err != nil {
		log.Println(err)
	}
}

func registerProcess(id, url string) error {
	conn := global.ConnPool.Get()
	defer conn.Close()
	_, err := conn.Do("HMSET", id, "status", "Started Processing", "url", url, "msg", "")
	return err
}

func processTimeStamps(t []string) ([]int, error) {
	//regex to remove special characters like \n
	var pT []int

	for _, ts := range t {
		//to replace \n with blank space
		pt := strings.ReplaceAll(ts, "\n", "")
		pt = strings.ReplaceAll(ts, "\r", "")
		//to trim and remove whitespace
		pt = strings.Trim(pt, " ")

		pt = strings.TrimSuffix(pt, "\n")
		pt = strings.TrimPrefix(pt, "\n")

		time := strings.Split(pt, ":")
		minS := strings.Trim(time[0]," ")
		//to get time passed in seconds from string timestamp
		min, err := strconv.Atoi(minS)

		sec, err := strconv.Atoi(time[1])

		if err != nil {
			return nil, err
		}
		t := min*60 + sec
		pT = append(pT, t)
	}
	return pT, nil

}

func processCaption(c []string) ([]string, error) {
	//regex to remove special characters like \n
	re, err := regexp.Compile("/\\r?\\n|\\r/g")
	if err != nil {
		return nil, err
	}
	var captions []string

	for _, ts := range c {
		//to replace \n with blank space
		pt := re.ReplaceAllString(ts, "")
		//to trim and remove whitespace
		pt = strings.Trim(pt, " ")
		pt = strings.TrimPrefix(pt, "\n")
		pt = strings.TrimSuffix(pt, "\n")
		pt = strings.Trim(pt, " ")
		pt = pt + " "
		captions = append(captions, pt)
	}
	return captions, nil

}

func merging(c []string, t []int) ([]combine, error) {
	if len(c) != len(t) {
		return nil, errors.New("length of caption and timestamps not same")
	}
	tl := 60
	start := 0
	cpt := ""
	sAdd := true
	var result []combine
	for index, val := range t {

		if val > tl {

			result = append(result, combine{
				Start:   start,
				End:     tl,
				Caption: cpt,
			})
			start = tl
			tl = tl + 60
			cpt = ""
			sAdd = false
		}
		cpt = cpt + c[index]
		sAdd = true

	}
	if  sAdd {
		result = append(result, combine{
			Start:   start,
			End:     tl,
			Caption: cpt,
		})
	}

	return result, nil
}
