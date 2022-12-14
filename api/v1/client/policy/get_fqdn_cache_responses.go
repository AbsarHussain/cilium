// Code generated by go-swagger; DO NOT EDIT.

// Copyright Authors of Cilium
// SPDX-License-Identifier: Apache-2.0

package policy

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"

	"github.com/cilium/cilium/api/v1/models"
)

// GetFqdnCacheReader is a Reader for the GetFqdnCache structure.
type GetFqdnCacheReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *GetFqdnCacheReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewGetFqdnCacheOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 400:
		result := NewGetFqdnCacheBadRequest()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 404:
		result := NewGetFqdnCacheNotFound()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	default:
		return nil, runtime.NewAPIError("response status code does not match any response statuses defined for this endpoint in the swagger spec", response, response.Code())
	}
}

// NewGetFqdnCacheOK creates a GetFqdnCacheOK with default headers values
func NewGetFqdnCacheOK() *GetFqdnCacheOK {
	return &GetFqdnCacheOK{}
}

/*
GetFqdnCacheOK handles this case with default header values.

Success
*/
type GetFqdnCacheOK struct {
	Payload []*models.DNSLookup
}

func (o *GetFqdnCacheOK) Error() string {
	return fmt.Sprintf("[GET /fqdn/cache][%d] getFqdnCacheOK  %+v", 200, o.Payload)
}

func (o *GetFqdnCacheOK) GetPayload() []*models.DNSLookup {
	return o.Payload
}

func (o *GetFqdnCacheOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response payload
	if err := consumer.Consume(response.Body(), &o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetFqdnCacheBadRequest creates a GetFqdnCacheBadRequest with default headers values
func NewGetFqdnCacheBadRequest() *GetFqdnCacheBadRequest {
	return &GetFqdnCacheBadRequest{}
}

/*
GetFqdnCacheBadRequest handles this case with default header values.

Invalid request (error parsing parameters)
*/
type GetFqdnCacheBadRequest struct {
	Payload models.Error
}

func (o *GetFqdnCacheBadRequest) Error() string {
	return fmt.Sprintf("[GET /fqdn/cache][%d] getFqdnCacheBadRequest  %+v", 400, o.Payload)
}

func (o *GetFqdnCacheBadRequest) GetPayload() models.Error {
	return o.Payload
}

func (o *GetFqdnCacheBadRequest) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response payload
	if err := consumer.Consume(response.Body(), &o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetFqdnCacheNotFound creates a GetFqdnCacheNotFound with default headers values
func NewGetFqdnCacheNotFound() *GetFqdnCacheNotFound {
	return &GetFqdnCacheNotFound{}
}

/*
GetFqdnCacheNotFound handles this case with default header values.

No DNS data with provided parameters found
*/
type GetFqdnCacheNotFound struct {
}

func (o *GetFqdnCacheNotFound) Error() string {
	return fmt.Sprintf("[GET /fqdn/cache][%d] getFqdnCacheNotFound ", 404)
}

func (o *GetFqdnCacheNotFound) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	return nil
}
