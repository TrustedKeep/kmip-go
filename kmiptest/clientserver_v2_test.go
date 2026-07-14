package kmiptest

import (
	"context"
	"testing"

	"github.com/ovh/kmip-go"
	"github.com/ovh/kmip-go/kmipclient"
	"github.com/ovh/kmip-go/kmipserver"
	"github.com/ovh/kmip-go/payloads"

	"github.com/stretchr/testify/require"
)

// TestClientServerV2Attributes verifies end-to-end that a KMIP 2.0 client routes
// object attributes through the 2.0 Attributes structure (not the v1.x
// Template-Attribute), that the server (which now advertises 2.0 by default)
// negotiates 2.0, and that the attributes arrive intact.
func TestClientServerV2Attributes(t *testing.T) {
	var got *payloads.CreateRequestPayload
	exec := kmipserver.NewBatchExecutor()
	exec.Route(kmip.OperationCreate, kmipserver.HandleFunc(
		func(_ context.Context, req *payloads.CreateRequestPayload) (*payloads.CreateResponsePayload, error) {
			got = req
			return &payloads.CreateResponsePayload{
				ObjectType:       req.ObjectType,
				UniqueIdentifier: "v2-key-id",
			}, nil
		}))

	addr, ca := NewServer(t, exec)
	client, err := kmipclient.Dial(addr,
		kmipclient.WithRootCAPem([]byte(ca)),
		kmipclient.EnforceVersion(kmip.V2_0),
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = client.Close() })

	require.Equal(t, kmip.V2_0, client.Version())

	resp, err := client.Create().
		AES(128, kmip.CryptographicUsageEncrypt|kmip.CryptographicUsageDecrypt).
		WithName("v2-key").
		Exec()
	require.NoError(t, err)
	require.Equal(t, "v2-key-id", resp.UniqueIdentifier)

	// The attributes must have travelled in the 2.0 Attributes structure, not the
	// deprecated v1.x Template-Attribute.
	require.NotNil(t, got)
	require.Empty(t, got.TemplateAttribute.Attribute, "v1.x Template-Attribute must be empty at 2.0")
	require.NotEmpty(t, got.Attributes, "2.0 Attributes must carry the attributes")

	assert := require.New(t)
	assert.Equal(kmip.CryptographicAlgorithmAES, got.Attributes.Get(kmip.AttributeNameCryptographicAlgorithm))
	assert.Equal(int32(128), got.Attributes.Get(kmip.AttributeNameCryptographicLength))
	name, ok := got.Attributes.Get(kmip.AttributeNameName).(kmip.Name)
	assert.True(ok)
	assert.Equal("v2-key", name.NameValue)
}

// TestClientServerV2Locate verifies a 2.0 Locate routes its search attributes
// through the Attributes structure end-to-end.
func TestClientServerV2Locate(t *testing.T) {
	var got *payloads.LocateRequestPayload
	exec := kmipserver.NewBatchExecutor()
	exec.Route(kmip.OperationLocate, kmipserver.HandleFunc(
		func(_ context.Context, req *payloads.LocateRequestPayload) (*payloads.LocateResponsePayload, error) {
			got = req
			return &payloads.LocateResponsePayload{UniqueIdentifier: []string{"found-1"}}, nil
		}))

	addr, ca := NewServer(t, exec)
	client, err := kmipclient.Dial(addr,
		kmipclient.WithRootCAPem([]byte(ca)),
		kmipclient.EnforceVersion(kmip.V2_0),
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = client.Close() })

	resp, err := client.Locate().
		WithAttribute(kmip.AttributeNameObjectType, kmip.ObjectTypeSymmetricKey).
		Exec()
	require.NoError(t, err)
	require.Equal(t, []string{"found-1"}, resp.UniqueIdentifier)

	require.NotNil(t, got)
	require.Empty(t, got.Attribute, "v1.x Attribute list must be empty at 2.0")
	require.NotEmpty(t, got.Attributes)
	require.Equal(t, kmip.ObjectTypeSymmetricKey, got.Attributes.Get(kmip.AttributeNameObjectType))
}

