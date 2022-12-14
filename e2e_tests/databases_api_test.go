package e2e

import (
	"bytes"
	"io/ioutil"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Databases: API call", func() {
	Describe("POST /databases", func() {
		var payload []byte
		var resp *http.Response

		BeforeEach(func() {
			destroyTestDatabases(storage)
		})

		JustBeforeEach(func() {
			By("request creation")
			body := bytes.NewReader(payload)
			req, err := http.NewRequest(http.MethodPost, server.URL+"/databases", body)
			Expect(err).To(BeNil())
			req.Header.Set("grpc-metadata-space", testChannelID)
			req.Header.Set("authorization", bearerToken)

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
					createTestDatabases()
				})

				It("should return 'No content' response", func() {
					Expect(resp).To(HaveHTTPStatus(http.StatusNoContent))
					Expect(ioutil.ReadAll(resp.Body)).To(BeEmpty())
				})
			})
		})
	})
})
