// Copyright (c) HashiCorp, Inc.

package provider

import (
	"context"
	"time"

	"github.com/ProtonMail/gopenpgp/v3/constants"
	"github.com/ProtonMail/gopenpgp/v3/crypto"
	"github.com/ProtonMail/gopenpgp/v3/profile"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type keyPairResource struct{}

func NewKeyPairResource() resource.Resource {
	return &keyPairResource{}
}

func (k *keyPairResource) Metadata(ctx context.Context, req resource.MetadataRequest, res *resource.MetadataResponse) {
	res.TypeName = req.ProviderTypeName + "_key_pair"
}

func (k *keyPairResource) Schema(ctx context.Context, req resource.SchemaRequest, res *resource.SchemaResponse) {
	res.Schema = schema.Schema{

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "ID of the key-pair",
			},
			"kind": schema.StringAttribute{
				Computed:            true,
				Optional:            true,
				MarkdownDescription: "Which kind of key to use - `default` will result in an **curve25519 v4** key, `rfc4880` will result in a **rsa 4096 bits** key and `rf9580` will result in a **curve448 v6** key",
				Default:             stringdefault.StaticString("default"),
				Validators: []validator.String{
					stringvalidator.OneOf("default", "rfc4880", "rfc9580"),
				},
			},
			"passphrase": schema.StringAttribute{
				Sensitive:   true,
				Description: "Passphrase of the PGP Key",
				Required:    true,
			},
			"identity": schema.SingleNestedAttribute{
				Description: "User ID of the key",
				Required:    true,

				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{
						Required:    true,
						Description: "Real name",
					},
					"email": schema.StringAttribute{
						Required:    true,
						Description: "Email address",
					},
				},
			},
			"fingerprint": schema.StringAttribute{
				Computed:    true,
				Description: "Fingerprint of the key",
			},
			"private_key": schema.StringAttribute{
				Computed:    true,
				Sensitive:   true,
				Description: "Private key in Armory format",
			},
			"public_key": schema.StringAttribute{
				Computed:    true,
				Description: "Public key in Armory Format",
			},
			"expires_at": schema.StringAttribute{
				Optional:    true,
				Description: "When the key should expire",
			},
		},
	}
}

type keyPairModel struct {
	Id         types.String `tfsdk:"id"`
	Kind       types.String `tfsdk:"kind"`
	Passphrase types.String `tfsdk:"passphrase"`
	Identity   struct {
		Name  types.String `tfsdk:"name"`
		Email types.String `tfsdk:"email"`
	} `tfsdk:"identity"`
	Fingerprint types.String `tfsdk:"fingerprint"`
	PrivateKey  types.String `tfsdk:"private_key"`
	PublicKey   types.String `tfsdk:"public_key"`
	ExpiresAt   types.String `tfsdk:"expires_at"`
}

func (k *keyPairResource) getGenerator(kind string) *profile.Custom {
	switch kind {
	case "default":
		return profile.Default()
	case "rfc4880":
		return profile.RFC4880()
	case "rfc9580":
		return profile.RFC9580()
	default:
		panic("Not a valid PGP key kind")
	}
}

func (k *keyPairResource) Create(ctx context.Context, req resource.CreateRequest, res *resource.CreateResponse) {
	var model keyPairModel

	res.Diagnostics.Append(req.Plan.Get(ctx, &model)...)

	if res.Diagnostics.HasError() {
		return
	}

	pgp := crypto.PGPWithProfile(k.getGenerator(model.Kind.ValueString()))
	builder := pgp.KeyGeneration().AddUserId(model.Identity.Name.ValueString(), model.Identity.Email.ValueString())

	if !model.ExpiresAt.IsNull() {
		expires_at, err := time.Parse(time.RFC3339, model.ExpiresAt.ValueString())

		if err != nil {
			res.Diagnostics.AddError("Could not parse date specified in `expires_at`", err.Error())
			return
		}

		builder = builder.Lifetime(int32(time.Until(expires_at).Seconds()))
	}

	key, err := builder.New().GenerateKeyWithSecurity(constants.HighSecurity)

	if err != nil {
		res.Diagnostics.AddError("Could not create PGP key", err.Error())
	}

	key, err = pgp.LockKey(key, []byte(model.Passphrase.ValueString()))

	if err != nil {
		res.Diagnostics.AddError("Could not add passphrase to PGP key", err.Error())
	}

	privateKey, err := key.Armor()

	if err != nil {
		res.Diagnostics.AddError("Could not generate private key", err.Error())
	}

	publicKey, err := key.GetArmoredPublicKey()

	if err != nil {
		res.Diagnostics.AddError("Could not generate public key", err.Error())
	}

	model.Id = types.StringValue(key.GetHexKeyID())
	model.Fingerprint = types.StringValue(key.GetFingerprint())
	model.PublicKey = types.StringValue(publicKey)
	model.PrivateKey = types.StringValue(privateKey)

	res.Diagnostics.Append(res.State.Set(ctx, &model)...)
}

func (k *keyPairResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Nothing to do here.
}

func (k *keyPairResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Nothing to do here.
}

func (k *keyPairResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Nothing to do here.
}
