# KMIP 1.x → 2.0 Handler Migration Guide

> **Audience:** product teams (e.g. TrustedHSM) whose KMIP interface is built on this
> library and who implement the **server-side operation handlers** registered with
> `kmipserver.BatchExecutor.Route(...)` / `kmipserver.HandleFunc(...)`.
>
> **Scope:** this guide covers what your *handler / business-logic* layer must change to
> support KMIP 2.0. It does **not** cover the wire format — the library handles TTLV 2.0
> encoding/decoding, the new `DateTimeExtended` type, all 2.0 tags/enums, version
> negotiation, and per-version attribute placement for you.

---

## 1. Division of labor

| Concern | Handled by the library | Your handlers must do |
|---|---|---|
| TTLV 2.0 wire encoding/decoding (binary/XML/JSON) | ✅ | — |
| New 2.0 tags, `DateTimeExtended`, 2.0 enums/bitmasks | ✅ | — |
| Version negotiation (`DiscoverVersions`), `set-version` threading | ✅ | — |
| Routing the correct wire form of attributes per version | ✅ | — |
| Reading attributes from the version-appropriate struct field | — | ✅ |
| Producing 2.0-shaped responses | — | ✅ |
| Implementing new 2.0 operations you choose to support | — | ✅ |
| Returning appropriate 2.0 result reasons | — | ✅ |

The net effect: your handlers keep working with Go structs (`payloads.*`), but a few
payload fields moved/changed between versions, and a handful of operations are new.

---

## 2. Enabling 2.0 on the server

`kmipserver.NewBatchExecutor()` already advertises 2.0 by default (its
`defaultSupportedVersion` includes `kmip.V2_0`). To pin an explicit set:

```go
exec := kmipserver.NewBatchExecutor()
exec.SetSupportedProtocolVersions(kmip.V2_0, kmip.V1_4, kmip.V1_3, kmip.V1_2, kmip.V1_1, kmip.V1_0)
```

A handler can read the negotiated version of the request it is serving:

```go
ver := kmipserver.GetProtocolVersion(ctx) // kmip.ProtocolVersion
isV2 := ttlv.CompareVersions(ver, kmip.V2_0) >= 0
```

Use this to shape responses (section 4).

---

## 3. The core change: the attribute model

This is the bulk of the migration work.

- **1.x:** attributes arrive in `payload.TemplateAttribute.Attribute` (a
  `[]kmip.Attribute`, name-on-the-wire).
- **2.0:** attributes arrive in `payload.Attributes` (also `[]kmip.Attribute` in this
  library, but a **different field**, with each attribute under its own dedicated tag —
  no attribute-name strings on the wire).

Because both are `[]kmip.Attribute`, your business logic barely changes — you just need
to read the field that is populated for the request's version. Add one helper and route
all attribute reads through it. For example, for Create:

```go
func createAttrs(p *payloads.CreateRequestPayload) []kmip.Attribute {
    if len(p.Attributes) > 0 {
        return p.Attributes            // KMIP 2.0
    }
    return p.TemplateAttribute.Attribute // KMIP 1.x
}
```

Apply the same pattern to every attribute-bearing request:

| Payload | v1.x field | v2.0 field |
|---|---|---|
| `CreateRequestPayload` | `TemplateAttribute.Attribute` | `Attributes` |
| `RegisterRequestPayload` | `TemplateAttribute.Attribute` | `Attributes` |
| `DeriveKeyRequestPayload` | `TemplateAttribute.Attribute` | `Attributes` |
| `RekeyRequestPayload` | `TemplateAttribute.Attribute` | `Attributes` |
| `LocateRequestPayload` | `Attribute` | `Attributes` |
| `ImportRequestPayload` | `Attribute` | `Attributes` |
| `CreateKeyPairRequestPayload` | `Common/Private/PublicKeyTemplateAttribute.Attribute` | `Common/Private/PublicKeyAttributes` |

Reading a typed value out of an attribute list is unchanged; `Attributes` also offers a
convenience getter:

```go
alg := attrs.Get(kmip.AttributeNameCryptographicAlgorithm) // any; nil if absent
```

### Vendor / custom attributes

- **1.x:** custom attributes were identified by an `x-`/`y-` name prefix.
- **2.0:** they use a **Vendor Attribute** structure. In this library, a vendor
  attribute is a `kmip.Attribute` whose `VendorIdentification` field is non-empty
  (with `AttributeName` and `AttributeValue`). Handle those explicitly where you support
  custom attributes.

---

## 4. Responses no longer echo request attributes

KMIP 2.0 **removed** the practice of echoing server-set attributes in operation
responses. Populate response attribute fields **only for 1.x**:

