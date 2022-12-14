// Package rest Commenting service API
//
// Documentation for Commenting service API.
//
// It works for both comments and worknotes resources. If you want to use it for worknotes, just switch 'comment'
// part of the endpoint path with 'worknote', for example:
//
// "GET /comments" endpoint returns comments
//
// "GET /worknotes" endpoint returns worknotes
//
//	Schemes: http
//	BasePath: /
//	Version: 1.0.0
//
//	Consumes:
//	- application/json
//
//	Produces:
//	- application/json
//
// swagger:meta
package rest

import (
	"github.com/KompiTech/itsm-commenting-service/pkg/domain/comment"
	"github.com/KompiTech/itsm-commenting-service/pkg/domain/entity"
)

// NOTE: Types defined here are purely for documentation purposes
// these types are not used by any of the handlers

// Error
// swagger:response errorResponse
type errorResponseWrapper struct {
	// in: body
	Body struct {
		// required: true
		// Description of the error
		ErrorMessage string `json:"error"`
	}
}

// Bad Request
// swagger:response errorResponse400
type errorResponseWrapper400 errorResponseWrapper

// Unauthorized
// swagger:response errorResponse401
type errorResponseWrapper401 errorResponseWrapper

// Forbidden
// swagger:response errorResponse403
type errorResponseWrapper403 errorResponseWrapper

// Not Found
// swagger:response errorResponse404
type errorResponseWrapper404 errorResponseWrapper

// Conflict
// swagger:response errorResponse409
type errorResponseWrapper409 errorResponseWrapper

// Created
// swagger:response createdResponse
type createdResponseWrapper struct {
	// URI of the resource
	// example: http://localhost:8080/comments/2af4f493-0bd5-4513-b440-6cbb465feadb
	// in: header
	Location string
}

// HypermediaLinks contain links to other API calls
type HypermediaLinks struct {
	Rel struct {
		// swagger:strfmt uri
		Href string `json:"href"`
	} `json:"self"`
}

// Created
// swagger:response commentCreatedResponse
type commentCreatedResponseWrapper struct {
	// URI of the resource
	// example: http://localhost:8080/comments/2af4f493-0bd5-4513-b440-6cbb465feadb
	// in: header
	Location string
	// in: body
	Body comment.Comment
}

// No content
// swagger:response noContentResponse
type noContentResponseWrapper struct {
	// URI of the resource
	// example: http://localhost:8080/comments/2af4f493-0bd5-4513-b440-6cbb465feadb
	// in: header
	Location string
}

// No content
// swagger:response databasesNoContentResponse
type databasesNoContentResponseWrapper struct{}

// Created
// swagger:response databasesCreatedResponse
type databasesCreatedResponseWrapper struct{}

// A list of comments or worknotes
// swagger:response commentsListResponse
type commentsListResponseWrapper struct {
	// in: body
	Body struct {
		// required: true
		Result []comment.Comment `json:"result"`
		// Pagination bookmark
		Bookmark string          `json:"bookmark"`
		Links    HypermediaLinks `json:"_links"`
	}
}

// Data structure representing a single comment or worknote
// swagger:response commentResponse
type commentResponseWrapper struct {
	// in: body
	Body struct {
		comment.Comment
		Links HypermediaLinks `json:"_links"`
	}
}

// AuthorizationHeaders represents general authorization header parameters used in many API calls
type AuthorizationHeaders struct {
	// Bearer token
	// in: header
	// required: true
	Authorization string `json:"authorization"`

	// in: header
	// required: true
	// swagger:strfmt uuid
	ChannelID string `json:"grpc-metadata-space"`
}

// swagger:parameters GetComment GetWorknote MarkCommentAsReadByUser MarkWorknoteAsReadByUser
type commentIDParameterWrapper struct {
	AuthorizationHeaders

	// ID of the comment/worknote
	// in: path
	// required: true
	// swagger:strfmt uuid
	UUID string `json:"uuid"`
}

// swagger:parameters ListComments ListWorknotes
type listCommentsParameterWrapper struct {
	AuthorizationHeaders

	// Entity represents some external entity reference in the form "&lt;entity&gt;:&lt;UUID&gt;"
	// example: incident:f49d5fd5-8da4-4779-b5ba-32e78aa2c444
	// in: query
	// swagger:strfmt string
	Entity entity.Entity `json:"entity"`

	// Amount of records to be returned (pagination)
	// default: 25
	// in: query
	Limit int `json:"limit"`

	// Pagination bookmark
	// in: query
	Bookmark string `json:"bookmark"`
}

// swagger:parameters AddComment AddWorknote
type commentParamWrapper struct {
	AuthorizationHeaders

	// in: header
	// swagger:strfmt uuid
	OnBehalf string `json:"on_behalf"`

	// Origin of the request (will be present in event message)
	// in: header
	// example: ServiceNow
	XOrigin string `json:"X-Origin"`

	// Comment/Worknote data structure to create
	// in: body
	Body struct {
		// Entity represents some external entity reference in the form "&lt;entity&gt;:&lt;UUID&gt;"
		// required: true
		// example: incident:f49d5fd5-8da4-4779-b5ba-32e78aa2c444
		Entity string `json:"entity"`

		// ID in external system
		// required: false
		ExternalID string `json:"external_id"`

		// Content of the comment/worknote
		// required: true
		Text string `json:"text"`
	}
}

// swagger:parameters CreateDatabases
type databasesParamWrapper struct {
	// Bearer token
	// in: header
	// required: true
	Authorization string `json:"authorization"`

	// ChannelID for which to create databases
	// in: body
	Body struct {
		// required: true
		// swagger:strfmt uuid
		ChannelID string `json:"channel_id"`
	}
}
