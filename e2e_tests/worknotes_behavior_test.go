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

var _ = Describe("Worknotes behavior", func() {
	JustBeforeEach(func() {
		msgQueue.Clear()
	})

	Describe("Creation", func() {
		var payload []byte
		var resp *http.Response

		BeforeEach(func() {
			createTestDatabases()
		})

		When("when creating worknote on behalf of other user (by using 'on_behalf' header)", func() {
			BeforeEach(func() {
				payload = []byte(`{
					"entity":"k_request:fef298da-8ee8-4e0b-8c8f-327b17dc16bb",
					"text": "Worknote On Behalf of Someone Else"
				}`)
			})

			BeforeEach(func() {
				By("request creation")
				body := bytes.NewReader(payload)
				req, err := http.NewRequest(http.MethodPost, server.URL+"/worknotes", body)
				Expect(err).To(BeNil())
				req.Header.Set("grpc-metadata-space", testChannelID)
				req.Header.Set("authorization", bearerToken)
				req.Header.Set("on_behalf", "9abc8dc2-a894-40b1-81ea-22a476fe6d34")

				By("calling the endpoint")
				c := http.Client{}
				resp, err = c.Do(req)
				Expect(err).To(BeNil())
			})

			var locationHeader string

			JustBeforeEach(func() {
				By("calling the Location header URL")
				locationHeader = resp.Header.Get("Location")
				req, err := http.NewRequest(http.MethodGet, locationHeader, nil)
				Expect(err).To(BeNil())
				req.Header.Set("grpc-metadata-space", testChannelID)
				req.Header.Set("authorization", bearerToken)

				c := http.Client{}
				resp, err = c.Do(req)
				Expect(err).To(BeNil())
			})

			Specify("created worknote should have 'created_by' field filled with correct user data", func() {
				Expect(resp).To(HaveHTTPStatus(http.StatusOK))
				Expect(resp.Header.Get("Content-Type")).To(Equal("application/json"))

				body, err := ioutil.ReadAll(resp.Body)
				Expect(err).To(BeNil())

				bodyMap := mapFromJSON(body)

				u, err := url.Parse(locationHeader)
				Expect(err).To(BeNil())
				uuid := strings.Split(u.Path, "/")[2]

				Expect(bodyMap).To(HaveKeyWithValue("uuid", uuid))
				Expect(bodyMap).To(HaveKeyWithValue("text", "Worknote On Behalf of Someone Else"))
				Expect(bodyMap).To(HaveKeyWithValue("entity", "k_request:fef298da-8ee8-4e0b-8c8f-327b17dc16bb"))
				Expect(bodyMap).To(HaveKey("created_at"))

				createdBy, err := json.Marshal(bodyMap["created_by"])
				Expect(err).To(BeNil())

				Expect(createdBy).To(MatchJSON(expectedMockOnBehalfUserJSON))
			})
		})
	})

	Describe("Reading", func() {
		var uuid string
		var worknoteResp *http.Response

		BeforeEach(func() {
			createTestDatabases()
		})

		When("reading worknote on behalf of other user (by using 'on_behalf' header)", func() {
			BeforeEach(func() {
				payload := []byte(`{
					"entity":"incident:fef298da-8ee8-4e0b-8c8f-327b17dc16bb",
					"text": "Worknote to be read by someone else"
				}`)

				uuid = createWorknote(payload)

				By("marking as read")
				req, err := http.NewRequest(http.MethodPost, server.URL+"/worknotes/"+uuid+"/read_by", nil)
				Expect(err).To(BeNil())
				req.Header.Set("grpc-metadata-space", testChannelID)
				req.Header.Set("authorization", bearerToken)
				req.Header.Set("on_behalf", "9abc8dc2-a894-40b1-81ea-22a476fe6d34")

				c := http.Client{}
				_, err = c.Do(req)
				Expect(err).To(BeNil())
			})

			JustBeforeEach(func() {
				By("getting worknote")
				req, err := http.NewRequest(http.MethodGet, server.URL+"/worknotes/"+uuid, nil)
				Expect(err).To(BeNil())
				req.Header.Set("grpc-metadata-space", testChannelID)
				req.Header.Set("authorization", bearerToken)

				c := http.Client{}
				worknoteResp, err = c.Do(req)
				Expect(err).To(BeNil())
			})

			Specify("worknote should be marked as read by correct user", func() {
				Expect(worknoteResp).To(HaveHTTPStatus(http.StatusOK))
				Expect(worknoteResp.Header.Get("Content-Type")).To(Equal("application/json"))

				body, err := ioutil.ReadAll(worknoteResp.Body)
				Expect(err).To(BeNil())

				bodyMap := mapFromJSON(body)
				Expect(bodyMap).To(HaveKeyWithValue("uuid", uuid))
				Expect(bodyMap).To(HaveKeyWithValue("text", "Worknote to be read by someone else"))
				Expect(bodyMap).To(HaveKeyWithValue("entity", "incident:fef298da-8ee8-4e0b-8c8f-327b17dc16bb"))
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

				Expect(user).To(MatchJSON(expectedMockOnBehalfUserJSON))
			})
		})
	})

})
