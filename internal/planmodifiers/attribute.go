package planmodifiers

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

type requiresReplaceOnValueChange struct{}

func RequiresReplaceOnValueChange() planmodifier.String {
	return requiresReplaceOnValueChange{}
}

func (r requiresReplaceOnValueChange) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, res *planmodifier.StringResponse) {
	if req.State.Raw.IsNull() {
		// Creating resource, nothing to do here
		return
	}

	if req.Plan.Raw.IsNull() {
		// Deleting resource, nothing to do here
	}

	res.RequiresReplace = true
}

const description = "If the value of this attribute changes, Terraform will destroy and recreate the resource."

func (r requiresReplaceOnValueChange) Description(ctx context.Context) string {
	return description
}

func (r requiresReplaceOnValueChange) MarkdownDescription(ctx context.Context) string {
	return description
}
