package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ktham/terraform-provider-postgresql/internal/postgresql"
)

var _ provider.Provider = &PostgresqlProvider{}

type PostgresqlProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

type PostgresqlProviderData struct {
	DbPool    *pgxpool.Pool
	DbType    string
	DbVersion string
}

type PostgresqlProviderModel struct {
	Hostname       types.String `tfsdk:"hostname"`
	Port           types.Int32  `tfsdk:"port"`
	DatabaseName   types.String `tfsdk:"database_name"`
	Username       types.String `tfsdk:"username"`
	Password       types.String `tfsdk:"password"`
	MaxConnections types.Int32  `tfsdk:"max_connections"`
}

func (p *PostgresqlProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "postgresql"
	resp.Version = p.version
}

func (p *PostgresqlProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"hostname": schema.StringAttribute{
				Description: "The host name of the Postgres server.",
				Required:    true,
			},
			"port": schema.Int32Attribute{
				Description: "The TCP port on which Postgres is listening for connections.",
				Required:    true,
				Validators: []validator.Int32{
					int32validator.Between(1, 65535),
				},
			},
			"database_name": schema.StringAttribute{
				Description: "The name of the database to connect to.",
				Optional:    true,
			},
			"username": schema.StringAttribute{
				Description: "The user used for connecting to Postgres.",
				Required:    true,
			},
			"password": schema.StringAttribute{
				Description: "The password to use for authentication.",
				Optional:    true,
				Sensitive:   true,
			},
			"max_connections": schema.Int32Attribute{
				Description: "Maximum number of connections to establish to the database. Zero means unlimited.",
				Optional:    true,
			},
		},
	}
}

func (p *PostgresqlProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config PostgresqlProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if config.Hostname.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("hostname"),
			"Unknown Postgresql hostname value",
			"The provider cannot create a connection to the Postgres server as `hostname` is an unknown configuration value",
		)
	}
	if config.Port.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("port"),
			"Unknown Postgresql port value",
			"The provider cannot create a connection to the Postgres server as `port` is an unknown configuration value",
		)
	}
	if config.DatabaseName.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("database_name"),
			"Unknown Postgresql database_name value",
			"The provider cannot create a connection to the Postgres server as `database_name` is an unknown configuration value",
		)
	}
	if config.Username.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("username"),
			"Unknown Postgresql username value",
			"The provider cannot create a connection to the Postgres server as `username` is an unknown configuration value",
		)
	}
	if config.Password.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("password"),
			"Unknown Postgresql password value",
			"The provider cannot create a connection to the Postgres server as `password` is an unknown configuration value",
		)
	}
	if config.MaxConnections.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("max_connections"),
			"Unknown Postgresql max_connections value",
			"The provider cannot create a connection to the Postgres server as `max_connections` is an unknown configuration value",
		)
	}

	dbName := "postgres"
	dbPort := int32(5432)
	dbUserPass := "postgres"

	if !config.DatabaseName.IsNull() {
		dbName = config.DatabaseName.ValueString()
	}
	if !config.Port.IsNull() {
		dbPort = config.Port.ValueInt32()
	}

	if !config.Username.IsNull() && !config.Password.IsNull() {
		dbUserPass = fmt.Sprintf("%s:%s", config.Username.ValueString(), config.Password.ValueString())
	}

	if !config.Username.IsNull() && config.Password.IsNull() {
		dbUserPass = config.Username.ValueString()
	}

	databaseURL := fmt.Sprintf("postgresql://%s@%s:%d/%s",
		dbUserPass,
		config.Hostname.ValueString(),
		dbPort,
		dbName,
	)

	tflog.Info(ctx, "Configuring DB Connection Pool")
	dbConnPool, err := pgxpool.New(context.Background(), databaseURL)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create DB connection pool",
			"An unexpected error occurred when creating the DB connection pool. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"Error: "+err.Error(),
		)
		return
	}

	var version string
	err = dbConnPool.QueryRow(context.Background(), "SELECT VERSION();").Scan(&version)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to determine database version/type.",
			"An unexpected error occurred when running `SELECT VERSION();`, please share this error with the provider developers. Error: "+err.Error(),
		)
		return
	}

	parsedVersion, err := postgresql.ParseDbVersion(version)

	// TODO: Gracefully handle - avoid failure if database vendor decides to change version string format in a way
	// that we currently don't support parsing for yet.
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to parse results of `SELECT VERSION();` from this database",
			"An unexpected error occurred when creating the DB connection pool. Error: "+err.Error(),
		)
		return
	}

	providerData := PostgresqlProviderData{
		DbPool:    dbConnPool,
		DbType:    parsedVersion.DbType,
		DbVersion: parsedVersion.DbVersion,
	}

	resp.DataSourceData = providerData
	resp.ResourceData = providerData
}

func (p *PostgresqlProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewRoleResource,
	}
}

func (p *PostgresqlProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &PostgresqlProvider{
			version: version,
		}
	}
}
