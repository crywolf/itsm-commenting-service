//go:build migrationTool
// +build migrationTool

package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/KompiTech/cgi-tools/pkg/client"
	"github.com/KompiTech/cgi-tools/pkg/spacedata"
	"github.com/KompiTech/cgi-tools/pkg/stuff"
	"github.com/KompiTech/rmap"
)

const (
	endpointEnv    = "ENDPOINT"
	tokenEnv       = "TOKEN"
	spaceIDJsonEnv = "SPACEIDJSON"
)

func main() {
	spaceIDJson := os.Getenv(spaceIDJsonEnv)
	spaceData, spaceIDs, err := spacedata.Init(spaceIDJson, true)
	if err != nil {
		log.Fatal(err)
	}
	log.Print("Migration started")

	transport := &http.Transport{
		//TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		Dial: (&net.Dialer{
			Timeout:   0,
			KeepAlive: 0,
		}).Dial,
		TLSHandshakeTimeout: 10 * time.Second,
	}
	sem := make(chan bool, 20)
	for _, spaceID := range spaceIDs {
		log.Println(spaceData.FindNameByUUID(spaceID))

		cli := client.New(stuff.MandatoryEnv(endpointEnv), spaceID, func() string {
			return stuff.MandatoryEnv(tokenEnv)
		}, transport)

		request := rmap.NewFromMap(map[string]interface{}{
			"selector": map[string]interface{}{
				"created_at": map[string]interface{}{
					"$gt": "2021-11-08T00:00:00Z",
				},
			},
		})

		//data, err := cli.GetAllData("/assets/comment?resolve=true", request)
		//if err != nil {
		//	log.Fatal(err)
		//}

		apiPath := "/assets/comment?resolve=true"
		//apiPath := "/assets/work_note?resolve=true"

		log.Println("init couchDB")
		cdb, err := NewCouchDBClientFromEnv()
		if err != nil {
			log.Fatal(err)
		}

		cdb.DBName = fmt.Sprintf("p_%s_comments", spaceID)

		log.Println("init couchDB done")

		bookmark := ""

		for {
			if request.Mapa["bookmark"] == bookmark {
				break
			}
			request.Mapa["bookmark"] = bookmark
			resp, err := rmap.NewFromBytes(cli.MakeBodyRequest(apiPath, "GET", request.Bytes()))
			if err != nil {
				log.Fatalf("GetAllData rmap.NewFromBytes() failed: %s", err)
			}

			results := resp.MustGetIterable("result")
			log.Println(len(results))

			for _, itemI := range results {
				sem <- true
				go func(itemI interface{}) {
					defer func() { <-sem }()
					item, err := rmap.NewFromInterface(itemI)
					if err != nil {
						log.Fatalf("GetAllData rmap.NewFromInterface() failed: %s", err)
					}

					// GetReadByListDetails
					docType := item.MustGetString("docType")
					uuid := item.MustGetString("uuid")
					readByList, err := getReadByListDetails(cli, fmt.Sprintf("%s:%s", strings.ToUpper(docType), uuid))
					if err != nil {
						log.Fatal(err)
					}

					transItem, err := transformComment(item, readByList)
					if err != nil {
						log.Fatal(err)
					}

					log.Println(transItem)
					// store it into couchDB
					if err := cdb.InsertIfNotExist(uuid, transItem.String()); err != nil {
						log.Fatal(err)
					}
				}(itemI)

			}

			if len(results) < 10 {
				break
			}

			bookmark = resp.MustGetString("bookmark")
		}
	}
	for i := 0; i < cap(sem); i++ {
		sem <- true
	}
}

func getReadByListDetails(client2 *client.Client, entity string) ([]interface{}, error) {
	payload := []byte(`{"entity": "` + entity + `"}`)
	respBytes, err := client2.MakeBodyRequestError("/functions-query/getReadByList", http.MethodOptions, payload)
	if err != nil {
		return nil, err
	}

	resp, err := rmap.NewFromBytes(respBytes)
	if err != nil {
		return nil, err
	}

	results, err := resp.GetJPtrIterable("/result/result")
	if err != nil {
		return nil, err
	}

	return results, nil
}

func transformComment(originalComment rmap.Rmap, originalReadByList []interface{}) (rmap.Rmap, error) {
	uuid := originalComment.MustGetString("uuid")
	entity := originalComment.MustGetString("entity")
	text := originalComment.MustGetString("text")
	createdAt := originalComment.MustGetString("created_at")
	createdByName := originalComment.MustGetJPtrString("/created_by/private/name")
	createdBySurname := originalComment.MustGetJPtrString("/created_by/private/surname")
	createdByUUID := originalComment.MustGetJPtrString("/created_by/uuid")
	createdByOrg := originalComment.MustGetJPtrString("/created_by/org_name")
	createdByOrgDisplay := originalComment.MustGetJPtrString("/created_by/org_display_name")
	comment := rmap.MustNewFromBytes([]byte(`
		{
		  "_id": "` + uuid + `",
		  "uuid": "` + uuid + `",
		  "entity": "` + entity + `",
		  "created_at": "` + createdAt + `",
		  "created_by": {
			"uuid": "` + createdByUUID + `",
			"name": "` + createdByName + `",
			"surname": "` + createdBySurname + `",
			"org_display_name": "` + createdByOrgDisplay + `",
			"org_name": "` + createdByOrg + `"
		  }
		}`))
	comment.Mapa["text"] = text

	readByList, err := transformReadByList(originalReadByList)
	if err != nil {
		return rmap.Rmap{}, err
	}

	comment.Mapa["read_by"] = readByList

	return comment, nil
}

func transformReadByList(originalList []interface{}) ([]rmap.Rmap, error) {
	output := []rmap.Rmap{}
	for _, itemI := range originalList {
		item, err := rmap.NewFromInterface(itemI)
		if err != nil {
			return nil, err
		}
		log.Println(item)
		name := item.MustGetJPtrString("/name")
		surname := item.MustGetJPtrString("/surname")
		createdByUUID := item.MustGetJPtrString("/uuid")
		org := item.MustGetJPtrString("/org_name")
		orgDisplay := item.MustGetJPtrString("/org_display_name")
		timeValue := item.MustGetJPtrString("/time")

		readBy := rmap.MustNewFromBytes([]byte(`{
			"time": "` + timeValue + `",
			"user":{"uuid": "` + createdByUUID + `",
			"name": "` + name + `",
			"surname": "` + surname + `",
			"org_display_name": "` + orgDisplay + `",
			"org_name": "` + org + `"}}`))

		output = append(output, readBy)
	}

	return output, nil
}