// TestClientServerV2AttributeOps verifies the KMIP 2.0 attribute-mutation and query
// client builders route through the 2.0 New Attribute / Attribute Reference forms.
func TestClientServerV2AttributeOps(t *testing.T) {
	var add *payloads.AddAttributeRequestPayload
	var set *payloads.SetAttributeRequestPayload
	var del *payloads.DeleteAttributeRequestPayload
	var geta *payloads.GetAttributesRequestPayload

	exec := kmipserver.NewBatchExecutor()
	exec.Route(kmip.OperationAddAttribute, kmipserver.HandleFunc(
		func(_ context.Context, req *payloads.AddAttributeRequestPayload) (*payloads.AddAttributeResponsePayload, error) {
			add = req
			return &payloads.AddAttributeResponsePayload{UniqueIdentifier: req.UniqueIdentifier}, nil
		}))
	exec.Route(kmip.OperationSetAttribute, kmipserver.HandleFunc(
		func(_ context.Context, req *payloads.SetAttributeRequestPayload) (*payloads.SetAttributeResponsePayload, error) {
			set = req
			return &payloads.SetAttributeResponsePayload{UniqueIdentifier: req.UniqueIdentifier}, nil
		}))
	exec.Route(kmip.OperationDeleteAttribute, kmipserver.HandleFunc(
		func(_ context.Context, req *payloads.DeleteAttributeRequestPayload) (*payloads.DeleteAttributeResponsePayload, error) {
			del = req
			return &payloads.DeleteAttributeResponsePayload{UniqueIdentifier: req.UniqueIdentifier}, nil
		}))
	exec.Route(kmip.OperationGetAttributes, kmipserver.HandleFunc(
		func(_ context.Context, req *payloads.GetAttributesRequestPayload) (*payloads.GetAttributesResponsePayload, error) {
			geta = req
			return &payloads.GetAttributesResponsePayload{
				UniqueIdentifier: req.UniqueIdentifier,
				Attributes:       kmip.Attributes{{AttributeName: kmip.AttributeNameSensitive, AttributeValue: true}},
			}, nil
		}))

	addr, ca := NewServer(t, exec)
	client, err := kmipclient.Dial(addr,
		kmipclient.WithRootCAPem([]byte(ca)),
		kmipclient.EnforceVersion(kmip.V2_0),
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = client.Close() })

	_, err = client.AddAttribute("obj-1", kmip.AttributeNameSensitive, true).Exec()
	require.NoError(t, err)
	require.Empty(t, add.Attribute.AttributeName, "v1.x Attribute must be empty at 2.0")
	require.Equal(t, kmip.AttributeNameSensitive, add.NewAttribute.AttributeName)

	_, err = client.SetAttribute("obj-1", kmip.AttributeNameSensitive, true).Exec()
	require.NoError(t, err)
	require.Equal(t, kmip.AttributeNameSensitive, set.NewAttribute.AttributeName)

	_, err = client.DeleteAttribute("obj-1", kmip.AttributeNameSensitive).Exec()
	require.NoError(t, err)
	require.Empty(t, del.AttributeName, "v1.x AttributeName must be empty at 2.0")
	require.NotNil(t, del.AttributeReference)
	require.Equal(t, kmip.TagSensitive, del.AttributeReference.Tag)

	resp, err := client.GetAttributes("obj-1", kmip.AttributeNameSensitive).Exec()
	require.NoError(t, err)
	require.Empty(t, geta.AttributeName, "v1.x AttributeName must be empty at 2.0")
	require.Len(t, geta.AttributeReference, 1)
	require.Equal(t, kmip.TagSensitive, geta.AttributeReference[0].Tag)
	require.Equal(t, true, resp.Attributes.Get(kmip.AttributeNameSensitive))
}