```go
resp := &payloads.CreateResponsePayload{
    ObjectType:       req.ObjectType,
    UniqueIdentifier: newID,
}
if ttlv.CompareVersions(kmipserver.GetProtocolVersion(ctx), kmip.V2_0) < 0 {
    resp.Attributes = &kmip.TemplateAttribute{ /* server-set attrs (1.x only) */ }
}
return resp, nil
```

The library gates the unused field off the wire regardless, but keeping response
construction version-correct avoids confusion. Affected responses: Create, Register,
CreateKeyPair, DeriveKey, ReKey, Add/Modify/Delete Attribute (2.0 returns only the
Unique Identifier).

---

## 5. Attribute operations — shape changes

The attribute-mutation and query operations changed shape. The **Attribute Index**
concept is **gone** in 2.0.

### AddAttribute / ModifyAttribute

- **1.x:** `req.Attribute` (a `kmip.Attribute`).
- **2.0:** `req.NewAttribute` (a `kmip.NewAttribute`; Modify may also carry an optional
  `req.CurrentAttribute *kmip.CurrentAttribute`). Both embed `kmip.Attribute`, so fields
  are promoted:

```go
func handleAdd(ctx context.Context, req *payloads.AddAttributeRequestPayload) (*payloads.AddAttributeResponsePayload, error) {
    var attr kmip.Attribute
    if ttlv.CompareVersions(kmipserver.GetProtocolVersion(ctx), kmip.V2_0) >= 0 {
        attr = req.NewAttribute.Attribute // 2.0
    } else {
        attr = req.Attribute              // 1.x
    }
    // ... apply attr to the object identified by req.UniqueIdentifier ...
    return &payloads.AddAttributeResponsePayload{UniqueIdentifier: req.UniqueIdentifier}, nil
}
```

### DeleteAttribute

- **1.x:** `req.AttributeName` (+ optional `req.AttributeIndex`).
- **2.0:** `req.AttributeReference *kmip.AttributeReference` (or `req.CurrentAttribute`).
  An `AttributeReference` is either an enumeration form (`ref.Tag` = the attribute's
  dedicated tag) or a vendor form (`ref.VendorIdentification` + `ref.AttributeName`).
  Compare `ref.Tag` against the `kmip.Tag*` constants (e.g. `kmip.TagSensitive`) to
  identify the attribute.

### GetAttributes

- **1.x request:** `req.AttributeName []kmip.AttributeName`.
- **2.0 request:** `req.AttributeReference []kmip.AttributeReference` (empty ⇒ return
  all attributes).
- **Response:** populate `resp.Attribute` ([]Attribute) for 1.x, `resp.Attributes`
  (`kmip.Attributes`) for 2.0.

### GetAttributeList

- **Response:** `resp.AttributeName` for 1.x, `resp.AttributeReference` for 2.0.
  Build a reference from a name with `kmip.AttributeReferenceFor(name)`.

### SetAttribute — **new operation**

New in 2.0: create-or-replace a single-instance attribute. Register a handler if you
want it:

```go
exec.Route(kmip.OperationSetAttribute, kmipserver.HandleFunc(
    func(ctx context.Context, req *payloads.SetAttributeRequestPayload) (*payloads.SetAttributeResponsePayload, error) {
        attr := req.NewAttribute.Attribute
        // ... set attr on req.UniqueIdentifier ...
        return &payloads.SetAttributeResponsePayload{UniqueIdentifier: req.UniqueIdentifier}, nil
    }))
```

---

## 6. Other new 2.0 operations (optional)

Decide per your product/profile which to support; the library routes them, you supply
handlers:

- **Login / Logout** — session establishment. `Login` returns a **Ticket**; the client
  presents it on later requests via the **Ticket credential**. `Logout` ends the session.
- **DelegatedLogin** — issues a ticket scoped to a set of **Rights** (specific
  operations/objects).
- **Interop** — test-harness begin/end/reset (used by conformance vectors).
- **Log**, **AdjustAttribute**, **SetEndpointRole**, **PKCS_11**, **ReProvision** —
  optional; implement only if needed.

### Authentication additions

New `kmip.CredentialType` values: `Ticket` (session ticket from Login),
`OneTimePassword`, `HashedPassword` (in addition to 1.x UsernameAndPassword / Device /
Attestation). The Ticket credential's value is a `kmip.CredentialValueTicket{Ticket}`.
If TrustedHSM implements sessions, validate the presented ticket in a middleware or at
the start of each handler via `kmipserver.GetRequestHeader(ctx).Authentication`.

---

## 7. New attributes / enums / result reasons to be aware of

