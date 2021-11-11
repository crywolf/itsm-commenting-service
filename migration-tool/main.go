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
	"github.com/pkg/errors"
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

	for _, spaceID := range spaceIDs {
		log.Println(spaceData.FindNameByUUID(spaceID))

		cli := client.New(stuff.MandatoryEnv(endpointEnv), spaceID, func() string {
			return stuff.MandatoryEnv(tokenEnv)
		}, transport)

		request := rmap.NewFromMap(map[string]interface{}{
			"selector": map[string]interface{}{
				"work_notes": map[string]interface{}{"$exists": true},
				"created_at": map[string]interface{}{
					"$gt": "2021-06-08T00:00:00Z",
				},
			},
		})

		//data, err := cli.GetAllData("/assets/comment?resolve=true", request)
		//if err != nil {
		//	log.Fatal(err)
		//}

		//apiPath := "/assets/comment?resolve=true"
		apiPath := "/assets/k_request"

		log.Println("init couchDB")
		cdb, err := NewCouchDBClientFromEnv()
		if err != nil {
			log.Fatal(err)
		}

		cdb.DBName = fmt.Sprintf("p_%s_worknotes", spaceID)

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
			// TODO concurrent
			sem := make(chan bool, 20)
			for _, itemI := range results {
				item, err := rmap.NewFromInterface(itemI)
				if err != nil {
					log.Fatalf("GetAllData rmap.NewFromInterface() failed: %s", err)
				}

				if !item.Exists("work_notes") {
					continue
				}

				worknotes := item.MustGetIterableString("work_notes")

				for _, worknoteUUID := range worknotes {
					sem <- true
					go func(worknoteUUID string) {
						defer func() { <-sem }()
						cliInternal := client.New(stuff.MandatoryEnv(endpointEnv), spaceID, func() string {
							return stuff.MandatoryEnv(tokenEnv)
						}, transport)
						log.Printf("worknoteUUID: %s", worknoteUUID)
						respI, err := rmap.NewFromBytes(cliInternal.MakeBodyRequest(fmt.Sprintf("/assets/work_note/%s?resolve=true", worknoteUUID), "GET", nil))
						if err != nil {
							log.Fatalf("GetAllData rmap.NewFromBytes() failed: %s", err)
						}

						worknote := respI.MustGetRmap("result")

						// GetReadByListDetails
						docType := worknote.MustGetString("docType")
						uuid := worknote.MustGetString("uuid")
						readByList, err := getReadByListDetails(cli, fmt.Sprintf("%s:%s", strings.ToUpper(docType), uuid))
						if err != nil {
							log.Fatal(err)
						}

						transItem, err := transformComment(worknote, readByList)
						if err != nil {
							log.Fatal(err)
						}

						log.Println(transItem)
						// store it into couchDB
						if err := cdb.InsertIfNotExist(uuid, transItem.String()); err != nil {
							log.Fatal(err)
						}
					}(worknoteUUID)
				}
			}
			for i := 0; i < cap(sem); i++ {
				sem <- true
			}
			if len(results) < 10 {
				break
			}

			bookmark = resp.MustGetString("bookmark")
		}
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
	createdByUUID := originalComment.MustGetJPtrString("/created_by/uuid")
	createdByName := ""
	createdBySurname := ""
	if createdByUUID != "941516f9-5d95-d857-c572-96f729373faa" && createdByUUID != "23e95eb8-3efe-6eaa-9cc0-f0a1bf2eea25" {
		var err error
		createdByName, err = originalComment.GetJPtrString("/created_by/private/name")
		if err != nil {
			return rmap.Rmap{}, errors.Wrapf(err, "%v ", originalComment)
		}
		createdBySurname = originalComment.MustGetJPtrString("/created_by/private/surname")
	} else {
		createdByName = "Valmet"
		createdBySurname = "Valmet"
	}

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
		//log.Println(item)
		name, err := item.GetJPtrString("/name")
		if err != nil {
			return nil, errors.Wrapf(err, "%v ", item)
		}
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
