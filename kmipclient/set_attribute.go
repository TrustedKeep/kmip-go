package kmipclient

import (
	"github.com/ovh/kmip-go"
	"github.com/ovh/kmip-go/payloads"
)

// SetAttribute creates an ExecSetAttribute operation that sets (creating or replacing)
// the value of a single-instance attribute on the object identified by id. It is a
// KMIP 2.0 operation.
//
// Parameters:
//   - id: The unique identifier of the object whose attribute is being set.
//   - name: The name of the attribute to set.
//   - value: The value to set.
func (c *Client) SetAttribute(id string, name kmip.AttributeName, value any) ExecSetAttribute {
	return ExecSetAttribute{
		Executor[*payloads.SetAttributeRequestPayload, *payloads.SetAttributeResponsePayload]{
			client: c,
			req: &payloads.SetAttributeRequestPayload{
				UniqueIdentifier: id,
				NewAttribute: kmip.NewAttribute{
					Attribute: kmip.Attribute{AttributeName: name, AttributeValue: value},
				},
			},
		},
	}
}

// ExecSetAttribute is a specialized executor for the KMIP 2.0 SetAttribute operation.
type ExecSetAttribute struct {
	Executor[*payloads.SetAttributeRequestPayload, *payloads.SetAttributeResponsePayload]
}
