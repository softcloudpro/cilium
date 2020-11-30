// Code generated by go-swagger; DO NOT EDIT.

// Copyright 2017-2020 Authors of Cilium
// SPDX-License-Identifier: Apache-2.0

package daemon

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"github.com/cilium/cilium/api/v1/models"
)

// DeleteClusterNodesNeighOKCode is the HTTP code returned for type DeleteClusterNodesNeighOK
const DeleteClusterNodesNeighOKCode int = 200

/*DeleteClusterNodesNeighOK Success

swagger:response deleteClusterNodesNeighOK
*/
type DeleteClusterNodesNeighOK struct {
}

// NewDeleteClusterNodesNeighOK creates DeleteClusterNodesNeighOK with default headers values
func NewDeleteClusterNodesNeighOK() *DeleteClusterNodesNeighOK {

	return &DeleteClusterNodesNeighOK{}
}

// WriteResponse to the client
func (o *DeleteClusterNodesNeighOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.Header().Del(runtime.HeaderContentType) //Remove Content-Type on empty responses

	rw.WriteHeader(200)
}

// DeleteClusterNodesNeighNotFoundCode is the HTTP code returned for type DeleteClusterNodesNeighNotFound
const DeleteClusterNodesNeighNotFoundCode int = 404

/*DeleteClusterNodesNeighNotFound Node neighbor not found

swagger:response deleteClusterNodesNeighNotFound
*/
type DeleteClusterNodesNeighNotFound struct {
}

// NewDeleteClusterNodesNeighNotFound creates DeleteClusterNodesNeighNotFound with default headers values
func NewDeleteClusterNodesNeighNotFound() *DeleteClusterNodesNeighNotFound {

	return &DeleteClusterNodesNeighNotFound{}
}

// WriteResponse to the client
func (o *DeleteClusterNodesNeighNotFound) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.Header().Del(runtime.HeaderContentType) //Remove Content-Type on empty responses

	rw.WriteHeader(404)
}

// DeleteClusterNodesNeighFailureCode is the HTTP code returned for type DeleteClusterNodesNeighFailure
const DeleteClusterNodesNeighFailureCode int = 500

/*DeleteClusterNodesNeighFailure Error while deleting node neighbor

swagger:response deleteClusterNodesNeighFailure
*/
type DeleteClusterNodesNeighFailure struct {

	/*
	  In: Body
	*/
	Payload models.Error `json:"body,omitempty"`
}

// NewDeleteClusterNodesNeighFailure creates DeleteClusterNodesNeighFailure with default headers values
func NewDeleteClusterNodesNeighFailure() *DeleteClusterNodesNeighFailure {

	return &DeleteClusterNodesNeighFailure{}
}

// WithPayload adds the payload to the delete cluster nodes neigh failure response
func (o *DeleteClusterNodesNeighFailure) WithPayload(payload models.Error) *DeleteClusterNodesNeighFailure {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the delete cluster nodes neigh failure response
func (o *DeleteClusterNodesNeighFailure) SetPayload(payload models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *DeleteClusterNodesNeighFailure) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(500)
	payload := o.Payload
	if err := producer.Produce(rw, payload); err != nil {
		panic(err) // let the recovery middleware deal with this
	}
}