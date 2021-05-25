package e2e

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/KompiTech/itsm-commenting-service/pkg/repository/couchdb"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// TestE2E initializes test suite
func TestE2E(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "End2End tests")
}

var server *httptest.Server
var storage *couchdb.DBStorage

var _ = BeforeSuite(func() {
	server, storage = StartServer()
})

var _ = AfterSuite(func() {
	destroyTestDatabases(storage)
	server.Close()
})

var _ = Describe("API call", func() {
	Describe("POST /databases", func() {
		var payload []byte
		var resp *http.Response

		JustBeforeEach(func() {
			By("request creation")
			body := bytes.NewReader(payload)
			req, err := http.NewRequest(http.MethodPost, server.URL+"/databases", body)
			Expect(err).To(BeNil())

			By("calling the endpoint")
			c := http.Client{}
			resp, err = c.Do(req)
			Expect(err).To(BeNil())
		})

		Context("with invalid payload - missing 'channel_id'", func() {
			BeforeEach(func() {
				payload = []byte(`{}`)
			})

			It("should return error response", func() {
				Expect(resp).To(HaveHTTPStatus(http.StatusBadRequest))
				Expect(resp.Header.Get("Content-Type")).To(Equal("application/json"))
				body, err := ioutil.ReadAll(resp.Body)
				Expect(err).To(BeNil())
				Expect(mapFromJSON(body)).To(HaveKey("error"))
				Expect(body).To(MatchJSON(`{"error": "/: 'channel_id' value is required"}`))
			})
		})

		Context("with valid payload", func() {
			BeforeEach(func() {
				payload = []byte(`{"channel_id": "` + testChannelID + `"}`)
			})

			When("databases do not exist", func() {
				It("should return 'Created' response", func() {
					Expect(resp).To(HaveHTTPStatus(http.StatusCreated))
					Expect(resp.Header.Get("Content-Type")).To(Equal("application/json"))

					body, err := ioutil.ReadAll(resp.Body)
					Expect(err).To(BeNil())
					Expect(body).To(MatchJSON(`{"message": "databases were successfully created"}`))
				})
			})

			When("databases already exist", func() {
				BeforeEach(func() {
					createDatabases()
				})

				It("should return 'No content' response", func() {
					Expect(resp).To(HaveHTTPStatus(http.StatusNoContent))
					Expect(ioutil.ReadAll(resp.Body)).To(BeEmpty())
				})
			})
		})
	})

	Describe("POST /comments", func() {
		var payload []byte
		var resp *http.Response

		BeforeEach(func() {
			createDatabases()
		})

		JustBeforeEach(func() {
			By("request creation")
			body := bytes.NewReader(payload)
			req, err := http.NewRequest(http.MethodPost, server.URL+"/comments", body)
			Expect(err).To(BeNil())
			req.Header.Set("grpc-metadata-space", testChannelID)
			req.Header.Set("authorization", bearerToken)

			By("calling the endpoint")
			c := http.Client{}
			resp, err = c.Do(req)
			Expect(err).To(BeNil())
		})

		Context("with valid payload", func() {
			BeforeEach(func() {
				payload = []byte(`{
					"entity":"incident:7e0d38d1-e5f5-4211-b2aa-3b142e4da80e",
					"text": "Test Comment 1"
				}`)
			})

			It("should return 'Created' response", func() {
				Expect(resp).To(HaveHTTPStatus(http.StatusCreated))
				Expect(resp.Header.Get("Location")).To(ContainSubstring(server.URL + "/comments/"))
				Expect(ioutil.ReadAll(resp.Body)).To(BeEmpty())
			})

			When("calling GET on returned Location header", func() {
				var locationHeader string

				JustBeforeEach(func() {
					locationHeader = resp.Header.Get("Location")
					req, err := http.NewRequest(http.MethodGet, locationHeader, nil)
					Expect(err).To(BeNil())
					req.Header.Set("grpc-metadata-space", testChannelID)

					c := http.Client{}
					resp, err = c.Do(req)
					Expect(err).To(BeNil())
				})

				Specify("we should get newly created comment", func() {
					Expect(resp).To(HaveHTTPStatus(http.StatusOK))
					Expect(resp.Header.Get("Content-Type")).To(Equal("application/json"))

					body, err := ioutil.ReadAll(resp.Body)
					Expect(err).To(BeNil())

					bodyMap := mapFromJSON(body)

					u, err := url.Parse(locationHeader)
					Expect(err).To(BeNil())
					uuid := strings.Split(u.Path, "/")[2]

					Expect(bodyMap).To(HaveKeyWithValue("uuid", uuid))
					Expect(bodyMap).To(HaveKeyWithValue("text", "Test Comment 1"))
					Expect(bodyMap).To(HaveKeyWithValue("entity", "incident:7e0d38d1-e5f5-4211-b2aa-3b142e4da80e"))
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

	Describe("GET /comments", func() {
		var resp *http.Response
		var query string

		BeforeEach(func() {
			createDatabases()
		})

		JustBeforeEach(func() {
			By("request creation")
			req, err := http.NewRequest(http.MethodGet, server.URL+"/comments"+query, nil)
			Expect(err).To(BeNil())
			req.Header.Set("grpc-metadata-space", testChannelID)

			By("calling the endpoint")
			c := http.Client{}
			resp, err = c.Do(req)
			Expect(err).To(BeNil())
		})

		When("no comments exist", func() {
			BeforeEach(func() {
				destroyTestDatabases(storage)
				createDatabases()
			})

			It("should return correct response", func() {
				Expect(resp).To(HaveHTTPStatus(http.StatusOK))
				Expect(resp.Header.Get("Content-Type")).To(Equal("application/json"))

				body, err := ioutil.ReadAll(resp.Body)
				Expect(err).To(BeNil())
				Expect(body).To(MatchJSON(`{"result": []}`))
			})
		})

		When("some comments already exist", func() {
			BeforeEach(func() {
				payload1 := []byte(`{
					"entity":"incident:fc11b416-3dce-4f00-8d4e-fc43824e0b4b",
					"text": "Comment 1"
				}`)
				createComment(payload1)

				payload2 := []byte(`{
					"entity":"request:cdfe52ca-0b7a-4afe-ae8d-ccb1446eae4a",
					"text": "Comment 2"
				}`)
				createComment(payload2)

				payload3 := []byte(`{
					"entity":"incident:fc11b416-3dce-4f00-8d4e-fc43824e0b4b",
					"text": "Comment 3"
				}`)
				createComment(payload3)

				payload4 := []byte(`{
					"entity":"incident:fc11b416-3dce-4f00-8d4e-fc43824e0b4b",
					"text": "Comment 4"
				}`)
				createComment(payload4)
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
				}
			})
		})

		Context("with 'entity' param in query", func() {
			var skipDBReset = false

			BeforeEach(func() {
				if skipDBReset {
					return
				}

				destroyTestDatabases(storage)
				createDatabases()

				payload1 := []byte(`{
					"entity":"incident:fc11b416-3dce-4f00-8d4e-fc43824e0b4b",
					"text": "Comment 1"
				}`)
				createComment(payload1)

				payload2 := []byte(`{
					"entity":"request:cdfe52ca-0b7a-4afe-ae8d-ccb1446eae4a",
					"text": "Comment 2"
				}`)
				createComment(payload2)

				payload3 := []byte(`{
					"entity":"incident:fc11b416-3dce-4f00-8d4e-fc43824e0b4b",
					"text": "Comment 3"
				}`)
				createComment(payload3)

				payload4 := []byte(`{
					"entity":"incident:fc11b416-3dce-4f00-8d4e-fc43824e0b4b",
					"text": "Comment 4"
				}`)
				createComment(payload4)
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
				})

				When("called again with returned bookmark", func() {
					skipDBReset = true

					BeforeEach(func() {
						query = query + "&bookmark=" + bookmark
					})

					It("should return next page of comments", func() {
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
					})
				})
			})
		})
	})

	Describe("GET /comments/{uuid}", func() {
		var resp *http.Response
		var uuid string

		BeforeEach(func() {
			destroyTestDatabases(storage)
			createDatabases()
		})

		JustBeforeEach(func() {
			By("request creation")
			req, err := http.NewRequest(http.MethodGet, server.URL+"/comments/"+uuid, nil)
			Expect(err).To(BeNil())
			req.Header.Set("grpc-metadata-space", testChannelID)

			By("calling the endpoint")
			c := http.Client{}
			resp, err = c.Do(req)
			Expect(err).To(BeNil())
		})

		When("comment with specified UUID does not exist", func() {
			BeforeEach(func() {
				uuid = "95f2af46-0a40-463e-b2c2-87ecd77a825c"
			})

			It("should return 'Not Found' error response", func() {
				Expect(resp).To(HaveHTTPStatus(http.StatusNotFound))
				Expect(resp.Header.Get("Content-Type")).To(Equal("application/json"))

				body, err := ioutil.ReadAll(resp.Body)
				Expect(err).To(BeNil())
				Expect(body).To(MatchJSON(`{"error": "Comment could not be retrieved: Comment with uuid='95f2af46-0a40-463e-b2c2-87ecd77a825c' does not exist"}`))
			})
		})

		When("comment exists", func() {
			BeforeEach(func() {
				payload1 := []byte(`{
					"entity":"request:cdfe52ca-0b7a-4afe-ae8d-ccb1446eae4a",
					"text": "Some New Comment"
				}`)
				uuid = createComment(payload1)

				payload2 := []byte(`{
					"entity":"incident:fc11b416-3dce-4f00-8d4e-fc43824e0b4b",
					"text": "Comment 2"
				}`)
				createComment(payload2)
			})

			It("should return correct response", func() {
				Expect(resp).To(HaveHTTPStatus(http.StatusOK))
				Expect(resp.Header.Get("Content-Type")).To(Equal("application/json"))

				body, err := ioutil.ReadAll(resp.Body)
				Expect(err).To(BeNil())

				bodyMap := mapFromJSON(body)
				Expect(bodyMap).To(HaveKeyWithValue("uuid", uuid))
				Expect(bodyMap).To(HaveKeyWithValue("text", "Some New Comment"))
				Expect(bodyMap).To(HaveKeyWithValue("entity", "request:cdfe52ca-0b7a-4afe-ae8d-ccb1446eae4a"))
				Expect(bodyMap).To(HaveKey("created_at"))

				createdBy, err := json.Marshal(bodyMap["created_by"])
				Expect(err).To(BeNil())

				Expect(createdBy).To(MatchJSON(expectedMockUserJSON))
			})
		})
	})

	Describe("POST /comments/{uuid}/read_by", func() {
		var resp *http.Response
		var uuid string

		BeforeEach(func() {
			destroyTestDatabases(storage)
			createDatabases()
		})

		JustBeforeEach(func() {
			By("request creation")
			req, err := http.NewRequest(http.MethodPost, server.URL+"/comments/"+uuid+"/read_by", nil)
			Expect(err).To(BeNil())
			req.Header.Set("grpc-metadata-space", testChannelID)
			req.Header.Set("authorization", bearerToken)

			By("calling the endpoint")
			c := http.Client{}
			resp, err = c.Do(req)
			Expect(err).To(BeNil())
		})

		When("comment with specified UUID does not exist", func() {
			BeforeEach(func() {
				uuid = "95f2af46-0a40-463e-b2c2-87ecd77a825c"
			})

			It("should return correct response", func() {
				Expect(resp).To(HaveHTTPStatus(http.StatusNotFound))
				Expect(resp.Header.Get("Content-Type")).To(Equal("application/json"))

				body, err := ioutil.ReadAll(resp.Body)
				Expect(err).To(BeNil())
				Expect(body).To(MatchJSON(`{"error": "Comment with uuid='95f2af46-0a40-463e-b2c2-87ecd77a825c' does not exist"}`))
			})
		})

		When("comment exists", func() {
			BeforeEach(func() {
				payload1 := []byte(`{
					"entity":"request:cdfe52ca-0b7a-4afe-ae8d-ccb1446eae4a",
					"text": "Comment to be read by user"
				}`)
				uuid = createComment(payload1)

				payload2 := []byte(`{
					"entity":"incident:fc11b416-3dce-4f00-8d4e-fc43824e0b4b",
					"text": "Comment 2"
				}`)
				createComment(payload2)
			})

			It("should return 'Created' response", func() {
				Expect(resp).To(HaveHTTPStatus(http.StatusCreated))
				Expect(ioutil.ReadAll(resp.Body)).To(BeEmpty())
			})

			Describe("comment", func() {
				var commentResp *http.Response

				JustBeforeEach(func() {
					req, err := http.NewRequest(http.MethodGet, server.URL+"/comments/"+uuid, nil)
					Expect(err).To(BeNil())
					req.Header.Set("grpc-metadata-space", testChannelID)

					c := http.Client{}
					commentResp, err = c.Do(req)
					Expect(err).To(BeNil())
				})

				It("should be marked as read", func() {
					Expect(commentResp).To(HaveHTTPStatus(http.StatusOK))
					Expect(commentResp.Header.Get("Content-Type")).To(Equal("application/json"))

					body, err := ioutil.ReadAll(commentResp.Body)
					Expect(err).To(BeNil())

					bodyMap := mapFromJSON(body)
					Expect(bodyMap).To(HaveKeyWithValue("uuid", uuid))
					Expect(bodyMap).To(HaveKeyWithValue("text", "Comment to be read by user"))
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
					"text": "Comment to be read by user"
				}`)
				uuid = createComment(payload1)
			})

			JustBeforeEach(func() {
				By("request creation")
				req, err := http.NewRequest(http.MethodPost, server.URL+"/comments/"+uuid+"/read_by", nil)
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

			Describe("comment", func() {
				var commentResp *http.Response

				JustBeforeEach(func() {
					req, err := http.NewRequest(http.MethodGet, server.URL+"/comments/"+uuid, nil)
					Expect(err).To(BeNil())
					req.Header.Set("grpc-metadata-space", testChannelID)

					c := http.Client{}
					commentResp, err = c.Do(req)
					Expect(err).To(BeNil())
				})

				It("should not be changed", func() {
					Expect(commentResp).To(HaveHTTPStatus(http.StatusOK))
					Expect(commentResp.Header.Get("Content-Type")).To(Equal("application/json"))

					body, err := ioutil.ReadAll(commentResp.Body)
					Expect(err).To(BeNil())

					bodyMap := mapFromJSON(body)
					Expect(bodyMap).To(HaveKeyWithValue("uuid", uuid))
					Expect(bodyMap).To(HaveKeyWithValue("text", "Comment to be read by user"))
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

func createDatabases() {
	payload := []byte(`{"channel_id": "` + testChannelID + `"}`)
	body := bytes.NewReader(payload)
	req, err := http.NewRequest(http.MethodPost, server.URL+"/databases", body)
	Expect(err).To(BeNil())

	c := http.Client{}
	resp, err := c.Do(req)
	Expect(err).To(BeNil())
	Expect(resp.StatusCode).To(Or(Equal(http.StatusCreated), Equal(http.StatusNoContent)), "createDatabases() failed: %s", resp.Status)
}

// createComment calls endpoint for comment creation and returns UUID of a newly created comment
func createComment(payload []byte) string {
	body := bytes.NewReader(payload)
	req, err := http.NewRequest(http.MethodPost, server.URL+"/comments", body)
	Expect(err).To(BeNil())
	req.Header.Set("grpc-metadata-space", testChannelID)
	req.Header.Set("authorization", bearerToken)

	c := http.Client{}
	resp, err := c.Do(req)
	Expect(err).To(BeNil())
	Expect(resp.StatusCode).To(Equal(http.StatusCreated), "createComment() failed: %s", resp.Status)

	u, err := url.Parse(resp.Header.Get("Location"))
	Expect(err).To(BeNil())
	return strings.Split(u.Path, "/")[2]
}
