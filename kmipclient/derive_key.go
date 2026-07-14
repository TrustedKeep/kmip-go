package kmipclient

import (
	"github.com/ovh/kmip-go"
	"github.com/ovh/kmip-go/payloads"
)

// DeriveKey initializes a DeriveKey operation that derives a new object of the given
// type from one or more existing base objects, identified by baseObjectID. The
// derivation method, parameters, and attributes of the derived object are set with
// the returned executor's builder methods.
//
// Example:
//
//	client.DeriveKey(kmip.ObjectTypeSymmetricKey, keyID).
//	    WithDerivationMethod(kmip.DerivationMethodHASH).
//	    WithDerivationParameters(kmip.DerivationParameters{
//	        CryptographicParameters: &kmip.CryptographicParameters{HashingAlgorithm: kmip.HashingAlgorithmSHA_256},
//	    }).
//	    WithAttribute(kmip.AttributeNameCryptographicAlgorithm, kmip.CryptographicAlgorithmAES).
//	    WithAttribute(kmip.AttributeNameCryptographicLength, int32(128)).
//	    Exec()
func (c *Client) DeriveKey(objectType kmip.ObjectType, baseObjectID ...string) ExecDeriveKey {
	req := &payloads.DeriveKeyRequestPayload{
		ObjectType:       objectType,
		UniqueIdentifier: baseObjectID,
	}
	return ExecDeriveKey{
		AttributeExecutor[*payloads.DeriveKeyRequestPayload, *payloads.DeriveKeyResponsePayload, ExecDeriveKey]{
			Executor[*payloads.DeriveKeyRequestPayload, *payloads.DeriveKeyResponsePayload]{
				client: c,
				req:    req,
			},
			func(drp **payloads.DeriveKeyRequestPayload) *[]kmip.Attribute {
				return attrTarget(c, &(*drp).TemplateAttribute, &(*drp).Attributes)
			},
			func(ae AttributeExecutor[*payloads.DeriveKeyRequestPayload, *payloads.DeriveKeyResponsePayload, ExecDeriveKey]) ExecDeriveKey {
				return ExecDeriveKey{ae}
			},
		},
	}
}

// ExecDeriveKey is a specialized executor for the DeriveKey operation. It provides the
// attribute-building helpers for the derived object plus the derivation configuration.
type ExecDeriveKey struct {
	AttributeExecutor[*payloads.DeriveKeyRequestPayload, *payloads.DeriveKeyResponsePayload, ExecDeriveKey]
}

// WithDerivationMethod sets the derivation method for the operation.
func (ex ExecDeriveKey) WithDerivationMethod(method kmip.DerivationMethod) ExecDeriveKey {
	ex.req.DerivationMethod = method
	return ex
}

// WithDerivationParameters sets the derivation parameters for the operation.
func (ex ExecDeriveKey) WithDerivationParameters(params kmip.DerivationParameters) ExecDeriveKey {
	ex.req.DerivationParameters = params
	return ex
}
