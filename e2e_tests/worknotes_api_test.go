package e2e

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Worknotes API calls", func() {
	JustBeforeEach(func() {
		msgQueue.Clear()
	})

	Describe("POST /worknotes", func() {
		var payload []byte
		var resp *http.Response
		var originHeader bool

		BeforeEach(func() {
			createTestDatabases()
			originHeader = false
		})

		JustBeforeEach(func() {
			By("request creation")
			body := bytes.NewReader(payload)
			req, err := http.NewRequest(http.MethodPost, server.URL+"/worknotes", body)
			Expect(err).To(BeNil())
			req.Header.Set("grpc-metadata-space", testChannelID)
			req.Header.Set("authorization", bearerToken)
			if originHeader {
				req.Header.Set("X-Origin", "internal")
			}

			By("calling the endpoint")
			c := http.Client{}
			resp, err = c.Do(req)
			Expect(err).To(BeNil())
		})

		Context("with valid payload", func() {
			BeforeEach(func() {
				payload = []byte(`{
						"entity":"incident:7e0d38d1-e5f5-4211-b2aa-3b142e4da80e",
						"text": "Test Worknote 1"
					}`)
			})

			It("should return 'Created' response", func() {
				Expect(resp).To(HaveHTTPStatus(http.StatusCreated))
				Expect(resp.Header.Get("Location")).To(ContainSubstring(server.URL + "/worknotes/"))
				Expect(ioutil.ReadAll(resp.Body)).NotTo(BeEmpty())
			})

			Context("without 'X-Origin' header set", func() {
				It("should publish correct event", func() {
					events := msgQueue.LastEvents()
					Expect(events).To(HaveLen(1))
					event := events[0]
					Expect(event).To(HaveKeyWithValue("docType", "worknote"))
					Expect(event).To(HaveKeyWithValue("event", "CREATED"))
					Expect(event).To(HaveKey("uuid"))
					Expect(event).To(HaveKeyWithValue("entity", "incident:7e0d38d1-e5f5-4211-b2aa-3b142e4da80e"))
					Expect(event).To(HaveKeyWithValue("text", "Test Worknote 1"))
					Expect(event).To(HaveKeyWithValue("origin", ""))
				})
			})

			Context("with 'X-Origin' header set", func() {
				BeforeEach(func() {
					originHeader = true
				})

				It("should publish correct event", func() {
					events := msgQueue.LastEvents()
					Expect(events).To(HaveLen(1))
					event := events[0]
					Expect(event).To(HaveKeyWithValue("docType", "worknote"))
					Expect(event).To(HaveKeyWithValue("event", "CREATED"))
					Expect(event).To(HaveKey("uuid"))
					Expect(event).To(HaveKeyWithValue("entity", "incident:7e0d38d1-e5f5-4211-b2aa-3b142e4da80e"))
					Expect(event).To(HaveKeyWithValue("text", "Test Worknote 1"))
					Expect(event).To(HaveKeyWithValue("origin", "internal"))
				})
			})

			When("calling GET on returned Location header", func() {
				var locationHeader string

				JustBeforeEach(func() {
					locationHeader = resp.Header.Get("Location")
					req, err := http.NewRequest(http.MethodGet, locationHeader, nil)
					Expect(err).To(BeNil())
					req.Header.Set("grpc-metadata-space", testChannelID)
					req.Header.Set("authorization", bearerToken)

					c := http.Client{}
					resp, err = c.Do(req)
					Expect(err).To(BeNil())
				})

				Specify("we should get newly created worknote", func() {
					Expect(resp).To(HaveHTTPStatus(http.StatusOK))
					Expect(resp.Header.Get("Content-Type")).To(Equal("application/json"))

					body, err := ioutil.ReadAll(resp.Body)
					Expect(err).To(BeNil())

					bodyMap := mapFromJSON(body)

					u, err := url.Parse(locationHeader)
					Expect(err).To(BeNil())
					uuid := strings.Split(u.Path, "/")[2]

					Expect(bodyMap).To(HaveKeyWithValue("uuid", uuid))
					Expect(bodyMap).To(HaveKeyWithValue("text", "Test Worknote 1"))
					Expect(bodyMap).To(HaveKeyWithValue("entity", "incident:7e0d38d1-e5f5-4211-b2aa-3b142e4da80e"))
					Expect(bodyMap).ToNot(HaveKey("origin"))
					Expect(bodyMap).To(HaveKey("created_at"))

					createdBy, err := json.Marshal(bodyMap["created_by"])
					Expect(err).To(BeNil())

					Expect(createdBy).To(MatchJSON(expectedMockUserJSON))
				})
			})
		})

		Context("with invalid payload", func() {
			BeforeEach(func() {
				payload = []byte(`{
					"entity":"incident:7e0d38d1-e5f5-4211-b2aa-3b142e4da80e"
				}`)
			})

			It("should return error response", func() {
				Expect(resp).To(HaveHTTPStatus(http.StatusBadRequest))
				Expect(resp.Header.Get("Content-Type")).To(Equal("application/json"))

				body, err := ioutil.ReadAll(resp.Body)
				Expect(err).To(BeNil())
				Expect(mapFromJSON(body)).To(HaveKey("error"))
				Expect(body).To(MatchJSON(`{"error": "/: 'text' value is required"}`))
			})
		})
	})

	Describe("GET /worknotes", func() {
		var resp *http.Response
		var query string

		BeforeEach(func() {
			createTestDatabases()
		})

		JustBeforeEach(func() {
			By("request creation")
			req, err := http.NewRequest(http.MethodGet, server.URL+"/worknotes"+query, nil)
			Expect(err).To(BeNil())
			req.Header.Set("grpc-metadata-space", testChannelID)
			req.Header.Set("authorization", bearerToken)

			By("calling the endpoint")
			c := http.Client{}
			resp, err = c.Do(req)
			Expect(err).To(BeNil())
		})

		When("no worknotes exist", func() {
			BeforeEach(func() {
				destroyTestDatabases(storage)
				createTestDatabases()
			})

			It("should return correct response", func() {
				Expect(resp).To(HaveHTTPStatus(http.StatusOK))
				Expect(resp.Header.Get("Content-Type")).To(Equal("application/json"))

				body, err := ioutil.ReadAll(resp.Body)
				Expect(err).To(BeNil())
				Expect(body).To(MatchJSON(`{"result": [], "_links":[{"rel":"self", "href":"/worknotes"}]}`))
			})
		})

		When("some worknotes already exist", func() {
			BeforeEach(func() {
				payload1 := []byte(`{
					"entity":"incident:fc11b416-3dce-4f00-8d4e-fc43824e0b4b",
					"text": "Worknote 1"
				}`)
				createWorknote(payload1)

				payload2 := []byte(`{
					"entity":"request:cdfe52ca-0b7a-4afe-ae8d-ccb1446eae4a",
					"text": "Worknote 2"
				}`)
				createWorknote(payload2)

				payload3 := []byte(`{
					"entity":"incident:fc11b416-3dce-4f00-8d4e-fc43824e0b4b",
					"text": "Worknote 3"
				}`)
				createWorknote(payload3)

				payload4 := []byte(`{
					"entity":"incident:fc11b416-3dce-4f00-8d4e-fc43824e0b4b",
					"text": "Worknote 4"
				}`)
				createWorknote(payload4)
			})

			It("should return correct response", func() {
				Expect(resp).To(HaveHTTPStatus(http.StatusOK))
				Expect(resp.Header.Get("Content-Type")).To(Equal("application/json"))

				body, err := ioutil.ReadAll(resp.Body)
				Expect(err).To(BeNil())

				bodyMap := mapFromJSON(body)
				Expect(bodyMap).To(HaveKey("result"))
				result := bodyMap["result"]

				Expect(result).To(HaveLen(4))

				resCollection := result.([]interface{})
				for _, r := range resCollection {
					Expect(r).To(HaveKey("uuid"))
					Expect(r).To(HaveKey("text"))
					Expect(r).To(HaveKey("entity"))
					Expect(r).To(HaveKey("created_at"))

					createdBy, err := json.Marshal(r.(map[string]interface{})["created_by"])
					Expect(err).To(BeNil())

					Expect(createdBy).To(MatchJSON(expectedMockUserJSON))

					// hypermedia
					Expect(bodyMap).To(HaveKey("_links"))
					links := bodyMap["_links"].([]interface{})
					Expect(links).To(HaveLen(1))
					Expect(links[0]).To(HaveKeyWithValue("rel", "self"))
					Expect(links[0]).To(HaveKeyWithValue("href", "/worknotes"))
				}
			})
		})

		Context("with 'entity' param in query", func() {
			var skipDBReset = false

			BeforeEach(func() {
				if skipDBReset {
					return
				}
				skipDBReset = true

				destroyTestDatabases(storage)
				createTestDatabases()

				payload1 := []byte(`{
					"entity":"incident:fc11b416-3dce-4f00-8d4e-fc43824e0b4b",
					"text": "Worknote 1"
				}`)
				createWorknote(payload1)

				// this one is for other entity
				payload2 := []byte(`{
					"entity":"request:cdfe52ca-0b7a-4afe-ae8d-ccb1446eae4a",
					"text": "Worknote 2"
				}`)
				createWorknote(payload2)

				payload3 := []byte(`{
					"entity":"incident:fc11b416-3dce-4f00-8d4e-fc43824e0b4b",
					"text": "Worknote 3"
				}`)
				createWorknote(payload3)

				payload4 := []byte(`{
					"entity":"incident:fc11b416-3dce-4f00-8d4e-fc43824e0b4b",
					"text": "Worknote 4"
				}`)
				createWorknote(payload4)
			})

			BeforeEach(func() {
				query = "?entity=incident:fc11b416-3dce-4f00-8d4e-fc43824e0b4b"
			})

			It("should return correct response", func() {
				Expect(resp).To(HaveHTTPStatus(http.StatusOK))
				Expect(resp.Header.Get("Content-Type")).To(Equal("application/json"))

				body, err := ioutil.ReadAll(resp.Body)
				Expect(err).To(BeNil())

				bodyMap := mapFromJSON(body)
				Expect(bodyMap).To(HaveKey("result"))
				result := bodyMap["result"]
				Expect(result).To(HaveLen(3))

				resCollection := result.([]interface{})
				for _, r := range resCollection {
					Expect(r).To(HaveKey("uuid"))
					Expect(r).To(HaveKey("text"))
					Expect(r).To(HaveKeyWithValue("entity", "incident:fc11b416-3dce-4f00-8d4e-fc43824e0b4b"))
					Expect(r).To(HaveKey("created_at"))

					createdBy, err := json.Marshal(r.(map[string]interface{})["created_by"])
					Expect(err).To(BeNil())

					Expect(createdBy).To(MatchJSON(expectedMockUserJSON))
				}

				// hypermedia
				Expect(bodyMap).To(HaveKey("_links"))
				links := bodyMap["_links"].([]interface{})
				Expect(links).To(HaveLen(1))
				Expect(links[0]).To(HaveKeyWithValue("rel", "self"))
				Expect(links[0]).To(HaveKeyWithValue("href", "/worknotes"+query))
			})

			Context("with also 'limit' param in query", func() {
				var bookmark string

				BeforeEach(func() {
					query = "?entity=incident:fc11b416-3dce-4f00-8d4e-fc43824e0b4b&limit=2"
				})

				It("should return correct response", func() {
					Expect(resp).To(HaveHTTPStatus(http.StatusOK))
					Expect(resp.Header.Get("Content-Type")).To(Equal("application/json"))

					body, err := ioutil.ReadAll(resp.Body)
					Expect(err).To(BeNil())

					bodyMap := mapFromJSON(body)
					Expect(bodyMap).To(HaveKey("result"))
					result := bodyMap["result"]
					Expect(result).To(HaveLen(2))

					resCollection := result.([]interface{})
					for _, r := range resCollection {
						Expect(r).To(HaveKey("uuid"))
						Expect(r).To(HaveKey("text"))
						Expect(r).To(HaveKeyWithValue("entity", "incident:fc11b416-3dce-4f00-8d4e-fc43824e0b4b"))
						Expect(r).To(HaveKey("created_at"))

						createdBy, err := json.Marshal(r.(map[string]interface{})["created_by"])
						Expect(err).To(BeNil())

						Expect(createdBy).To(MatchJSON(expectedMockUserJSON))
					}

					Expect(bodyMap).To(HaveKey("bookmark"))
					bookmark = bodyMap["bookmark"].(string)

					// hypermedia
					Expect(bodyMap).To(HaveKey("_links"))
					links := bodyMap["_links"].([]interface{})
					Expect(links).To(HaveLen(2))
					Expect(links[0]).To(HaveKeyWithValue("rel", "self"))
					Expect(links[0]).To(HaveKeyWithValue("href", "/worknotes"+query))

					Expect(links[1]).To(HaveKeyWithValue("rel", "next"))
					Expect(links[1]).To(HaveKeyWithValue("href", "/worknotes"+query+"&bookmark="+bookmark))
				})

				When("called again with returned bookmark", func() {
					BeforeEach(func() {
						query = query + "&bookmark=" + bookmark
					})

					It("should return next page of worknotes", func() {
						Expect(resp).To(HaveHTTPStatus(http.StatusOK))
						Expect(resp.Header.Get("Content-Type")).To(Equal("application/json"))

						body, err := ioutil.ReadAll(resp.Body)
						Expect(err).To(BeNil())

						bodyMap := mapFromJSON(body)
						Expect(bodyMap).To(HaveKey("result"))
						result := bodyMap["result"]
						Expect(result).To(HaveLen(1))

						resCollection := result.([]interface{})
						for _, r := range resCollection {
							Expect(r).To(HaveKey("uuid"))
							Expect(r).To(HaveKey("text"))
							Expect(r).To(HaveKeyWithValue("entity", "incident:fc11b416-3dce-4f00-8d4e-fc43824e0b4b"))
							Expect(r).To(HaveKey("created_at"))

							createdBy, err := json.Marshal(r.(map[string]interface{})["created_by"])
							Expect(err).To(BeNil())

							Expect(createdBy).To(MatchJSON(expectedMockUserJSON))
						}

						Expect(bodyMap).ToNot(HaveKey("bookmark")) // last page

						// hypermedia
						Expect(bodyMap).To(HaveKey("_links"))
						links := bodyMap["_links"].([]interface{})
						Expect(links).To(HaveLen(1))
						Expect(links[0]).To(HaveKeyWithValue("rel", "self"))
						Expect(links[0]).To(HaveKeyWithValue("href", "/worknotes"+query))
					})
				})
			})
		})
	})

	Describe("GET /worknotes/{uuid}", func() {
		var resp *http.Response
		var uuid string

		BeforeEach(func() {
			destroyTestDatabases(storage)
			createTestDatabases()
		})

		JustBeforeEach(func() {
			By("request creation")
			req, err := http.NewRequest(http.MethodGet, server.URL+"/worknotes/"+uuid, nil)
			Expect(err).To(BeNil())
			req.Header.Set("grpc-metadata-space", testChannelID)
			req.Header.Set("authorization", bearerToken)

			By("calling the endpoint")
			c := http.Client{}
			resp, err = c.Do(req)
			Expect(err).To(BeNil())
		})

		When("worknote with specified UUID does not exist", func() {
			BeforeEach(func() {
				uuid = "95f2af46-0a40-463e-b2c2-87ecd77a825c"
			})

			It("should return 'Not Found' error response", func() {
				Expect(resp).To(HaveHTTPStatus(http.StatusNotFound))
				Expect(resp.Header.Get("Content-Type")).To(Equal("application/json"))

				body, err := ioutil.ReadAll(resp.Body)
				Expect(err).To(BeNil())
				Expect(body).To(MatchJSON(`{"error": "Worknote could not be retrieved: Worknote with uuid='95f2af46-0a40-463e-b2c2-87ecd77a825c' does not exist"}`))
			})
		})

		When("worknote exists", func() {
			BeforeEach(func() {
				payload1 := []byte(`{
					"entity":"request:cdfe52ca-0b7a-4afe-ae8d-ccb1446eae4a",
					"text": "Some New Worknote"
				}`)
				uuid = createWorknote(payload1)

				payload2 := []byte(`{
					"entity":"incident:fc11b416-3dce-4f00-8d4e-fc43824e0b4b",
					"text": "Worknote 2"
				}`)
				createWorknote(payload2)
			})

			It("should return correct response", func() {
				Expect(resp).To(HaveHTTPStatus(http.StatusOK))
				Expect(resp.Header.Get("Content-Type")).To(Equal("application/json"))

				body, err := ioutil.ReadAll(resp.Body)
				Expect(err).To(BeNil())

				bodyMap := mapFromJSON(body)
				Expect(bodyMap).To(HaveKeyWithValue("uuid", uuid))
				Expect(bodyMap).To(HaveKeyWithValue("text", "Some New Worknote"))
				Expect(bodyMap).To(HaveKeyWithValue("entity", "request:cdfe52ca-0b7a-4afe-ae8d-ccb1446eae4a"))
				Expect(bodyMap).To(HaveKey("created_at"))

				createdBy, err := json.Marshal(bodyMap["created_by"])
				Expect(err).To(BeNil())

				Expect(createdBy).To(MatchJSON(expectedMockUserJSON))

				// hypermedia
				Expect(bodyMap).To(HaveKey("_links"))
				links := bodyMap["_links"].([]interface{})
				Expect(links).To(HaveLen(2))
				Expect(links[0]).To(HaveKeyWithValue("rel", "self"))
				Expect(links[0]).To(HaveKeyWithValue("href", "/worknotes/"+uuid))

				Expect(links[1]).To(HaveKeyWithValue("rel", "MarkWorknoteAsReadByUser"))
				Expect(links[1]).To(HaveKeyWithValue("href", "/worknotes/"+uuid+"/read_by"))
			})
		})
	})

	Describe("POST /worknotes/{uuid}/read_by", func() {
		var resp *http.Response
		var uuid string

		BeforeEach(func() {
			destroyTestDatabases(storage)
			createTestDatabases()
		})

		JustBeforeEach(func() {
			By("request creation")
			req, err := http.NewRequest(http.MethodPost, server.URL+"/worknotes/"+uuid+"/read_by", nil)
			Expect(err).To(BeNil())
			req.Header.Set("grpc-metadata-space", testChannelID)
			req.Header.Set("authorization", bearerToken)

			By("calling the endpoint")
			c := http.Client{}
			resp, err = c.Do(req)
			Expect(err).To(BeNil())
		})

		When("worknote with specified UUID does not exist", func() {
			BeforeEach(func() {
				uuid = "95f2af46-0a40-463e-b2c2-87ecd77a825c"
			})

			It("should return correct response", func() {
				Expect(resp).To(HaveHTTPStatus(http.StatusNotFound))
				Expect(resp.Header.Get("Content-Type")).To(Equal("application/json"))

				body, err := ioutil.ReadAll(resp.Body)
				Expect(err).To(BeNil())
				Expect(body).To(MatchJSON(`{"error": "Worknote with uuid='95f2af46-0a40-463e-b2c2-87ecd77a825c' does not exist"}`))
			})
		})

		When("worknote exists", func() {
			BeforeEach(func() {
				payload1 := []byte(`{
					"entity":"request:cdfe52ca-0b7a-4afe-ae8d-ccb1446eae4a",
					"text": "Worknote to be read by user"
				}`)
				uuid = createWorknote(payload1)

				payload2 := []byte(`{
					"entity":"incident:fc11b416-3dce-4f00-8d4e-fc43824e0b4b",
					"text": "Worknote 2"
				}`)
				createWorknote(payload2)
			})

			It("should return 'Created' response", func() {
				Expect(resp).To(HaveHTTPStatus(http.StatusCreated))
				Expect(ioutil.ReadAll(resp.Body)).To(BeEmpty())
			})

			Describe("worknote", func() {
				var worknoteResp *http.Response

				JustBeforeEach(func() {
					req, err := http.NewRequest(http.MethodGet, server.URL+"/worknotes/"+uuid, nil)
					Expect(err).To(BeNil())
					req.Header.Set("grpc-metadata-space", testChannelID)
					req.Header.Set("authorization", bearerToken)

					c := http.Client{}
					worknoteResp, err = c.Do(req)
					Expect(err).To(BeNil())
				})

				It("should be marked as read", func() {
					Expect(worknoteResp).To(HaveHTTPStatus(http.StatusOK))
					Expect(worknoteResp.Header.Get("Content-Type")).To(Equal("application/json"))

					body, err := ioutil.ReadAll(worknoteResp.Body)
					Expect(err).To(BeNil())

					bodyMap := mapFromJSON(body)
					Expect(bodyMap).To(HaveKeyWithValue("uuid", uuid))
					Expect(bodyMap).To(HaveKeyWithValue("text", "Worknote to be read by user"))
					Expect(bodyMap).To(HaveKeyWithValue("entity", "request:cdfe52ca-0b7a-4afe-ae8d-ccb1446eae4a"))
					Expect(bodyMap).To(HaveKey("created_at"))

					createdBy, err := json.Marshal(bodyMap["created_by"])
					Expect(err).To(BeNil())

					Expect(createdBy).To(MatchJSON(expectedMockUserJSON))

					Expect(bodyMap).To(HaveKey("read_by"))
					Expect(bodyMap["read_by"]).To(HaveLen(1))

					readBy := bodyMap["read_by"].([]interface{})[0]
					Expect(readBy).To(HaveKey("time"))
					Expect(readBy).To(HaveKey("user"))

					user, err := json.Marshal(readBy.(map[string]interface{})["user"])
					Expect(err).To(BeNil())

					Expect(user).To(MatchJSON(expectedMockUserJSON))
				})
			})
		})

		Context("when called multiple times by the same user", func() {
			BeforeEach(func() {
				payload1 := []byte(`{
					"entity":"request:cdfe52ca-0b7a-4afe-ae8d-ccb1446eae4a",
					"text": "Worknote to be read by user"
				}`)
				uuid = createWorknote(payload1)
			})

			JustBeforeEach(func() {
				By("request creation")
				req, err := http.NewRequest(http.MethodPost, server.URL+"/worknotes/"+uuid+"/read_by", nil)
				Expect(err).To(BeNil())
				req.Header.Set("grpc-metadata-space", testChannelID)
				req.Header.Set("authorization", bearerToken)

				By("calling the endpoint")
				c := http.Client{}
				resp, err = c.Do(req)
				Expect(err).To(BeNil())
			})

			It("should return 'No Content' response", func() {
				Expect(resp).To(HaveHTTPStatus(http.StatusNoContent))
				Expect(ioutil.ReadAll(resp.Body)).To(BeEmpty())
			})

			Describe("worknote", func() {
				var worknoteResp *http.Response

				JustBeforeEach(func() {
					req, err := http.NewRequest(http.MethodGet, server.URL+"/worknotes/"+uuid, nil)
					Expect(err).To(BeNil())
					req.Header.Set("grpc-metadata-space", testChannelID)
					req.Header.Set("authorization", bearerToken)

					c := http.Client{}
					worknoteResp, err = c.Do(req)
					Expect(err).To(BeNil())
				})

				It("should not be changed", func() {
					Expect(worknoteResp).To(HaveHTTPStatus(http.StatusOK))
					Expect(worknoteResp.Header.Get("Content-Type")).To(Equal("application/json"))

					body, err := ioutil.ReadAll(worknoteResp.Body)
					Expect(err).To(BeNil())

					bodyMap := mapFromJSON(body)
					Expect(bodyMap).To(HaveKeyWithValue("uuid", uuid))
					Expect(bodyMap).To(HaveKeyWithValue("text", "Worknote to be read by user"))
					Expect(bodyMap).To(HaveKeyWithValue("entity", "request:cdfe52ca-0b7a-4afe-ae8d-ccb1446eae4a"))
					Expect(bodyMap).To(HaveKey("created_at"))

					createdBy, err := json.Marshal(bodyMap["created_by"])
					Expect(err).To(BeNil())

					Expect(createdBy).To(MatchJSON(expectedMockUserJSON))

					Expect(bodyMap).To(HaveKey("read_by"))
					Expect(bodyMap["read_by"]).To(HaveLen(1))

					readBy := bodyMap["read_by"].([]interface{})[0]
					Expect(readBy).To(HaveKey("time"))
					Expect(readBy).To(HaveKey("user"))

					user, err := json.Marshal(readBy.(map[string]interface{})["user"])
					Expect(err).To(BeNil())

					Expect(user).To(MatchJSON(expectedMockUserJSON))
				})
			})
		})
	})
})
