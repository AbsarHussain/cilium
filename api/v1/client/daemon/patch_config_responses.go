// Code generated by go-swagger; DO NOT EDIT.

// Copyright Authors of Cilium
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

// PatchConfigReader is a Reader for the PatchConfig structure.
type PatchConfigReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *PatchConfigReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewPatchConfigOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 400:
		result := NewPatchConfigBadRequest()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 500:
		result := NewPatchConfigFailure()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	default:
		return nil, runtime.NewAPIError("response status code does not match any response statuses defined for this endpoint in the swagger spec", response, response.Code())
	}
}

// NewPatchConfigOK creates a PatchConfigOK with default headers values
func NewPatchConfigOK() *PatchConfigOK {
	return &PatchConfigOK{}
}

/*
PatchConfigOK handles this case with default header values.

Success
*/
type PatchConfigOK struct {
}

func (o *PatchConfigOK) Error() string {
	return fmt.Sprintf("[PATCH /config][%d] patchConfigOK ", 200)
}

func (o *PatchConfigOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	return nil
}

// NewPatchConfigBadRequest creates a PatchConfigBadRequest with default headers values
func NewPatchConfigBadRequest() *PatchConfigBadRequest {
	return &PatchConfigBadRequest{}
}

/*
PatchConfigBadRequest handles this case with default header values.

Bad configuration parameters
*/
type PatchConfigBadRequest struct {
	Payload models.Error
}

func (o *PatchConfigBadRequest) Error() string {
	return fmt.Sprintf("[PATCH /config][%d] patchConfigBadRequest  %+v", 400, o.Payload)
}

func (o *PatchConfigBadRequest) GetPayload() models.Error {
	return o.Payload
}

func (o *PatchConfigBadRequest) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response payload
	if err := consumer.Consume(response.Body(), &o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewPatchConfigFailure creates a PatchConfigFailure with default headers values
func NewPatchConfigFailure() *PatchConfigFailure {
	return &PatchConfigFailure{}
}

/*
PatchConfigFailure handles this case with default header values.

Recompilation failed
*/
type PatchConfigFailure struct {
	Payload models.Error
}

func (o *PatchConfigFailure) Error() string {
	return fmt.Sprintf("[PATCH /config][%d] patchConfigFailure  %+v", 500, o.Payload)
}

func (o *PatchConfigFailure) GetPayload() models.Error {
	return o.Payload
}

func (o *PatchConfigFailure) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response payload
	if err := consumer.Consume(response.Body(), &o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
