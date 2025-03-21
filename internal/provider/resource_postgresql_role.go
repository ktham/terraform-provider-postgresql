package provider

import (
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/jackc/pgx/v5"
	"strings"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &RoleResource{}
var _ resource.ResourceWithImportState = &RoleResource{}

func NewRoleResource() resource.Resource {
	return &RoleResource{}
}

type RoleResource struct {
	data PostgresqlProviderData
}

type RoleResourceModel struct {
	Oid                    types.Int64  `tfsdk:"oid"`
	Name                   types.String `tfsdk:"name"`
	BypassRowLevelSecurity types.Bool   `tfsdk:"bypass_row_level_security"`
	CanLogin               types.Bool   `tfsdk:"can_login"`
	ConnectionLimit        types.Int32  `tfsdk:"connection_limit"`
	CreateRole             types.Bool   `tfsdk:"create_role"`
	Inherit                types.Bool   `tfsdk:"inherit"`
	Replication            types.Bool   `tfsdk:"replication"`
	Superuser              types.Bool   `tfsdk:"superuser"`
}

func (r *RoleResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_role"
}

func (r *RoleResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Postgresql Role",

		Attributes: map[string]schema.Attribute{
			"oid": schema.Int64Attribute{
				Description: "The object ID of the Postgresql role.",
				Required:    false,
				Optional:    false,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the Postgresql role.",
				Required:    true,
				// The impact of a role re-name can have a problematic impact on downstream dependencies of this role,
				// so we won't support in-place update of the role's name.
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"bypass_row_level_security": schema.BoolAttribute{
				Description: "Determines whether a role bypasses every row-level security (RLS) policy.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"can_login": schema.BoolAttribute{
				Description: "Determines whether a role is allowed to log in",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"connection_limit": schema.Int32Attribute{
				Description: "Specifies how many concurrent connections the role can make. -1 (the default) means no limit.",
				Optional:    true,
				Computed:    true,
				Default:     int32default.StaticInt32(-1),
				Validators: []validator.Int32{
					int32validator.AtLeast(-1),
				},
			},
			"create_role": schema.BoolAttribute{
				Description: "Determines whether the role will be permitted to create, alter, drop, comment on, and change the security label for other roles.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"inherit": schema.BoolAttribute{
				Description: "Determines whether the role inherits privileges from other roles that it's a member of.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"replication": schema.BoolAttribute{
				Description: "Determines whether the role will have permissions to initiate replication.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"superuser": schema.BoolAttribute{
				Description: "Determines whether the new role is a “superuser”, which can override all access restrictions within the database.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
		},
	}
}

func (r *RoleResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	data, ok := req.ProviderData.(PostgresqlProviderData)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected PostgresqlProviderData, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.data = data
}

func (r *RoleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var dataFromPlan RoleResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &dataFromPlan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	txn, err := r.data.DbPool.Begin(ctx)

	if err != nil {
		resp.Diagnostics.AddError("DB Connection Pool Error", fmt.Sprintf("Unable to start a new transaction creating a connection pool, got error: %s", err))
		return
	}

	defer func() {
		if err != nil {
			err := txn.Rollback(ctx)
			if err != nil {
				resp.Diagnostics.AddError("Transaction Rollback Error", fmt.Sprintf("Unable to rollback transaction, got error: %s", err))
			}
		}
	}()

	createRoleSql := fmt.Sprintf("CREATE ROLE %s WITH %s;", dataFromPlan.Name.ValueString(), dataFromPlan.GetOptionsString(r))

	tflog.Info(ctx, createRoleSql)

	if _, err = txn.Exec(ctx, createRoleSql); err != nil {
		resp.Diagnostics.AddError("DB role creation error", fmt.Sprintf("Error executing query '%s', got error: %s", createRoleSql, err))
		return
	}

	var roleOID uint32
	selectOidQuery := fmt.Sprintf("SELECT oid FROM pg_roles WHERE rolname = '%s'", dataFromPlan.Name.ValueString())

	err = txn.QueryRow(ctx, selectOidQuery).Scan(&roleOID)
	if err != nil {
		resp.Diagnostics.AddError("Failed to retrieve role OID", fmt.Sprintf("Error retrieving role OID with query `%s`, got error: %s", selectOidQuery, err))
		return
	}

	dataFromPlan.Oid = types.Int64Value(int64(roleOID))

	err = txn.Commit(ctx)
	if err != nil {
		resp.Diagnostics.AddError("DB transaction error", fmt.Sprintf("Error committing DB transaction, got error: %s", err))
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("Successfully created Postgresql Role: %s", dataFromPlan.Name.ValueString()))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &dataFromPlan)...)
}

func (r *RoleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var dataFromState RoleResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &dataFromState)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Query for the actual state of the role from the database
	roleSql := fmt.Sprintf(`
SELECT
    rolbypassrls,
    rolcanlogin,
    rolconnlimit,
    rolcreaterole,
    rolinherit,
    rolname,
    rolreplication,
    rolsuper
FROM 
    pg_roles
WHERE 
    oid = %d;`, dataFromState.Oid.ValueInt64())

	var bypassRowLevelSecurity bool
	var canLogin bool
	var connectionLimit int32
	var createRole bool
	var inherit bool
	var name string
	var replication bool
	var superuser bool

	err := r.data.DbPool.QueryRow(ctx, roleSql).Scan(
		&bypassRowLevelSecurity,
		&canLogin,
		&connectionLimit,
		&createRole,
		&inherit,
		&name,
		&replication,
		&superuser,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			resp.Diagnostics.AddWarning("No results returned", fmt.Sprintf("The Postgres role couldn't be found. role: %s", dataFromState.Name.ValueString()))
			resp.State.RemoveResource(ctx)
		} else {
			resp.Diagnostics.AddError("DB Query Error", fmt.Sprintf("SQL query to read role encountered an unexpected error, please share this with the developer, query=`%s`, error: %s", roleSql, err))
			return
		}
	}

	dataFromState.BypassRowLevelSecurity = types.BoolValue(bypassRowLevelSecurity)
	dataFromState.CanLogin = types.BoolValue(canLogin)
	dataFromState.ConnectionLimit = types.Int32Value(connectionLimit)
	dataFromState.CreateRole = types.BoolValue(createRole)
	dataFromState.Inherit = types.BoolValue(inherit)
	dataFromState.Replication = types.BoolValue(replication)
	dataFromState.Superuser = types.BoolValue(superuser)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &dataFromState)...)
}

