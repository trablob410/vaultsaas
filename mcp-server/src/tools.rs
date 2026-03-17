use serde_json::{json, Value};
use crate::client::ValtClient;
use crate::error::Result;
use crate::keychain;
use crate::protocol::ToolDef;
use crate::scanner_tools;

pub fn list_tools() -> Vec<ToolDef> {
    vec![
        ToolDef {
            name: "request_secret_access".into(),
            description: "Request temporary access to a secret. Returns request ID for status checking.".into(),
            input_schema: json!({
                "type": "object",
                "properties": {
                    "secret_id": {"type": "string", "description": "UUID of the secret"},
                    "reason": {"type": "string", "description": "Why access is needed"},
                    "duration_minutes": {"type": "integer", "description": "Requested duration (1-60)", "default": 30}
                },
                "required": ["secret_id", "reason"]
            }),
        },
        ToolDef {
            name: "check_approval_status".into(),
            description: "Check the approval status of an access request.".into(),
            input_schema: json!({
                "type": "object",
                "properties": {
                    "request_id": {"type": "string", "description": "UUID of the access request"}
                },
                "required": ["request_id"]
            }),
        },
        ToolDef {
            name: "get_credential".into(),
            description: "Retrieve an approved credential. Only works for approved requests.".into(),
            input_schema: json!({
                "type": "object",
                "properties": {
                    "request_id": {"type": "string", "description": "UUID of the approved access request"}
                },
                "required": ["request_id"]
            }),
        },
        ToolDef {
            name: "revoke_credential".into(),
            description: "Revoke an active credential before it expires.".into(),
            input_schema: json!({
                "type": "object",
                "properties": {
                    "request_id": {"type": "string", "description": "UUID of the access request to revoke"}
                },
                "required": ["request_id"]
            }),
        },
        ToolDef {
            name: "list_my_secrets".into(),
            description: "List all secrets you own. Returns metadata only, no secret values.".into(),
            input_schema: json!({
                "type": "object",
                "properties": {
                    "page": {"type": "integer", "description": "Page number (default: 1)"},
                    "limit": {"type": "integer", "description": "Items per page (default: 20)"}
                }
            }),
        },
        ToolDef {
            name: "authenticate_agent".into(),
            description: "Store an agent token for authenticating with the Valt server. Run this once after receiving a token from the Valt dashboard.".into(),
            input_schema: json!({
                "type": "object",
                "properties": {
                    "token": {"type": "string", "description": "The agent token issued from the Valt dashboard"}
                },
                "required": ["token"]
            }),
        },
        ToolDef {
            name: "scan_secrets".into(),
            description: "Scan a directory for hardcoded secrets and credentials. Returns redacted findings - never exposes raw values.".into(),
            input_schema: json!({
                "type": "object",
                "properties": {
                    "path": {"type": "string", "description": "Directory path to scan"},
                    "recursive": {"type": "boolean", "description": "Scan subdirectories", "default": true},
                    "project_id": {"type": "string", "description": "Project ID to persist scan result (optional)"}
                },
                "required": ["path"]
            }),
        },
        ToolDef {
            name: "store_secret".into(),
            description: "Store a secret in the Valt vault. Use after scanning to import found credentials.".into(),
            input_schema: json!({
                "type": "object",
                "properties": {
                    "name": {"type": "string", "description": "Human-readable name for the secret"},
                    "credential_type": {"type": "string", "description": "Type of credential (e.g. api_key, database_uri, ssh_key)"},
                    "value": {"type": "string", "description": "The secret value to store (will be encrypted)"},
                    "source": {"type": "string", "description": "Where the secret came from (e.g. scanner, manual)"}
                },
                "required": ["name", "credential_type", "value"]
            }),
        },
    ]
}

pub async fn call_tool(name: &str, args: &Value, client: &ValtClient) -> Result<Value> {
    match name {
        "request_secret_access" => tool_request_access(args, client).await,
        "check_approval_status" => tool_check_status(args, client).await,
        "get_credential" => tool_get_credential(args, client).await,
        "revoke_credential" => tool_revoke_credential(args, client).await,
        "list_my_secrets" => tool_list_secrets(args, client).await,
        "authenticate_agent" => tool_authenticate_agent(args).await,
        "scan_secrets" => scanner_tools::tool_scan_secrets(args, client).await,
        "store_secret" => scanner_tools::tool_store_secret(args, client).await,
        _ => Err(crate::error::ValtError::Protocol(format!("Unknown tool: {name}"))),
    }
}

