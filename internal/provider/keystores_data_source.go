package provider

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"

    "github.com/hashicorp/terraform-plugin-framework/attr"
    "github.com/hashicorp/terraform-plugin-framework/datasource"
    "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
    "github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &KeyStoresDataSource{}

func NewKeyStoresDataSource() datasource.DataSource {
    return &KeyStoresDataSource{}
}

// KeyStoresDataSource defines the data source implementation.
type KeyStoresDataSource struct {
    client *ClientConfig
}

// KeyStoresDataSourceModel describes the data source data model.
type KeyStoresDataSourceModel struct {
    Id        types.Int64  `tfsdk:"id"`
    Name      types.String `tfsdk:"name"`
    Keystores types.List   `tfsdk:"keystores"`
}

// KeyStoreModel represents a single keystore entry in the list
type KeyStoreModel struct {
    Id    types.Int64  `tfsdk:"id"`
    Name  types.String `tfsdk:"name"`
    Value types.String `tfsdk:"value"`
}

func (d *KeyStoresDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
    resp.TypeName = req.ProviderTypeName + "_keystores"
}

func (d *KeyStoresDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
    resp.Schema = schema.Schema{
        MarkdownDescription: "KeyStores data source for Tactical RMM. Use this to fetch all keystore entries or filter by ID or name.",

        Attributes: map[string]schema.Attribute{
            "id": schema.Int64Attribute{
                MarkdownDescription: "Optional: Filter keystores by a specific ID.",
                Optional:            true,
            },
            "name": schema.StringAttribute{
                MarkdownDescription: "Optional: Filter keystores by name (exact match).",
                Optional:            true,
            },
            "keystores": schema.ListNestedAttribute{
                MarkdownDescription: "List of keystore entries matching the filter criteria, or all entries if no filter is specified.",
                Computed:            true,
                NestedObject: schema.NestedAttributeObject{
                    Attributes: map[string]schema.Attribute{
                        "id": schema.Int64Attribute{
                            MarkdownDescription: "KeyStore identifier",
                            Computed:            true,
                        },
                        "name": schema.StringAttribute{
                            MarkdownDescription: "Key name",
                            Computed:            true,
                        },
                        "value": schema.StringAttribute{
                            MarkdownDescription: "Key value",
                            Computed:            true,
                            Sensitive:           true,
                        },
                    },
                },
            },
        },
    }
}

func (d *KeyStoresDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *KeyStoresDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
    var data KeyStoresDataSourceModel

    resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
    if resp.Diagnostics.HasError() {
        return
    }

    // Fetch all keystore entries
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

    // Filter entries based on criteria
    var filteredEntries []map[string]interface{}
    
    if !data.Id.IsNull() {
        // Filter by ID
        targetId := data.Id.ValueInt64()
        for _, entry := range entries {
            if id, ok := entry["id"].(float64); ok && int64(id) == targetId {
                filteredEntries = append(filteredEntries, entry)
                break
            }
        }
    } else if !data.Name.IsNull() {
        // Filter by name
        targetName := data.Name.ValueString()
        for _, entry := range entries {
            if name, ok := entry["name"].(string); ok && name == targetName {
                filteredEntries = append(filteredEntries, entry)
            }
        }
    } else {
        // No filter, return all entries
        filteredEntries = entries
    }

    // Convert to KeyStoreModel list
    keystoresList := make([]KeyStoreModel, len(filteredEntries))
    for i, entry := range filteredEntries {
        model := KeyStoreModel{}
        
        if id, ok := entry["id"].(float64); ok {
            model.Id = types.Int64Value(int64(id))
        }
        if name, ok := entry["name"].(string); ok {
            model.Name = types.StringValue(name)
        }
        if value, ok := entry["value"].(string); ok {
            model.Value = types.StringValue(value)
        }
        
        keystoresList[i] = model
    }

    // Convert to list value
    keystoreObjectType := types.ObjectType{
        AttrTypes: map[string]attr.Type{
            "id":    types.Int64Type,
            "name":  types.StringType,
            "value": types.StringType,
        },
    }

    keystoresListValue := make([]attr.Value, len(keystoresList))
    for i, keystore := range keystoresList {
        objValue, diags := types.ObjectValueFrom(ctx, keystoreObjectType.AttrTypes, keystore)
        resp.Diagnostics.Append(diags...)
        keystoresListValue[i] = objValue
    }

    listValue, diags := types.ListValue(keystoreObjectType, keystoresListValue)
    resp.Diagnostics.Append(diags...)
    data.Keystores = listValue

    resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
