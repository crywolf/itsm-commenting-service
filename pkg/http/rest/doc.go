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
		// Description of the error
		ErrorMessage string `json:"error"`
	}
}

// Created is returned by this API endpoint
// swagger:response createdResponse
type createdResponseWrapper struct {
	// URI of the resource
	// example: http://localhost:8080/comments/2af4f493-0bd5-4513-b440-6cbb465feadb
	Location string
}

// No content is returned by this API endpoint
// swagger:response databasesNoContentResponse
type databasesNoContentResponseWrapper struct{}

// Created is returned by this API endpoint
// swagger:response databasesCreatedResponse
type databasesCreatedResponseWrapper struct{}

// No content is returned by this API endpoint
// swagger:response noContentResponse
type noContentResponseWrapper struct {
	// URI of the resource
	// example: http://localhost:8080/comments/2af4f493-0bd5-4513-b440-6cbb465feadb
	Location string
}

// A list of comments or worknotes
// swagger:response commentsResponse
type commentsResponseWrapper struct {
	// in: body
	Body []comment.Comment
}

// Data structure representing a single comment or worknote
// swagger:response commentResponse
type commentResponseWrapper struct {
	// in: body
	Body comment.Comment
}

// swagger:parameters GetComment MarkAsReadByUser
type commentIDParameterWrapper struct {
	// Bearer token
	// in: header
	// required: true
	Authorization string `json:"authorization"`

	// in: header
	// required: true
	// swagger:strfmt uuid
	ChannelID string `json:"grpc-metadata-space"`

	// ID of the comment
	// in: path
	// required: true
	// swagger:strfmt uuid
	UUID string `json:"uuid"`
}

// swagger:parameters ListComments
type listCommentsParameterWrapper struct {
	// Bearer token
	// in: header
	// required: true
	Authorization string `json:"authorization"`

	// in: header
	// required: true
	// swagger:strfmt uuid
	ChannelID string `json:"grpc-metadata-space"`

	// Entity represents some external entity reference in the form "&lt;entity&gt;:&lt;UUID&gt;"
	// in: query
	// example: incident:f49d5fd5-8da4-4779-b5ba-32e78aa2c444
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

// swagger:parameters AddComment
type commentParamWrapper struct {
	// Bearer token
	// in: header
	// required: true
	Authorization string `json:"authorization"`

	// in: header
	// required: true
	// swagger:strfmt uuid
	ChannelID string `json:"grpc-metadata-space"`

	// Comment/Worknote data structure to Create.
	// in: body
	Body struct {
		// Entity represents some external entity reference in the form "&lt;entity&gt;:&lt;UUID&gt;"
		// required: true
		// example: incident:f49d5fd5-8da4-4779-b5ba-32e78aa2c444
		Entity string `json:"entity"`

		// ID in external system
		// required: false
		ExternalID string `json:"external_id"`

		// Content of the comment
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
