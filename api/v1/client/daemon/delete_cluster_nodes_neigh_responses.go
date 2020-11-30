// Code generated by go-swagger; DO NOT EDIT.

// Copyright 2017-2020 Authors of Cilium
// SPDX-License-Identifier: Apache-2.0

package daemon

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"

	"github.com/cilium/cilium/api/v1/models"
)

// DeleteClusterNodesNeighReader is a Reader for the DeleteClusterNodesNeigh structure.
type DeleteClusterNodesNeighReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *DeleteClusterNodesNeighReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewDeleteClusterNodesNeighOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 404:
		result := NewDeleteClusterNodesNeighNotFound()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 500:
		result := NewDeleteClusterNodesNeighFailure()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	default:
		return nil, runtime.NewAPIError("response status code does not match any response statuses defined for this endpoint in the swagger spec", response, response.Code())
	}
}

// NewDeleteClusterNodesNeighOK creates a DeleteClusterNodesNeighOK with default headers values
func NewDeleteClusterNodesNeighOK() *DeleteClusterNodesNeighOK {
	return &DeleteClusterNodesNeighOK{}
}

/*DeleteClusterNodesNeighOK handles this case with default header values.

Success
*/
type DeleteClusterNodesNeighOK struct {
}

func (o *DeleteClusterNodesNeighOK) Error() string {
	return fmt.Sprintf("[DELETE /cluster/nodes/neigh][%d] deleteClusterNodesNeighOK ", 200)
}

func (o *DeleteClusterNodesNeighOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	return nil
}

// NewDeleteClusterNodesNeighNotFound creates a DeleteClusterNodesNeighNotFound with default headers values
func NewDeleteClusterNodesNeighNotFound() *DeleteClusterNodesNeighNotFound {
	return &DeleteClusterNodesNeighNotFound{}
}

/*DeleteClusterNodesNeighNotFound handles this case with default header values.

Node neighbor not found
*/
type DeleteClusterNodesNeighNotFound struct {
}

func (o *DeleteClusterNodesNeighNotFound) Error() string {
	return fmt.Sprintf("[DELETE /cluster/nodes/neigh][%d] deleteClusterNodesNeighNotFound ", 404)
}

func (o *DeleteClusterNodesNeighNotFound) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	return nil
}

// NewDeleteClusterNodesNeighFailure creates a DeleteClusterNodesNeighFailure with default headers values
func NewDeleteClusterNodesNeighFailure() *DeleteClusterNodesNeighFailure {
	return &DeleteClusterNodesNeighFailure{}
}

/*DeleteClusterNodesNeighFailure handles this case with default header values.

Error while deleting node neighbor
*/
type DeleteClusterNodesNeighFailure struct {
	Payload models.Error
}

func (o *DeleteClusterNodesNeighFailure) Error() string {
	return fmt.Sprintf("[DELETE /cluster/nodes/neigh][%d] deleteClusterNodesNeighFailure  %+v", 500, o.Payload)
}

func (o *DeleteClusterNodesNeighFailure) GetPayload() models.Error {
	return o.Payload
}

func (o *DeleteClusterNodesNeighFailure) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response payload
	if err := consumer.Consume(response.Body(), &o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