// TestClientServerV2DeriveKey verifies the DeriveKey client builder routes the derived
// object's attributes through the 2.0 Attributes structure and carries the derivation
// method and parameters.
func TestClientServerV2DeriveKey(t *testing.T) {
	var got *payloads.DeriveKeyRequestPayload
	exec := kmipserver.NewBatchExecutor()
	exec.Route(kmip.OperationDeriveKey, kmipserver.HandleFunc(
		func(_ context.Context, req *payloads.DeriveKeyRequestPayload) (*payloads.DeriveKeyResponsePayload, error) {
			got = req
			return &payloads.DeriveKeyResponsePayload{UniqueIdentifier: "derived-1"}, nil
		}))

	addr, ca := NewServer(t, exec)
	client, err := kmipclient.Dial(addr,
		kmipclient.WithRootCAPem([]byte(ca)),
		kmipclient.EnforceVersion(kmip.V2_0),
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = client.Close() })

	resp, err := client.DeriveKey(kmip.ObjectTypeSymmetricKey, "base-key").
		WithDerivationMethod(kmip.DerivationMethodHASH).
		WithDerivationParameters(kmip.DerivationParameters{
			CryptographicParameters: &kmip.CryptographicParameters{HashingAlgorithm: kmip.HashingAlgorithmSHA_256},
		}).
		WithAttribute(kmip.AttributeNameCryptographicAlgorithm, kmip.CryptographicAlgorithmAES).
		WithAttribute(kmip.AttributeNameCryptographicLength, int32(128)).
		Exec()
	require.NoError(t, err)
	require.Equal(t, "derived-1", resp.UniqueIdentifier)

	require.NotNil(t, got)
	require.Equal(t, []string{"base-key"}, got.UniqueIdentifier)
	require.Equal(t, kmip.DerivationMethodHASH, got.DerivationMethod)
	require.Empty(t, got.TemplateAttribute.Attribute, "v1.x Template-Attribute must be empty at 2.0")
	require.Equal(t, kmip.CryptographicAlgorithmAES, got.Attributes.Get(kmip.AttributeNameCryptographicAlgorithm))
}

// TestClientServerV2Import verifies the Import client builder carries the object, a
// top-level ObjectType (2.0), and attributes via the 2.0 Attributes structure.
func TestClientServerV2Import(t *testing.T) {
	var got *payloads.ImportRequestPayload
	exec := kmipserver.NewBatchExecutor()
	exec.Route(kmip.OperationImport, kmipserver.HandleFunc(
		func(_ context.Context, req *payloads.ImportRequestPayload) (*payloads.ImportResponsePayload, error) {
			got = req
			return &payloads.ImportResponsePayload{UniqueIdentifier: req.UniqueIdentifier}, nil
		}))

	addr, ca := NewServer(t, exec)
	client, err := kmipclient.Dial(addr,
		kmipclient.WithRootCAPem([]byte(ca)),
		kmipclient.EnforceVersion(kmip.V2_0),
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = client.Close() })

	secret := []byte("0123456789abcdef")
	obj := &kmip.SymmetricKey{
		KeyBlock: kmip.KeyBlock{
			KeyFormatType:          kmip.KeyFormatTypeRaw,
			CryptographicAlgorithm: kmip.CryptographicAlgorithmAES,
			CryptographicLength:    128,
			KeyValue:               &kmip.KeyValue{Plain: &kmip.PlainKeyValue{KeyMaterial: kmip.KeyMaterial{Bytes: &secret}}},
		},
	}

	_, err = client.Import("imp-1", obj).
		WithReplaceExisting(true).
		WithAttribute(kmip.AttributeNameCryptographicUsageMask, kmip.CryptographicUsageEncrypt).
		Exec()
	require.NoError(t, err)

	require.NotNil(t, got)
	require.Equal(t, "imp-1", got.UniqueIdentifier)
	require.Equal(t, kmip.ObjectTypeSymmetricKey, got.ObjectType, "2.0 carries ObjectType as a top-level field")
	require.Empty(t, got.Attribute, "v1.x Attribute list must be empty at 2.0")
	require.Equal(t, kmip.CryptographicUsageEncrypt, got.Attributes.Get(kmip.AttributeNameCryptographicUsageMask))
	require.NotNil(t, got.Object)
}