async fn tool_request_access(args: &Value, client: &ValtClient) -> Result<Value> {
    let secret_id = args["secret_id"].as_str()
        .ok_or_else(|| crate::error::ValtError::Protocol("secret_id required".into()))?;
    let reason = args["reason"].as_str()
        .ok_or_else(|| crate::error::ValtError::Protocol("reason required".into()))?;
    let duration = args.get("duration_minutes").and_then(|v| v.as_u64()).unwrap_or(30) as u32;
    let result = client.create_access_request(secret_id, reason, duration).await?;
    Ok(json!({
        "request_id": result.get("id"),
        "status": result.get("status"),
        "message": "Access request submitted. Use check_approval_status to poll for approval."
    }))
}

async fn tool_check_status(args: &Value, client: &ValtClient) -> Result<Value> {
    let request_id = args["request_id"].as_str()
        .ok_or_else(|| crate::error::ValtError::Protocol("request_id required".into()))?;
    let requests = client.get_access_requests(None).await?;
    let requests_arr = requests.get("requests")
        .and_then(|v| v.as_array())
        .cloned()
        .unwrap_or_default();
    let found = requests_arr.iter().find(|r| {
        r.get("id").and_then(|v| v.as_str()) == Some(request_id)
    });
    match found {
        Some(req) => Ok(json!({
            "request_id": request_id,
            "status": req.get("status"),
            "created_at": req.get("created_at"),
            "approved_at": req.get("approved_at"),
            "expires_at": req.get("expires_at"),
            "rejection_reason": req.get("rejection_reason"),
        })),
        None => Ok(json!({"error": "Request not found", "request_id": request_id})),
    }
}

async fn tool_get_credential(args: &Value, client: &ValtClient) -> Result<Value> {
    let request_id = args["request_id"].as_str()
        .ok_or_else(|| crate::error::ValtError::Protocol("request_id required".into()))?;
    let cred = client.get_credential(request_id).await?;
    // Return credential metadata; actual data accessible via resource URI
    Ok(json!({
        "credential_id": cred.get("id"),
        "request_id": request_id,
        "issued_at": cred.get("issued_at"),
        "expires_at": cred.get("expires_at"),
        "status": cred.get("status"),
        "message": "Credential retrieved. Access data via vault://requests/{request_id} resource."
    }))
}

async fn tool_revoke_credential(args: &Value, client: &ValtClient) -> Result<Value> {
    let request_id = args["request_id"].as_str()
        .ok_or_else(|| crate::error::ValtError::Protocol("request_id required".into()))?;
    client.revoke_credential(request_id).await?;
    Ok(json!({"message": "Credential revoked successfully", "request_id": request_id}))
}

async fn tool_authenticate_agent(args: &Value) -> Result<Value> {
    let token = args["token"].as_str()
        .ok_or_else(|| crate::error::ValtError::Protocol("token required".into()))?;
    keychain::set_agent_token(token)?;
    Ok(json!({"message": "Agent token stored successfully. You can now use Valt tools."}))
}

async fn tool_list_secrets(args: &Value, client: &ValtClient) -> Result<Value> {
    let _ = args; // pagination not yet implemented in backend query
    let result = client.list_secrets().await?;
    // Strip sensitive fields, return only metadata
    if let Some(secrets) = result.get("secrets").and_then(|v| v.as_array()) {
        let safe: Vec<Value> = secrets.iter().map(|s| json!({
            "id": s.get("id"),
            "name": s.get("name"),
            "credential_type": s.get("credential_type"),
            "status": s.get("status"),
            "created_at": s.get("created_at"),
        })).collect();
        let count = safe.len();
        return Ok(json!({"secrets": safe, "count": count}));
    }
    Ok(json!({"secrets": [], "count": 0}))
}


// Tests live in scanner_tools.rs to keep this file under 200 lines.
