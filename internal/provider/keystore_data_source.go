package provider

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"

    "github.com/hashicorp/terraform-plugin-framework/datasource"
    "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
    "github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &KeyStoreDataSource{}

func NewKeyStoreDataSource() datasource.DataSource {
    return &KeyStoreDataSource{}
}

// KeyStoreDataSource defines the data source implementation.
type KeyStoreDataSource struct {
    client *ClientConfig
}

// KeyStoreDataSourceModel describes the data source data model.
type KeyStoreDataSourceModel struct {
    Id    types.Int64  `tfsdk:"id"`
    Name  types.String `tfsdk:"name"`
    Value types.String `tfsdk:"value"`
}

func (d *KeyStoreDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
    resp.TypeName = req.ProviderTypeName + "_keystore"
}

func (d *KeyStoreDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
    resp.Schema = schema.Schema{
        MarkdownDescription: "KeyStore data source for Tactical RMM. Use this to look up existing keystore entries by ID or name.",

        Attributes: map[string]schema.Attribute{
            "id": schema.Int64Attribute{
                MarkdownDescription: "KeyStore identifier. Either `id` or `name` must be specified.",
                Optional:            true,
                Computed:            true,
            },
            "name": schema.StringAttribute{
                MarkdownDescription: "Key name. Either `id` or `name` must be specified.",
                Optional:            true,
                Computed:            true,
            },
            "value": schema.StringAttribute{
                MarkdownDescription: "Key value",
                Computed:            true,
                Sensitive:           true,
            },
        },
    }
}

func (d *KeyStoreDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
    if req.ProviderData == nil {
        return
    }

    client, ok := req.ProviderData.(*ClientConfig)
    if !ok {
        resp.Diagnostics.AddError(
            "Unexpected Data Source Configure Type",
            fmt.Sprintf("Expected *ClientConfig, got: %T. Please report this issue to the provider developers.", req.ProviderData),
        )
        return
    }

    d.client = client
}

func (d *KeyStoreDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
    var data KeyStoreDataSourceModel

    resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
    if resp.Diagnostics.HasError() {
        return
    }

    // Validate that either ID or name is provided
    if data.Id.IsNull() && data.Name.IsNull() {
        resp.Diagnostics.AddError(
            "Missing KeyStore Identifier",
            "Either 'id' or 'name' must be specified to look up a keystore entry.",
        )
        return
    }

    // Get all keystore entries since there's no individual GET endpoint
    httpReq, err := http.NewRequest("GET", fmt.Sprintf("%s/core/keystore/", d.client.BaseURL), nil)
    if err != nil {
        resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read keystore entries, got error: %s", err))
        return
    }

    httpResp, err := d.client.Do(httpReq)
    if err != nil {
        resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read keystore entries, got error: %s", err))
        return
    }
    defer httpResp.Body.Close()

    if httpResp.StatusCode != http.StatusOK {
        resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read keystore entries, status code: %d", httpResp.StatusCode))
        return
    }

    var entries []map[string]interface{}
    if err := json.NewDecoder(httpResp.Body).Decode(&entries); err != nil {
        resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to parse response, got error: %s", err))
        return
    }

    // Find the entry by ID or name
    var foundEntry map[string]interface{}
    
    if !data.Id.IsNull() {
        // Look up by ID
        for _, entry := range entries {
            if id, ok := entry["id"].(float64); ok && int64(id) == data.Id.ValueInt64() {
                foundEntry = entry
                break
            }
        }
        if foundEntry == nil {
            resp.Diagnostics.AddError("KeyStore Entry Not Found", fmt.Sprintf("KeyStore entry with ID %d not found", data.Id.ValueInt64()))
            return
        }
    } else {
        // Look up by name
        for _, entry := range entries {
            if name, ok := entry["name"].(string); ok && name == data.Name.ValueString() {
                foundEntry = entry
                break
            }
        }
        if foundEntry == nil {
            resp.Diagnostics.AddError("KeyStore Entry Not Found", fmt.Sprintf("KeyStore entry with name '%s' not found", data.Name.ValueString()))
            return
        }
    }

    // Update model with found entry data
    if id, ok := foundEntry["id"].(float64); ok {
        data.Id = types.Int64Value(int64(id))
    }
    if name, ok := foundEntry["name"].(string); ok {
        data.Name = types.StringValue(name)
    }
    if value, ok := foundEntry["value"].(string); ok {
        data.Value = types.StringValue(value)
    }

    resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