- **New attributes** you may need to return: `ShortUniqueIdentifier`,
  `ProtectionStorageMask` (a bitmask describing acceptable storage — software/hardware/
  on-system/etc., relevant to an HSM), certificate DN component attributes
  (`CertificateIssuerC/CN/O`, `CertificateSubjectC/CN/O`), and `KeyFormatType` as an
  attribute.
- **New enums / values:** `NISTKeyType`, `KeyFormatTypePKCS_12`, and the
  **Certificate Request** object type.
- **New `ResultReason` codes** — return these where appropriate instead of a generic
  failure: `AttributeReadOnly`, `MultiValuedAttribute`, `AttributeNotFound`,
  `AttributeSingleValued`, `IncompatibleCryptographicUsageMask`, `PermissionDenied`,
  `InvalidAttribute`, `InvalidAttributeValue`, `ObjectNotFound`, etc. Use
  `kmipserver.Errorf(kmip.ResultReason<X>, "msg")` to return a structured KMIP error.
- **`DateTimeExtended`** (`ttlv.DateTimeExtended`) — microsecond-precision timestamps
  are now possible on the wire.

---

## 8. Removed / deprecated in 2.0 (stop relying on these)

- **Template managed object** and the **Template-Attribute** wrapper — removed.
- **Operation Policy Name** — removed.
- **Custom Attribute** — replaced by Vendor Attribute.
- **Attribute Index** — removed.
- Deprecated certificate attributes, **PGP** certificate types, pre-2.0 profiles, and
  some crypto-usage-mask bits (Translate/Cryptogram) reserved.
- **Still present in 2.0** (do not remove): **Batch Count** and **Unique Batch Item ID**
  (these are removed in 3.0, not 2.0).

---

## 9. Transport

KMIP 2.0 mandates **TLS 1.2+**. The library already enforces `MinVersion: TLS 1.2`, so
no action is required.

---

## 10. Testing & validation

- The repository already runs the **official OASIS KMIP 2.0 Test Cases** (in
  `kmiptest/testdata/v2.0/`) as a parse/re-encode round-trip in `go test ./...`
  (`TestParseAndMarshalOasisTests`). This validates wire correctness of the library, not
  your handler logic.
- For handler validation, add in-process client↔server tests (see
  `kmiptest/clientserver_v2_test.go` for the pattern) that drive your handlers with a
  2.0 client (`kmipclient.Dial(..., kmipclient.EnforceVersion(kmip.V2_0))`).
- For independent interop, drive your running server with **PyKMIP** configured for
  `KMIPVersion.KMIP_2_0`, or with the **Cosmian KMS** `ckms` client; and/or point this
  library's client at an independent 2.0 server (e.g. Cosmian KMS).

---

## 11. Migration checklist

1. [ ] Pick a target KMIP 2.0 **profile** (e.g. Baseline Server) to scope required ops/attrs.
2. [ ] Add a **version-agnostic attribute accessor** per attribute-bearing request (section 3).
3. [ ] Update **attribute-mutation handlers** (Add/Modify/Delete) to `NewAttribute` /
       `CurrentAttribute` / `AttributeReference`; drop Attribute Index at 2.0.
4. [ ] Update **GetAttributes / GetAttributeList** to `AttributeReference` (request) and
       `Attributes` / `AttributeReference` (response).
5. [ ] Add a **SetAttribute** handler.
6. [ ] Make responses **not echo attributes** at 2.0 (section 4).
7. [ ] Handle **vendor attributes** via `VendorIdentification`.
8. [ ] Decide on **sessions** (Login/Logout/DelegatedLogin + Ticket credential) and any
       other optional new operations.
9. [ ] Return **2.0 result reasons** where applicable.
10. [ ] Add **2.0 handler tests** and run an **independent-client interop** check.

---

## Reference: key library symbols

| Purpose | Symbol |
|---|---|
| Negotiated version in a handler | `kmipserver.GetProtocolVersion(ctx)` |
| Request header (auth, etc.) in a handler | `kmipserver.GetRequestHeader(ctx)` |
| Structured KMIP error | `kmipserver.Errorf(kmip.ResultReason…, "…")` |
| Compare versions | `ttlv.CompareVersions(a, b)` (>= 0 vs `kmip.V2_0` ⇒ 2.0+) |
| 2.0 attribute collection | `kmip.Attributes` (`[]kmip.Attribute`), `.Get(name)` |
| Single-attribute wrappers | `kmip.NewAttribute`, `kmip.CurrentAttribute` (embed `kmip.Attribute`) |
| Attribute reference (enum or vendor) | `kmip.AttributeReference`, `kmip.AttributeReferenceFor(name)` |
| Vendor attribute marker | `kmip.Attribute.VendorIdentification` |
| Register a handler | `exec.Route(op, kmipserver.HandleFunc(fn))` |