func (r *RoleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var dataFromPlan RoleResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &dataFromPlan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	txn, err := r.data.DbPool.Begin(ctx)

	if err != nil {
		resp.Diagnostics.AddError("DB Connection Pool Error", fmt.Sprintf("Unable to start a new transaction creating a connection pool, got error: %s", err))
		return
	}

	defer func() {
		if err != nil {
			err := txn.Rollback(ctx)
			if err != nil {
				resp.Diagnostics.AddError("Transaction Rollback Error", fmt.Sprintf("Unable to rollback transaction, got error: %s", err))
			}
		}
	}()

	alterRoleSql := fmt.Sprintf("ALTER ROLE %s WITH %s;", dataFromPlan.Name.ValueString(), dataFromPlan.GetOptionsString(r))

	tflog.Info(ctx, alterRoleSql)

	if _, err = txn.Exec(ctx, alterRoleSql); err != nil {
		resp.Diagnostics.AddError("DB role update error", fmt.Sprintf("Error executing query '%s', got error: %s", alterRoleSql, err))
		return
	}

	if err = txn.Commit(ctx); err != nil {
		resp.Diagnostics.AddError("DB transaction error", fmt.Sprintf("Error committing DB transaction, got error: %s", err))
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("Successfully altered Postgresql Role: %s", dataFromPlan.Name.ValueString()))

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &dataFromPlan)...)
}

func (r *RoleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data RoleResourceModel

	// Read Terraform prior state data into the model...
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	txn, err := r.data.DbPool.Begin(ctx)

	if err != nil {
		resp.Diagnostics.AddError("DB Connection Pool Error", fmt.Sprintf("Unable to start a new transaction, got error: %s", err))
		return
	}

	dropRoleSql := fmt.Sprintf("DROP ROLE %s;", data.Name.ValueString())

	if _, err = txn.Exec(ctx, dropRoleSql); err != nil {
		resp.Diagnostics.AddError("DB role deletion error", fmt.Sprintf("Error executing query '%s', got error: %s", dropRoleSql, err))
		return
	}

	err = txn.Commit(ctx)
	if err != nil {
		resp.Diagnostics.AddError("DB transaction error", fmt.Sprintf("Error committing DB transaction, got error: %s", err))
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("Successfully dropped Postgresql Role: %s", data.Name.ValueString()))
}

func (r *RoleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

func (r *RoleResourceModel) BypassRowLevelSecurityAsOptionString() string {
	if r.BypassRowLevelSecurity.ValueBool() {
		return "BYPASSRLS"
	} else {
		return "NOBYPASSRLS"
	}
}

func (r *RoleResourceModel) CanLoginAsOptionString() string {
	if r.CanLogin.ValueBool() {
		return "LOGIN"
	} else {
		return "NOLOGIN"
	}
}

func (r *RoleResourceModel) ConnectionLimitAsOptionString() string {
	return fmt.Sprintf("CONNECTION LIMIT %d", r.ConnectionLimit.ValueInt32())
}

func (r *RoleResourceModel) CreateRoleAsOptionString() string {
	if r.CreateRole.ValueBool() {
		return "CREATEROLE"
	} else {
		return "NOCREATEROLE"
	}
}

func (r *RoleResourceModel) InheritAsOptionString() string {
	if r.Inherit.ValueBool() {
		return "INHERIT"
	} else {
		return "NOINHERIT"
	}
}

func (r *RoleResourceModel) ReplicationAsOptionString() string {
	if r.Replication.ValueBool() {
		return "REPLICATION"
	} else {
		return "NOREPLICATION"
	}
}

func (r *RoleResourceModel) SuperuserAsOptionString() string {
	if r.Superuser.ValueBool() {
		return "SUPERUSER"
	} else {
		return "NOSUPERUSER"
	}
}

func (r *RoleResourceModel) GetOptionsString(resource *RoleResource) string {
	options := []string{
		r.BypassRowLevelSecurityAsOptionString(),
		r.CanLoginAsOptionString(),
		r.ConnectionLimitAsOptionString(),
		r.CreateRoleAsOptionString(),
		r.InheritAsOptionString(),
		r.ReplicationAsOptionString(),
		r.SuperuserAsOptionString(),
	}

	return strings.Join(options, " ")
}
