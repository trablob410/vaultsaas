use serde_json::{json, Value};
use crate::client::ValtClient;
use crate::error::Result;
use crate::keychain;
use crate::protocol::ToolDef;
use crate::scanner;

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
        "scan_secrets" => tool_scan_secrets(args, client).await,
        "store_secret" => tool_store_secret(args, client).await,
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

async fn tool_scan_secrets(args: &Value, client: &ValtClient) -> Result<Value> {
    let path = args["path"].as_str()
        .ok_or_else(|| crate::error::ValtError::Protocol("path required".into()))?;
    let recursive = args.get("recursive").and_then(|v| v.as_bool()).unwrap_or(true);
    let findings = scanner::scan_directory(path, recursive);
    let count = findings.len();
    let findings_json: Vec<Value> = findings.iter().map(|f| json!({
        "file_path": f.file_path,
        "line_number": f.line_number,
        "pattern_name": f.pattern_name,
        "credential_type": f.credential_type,
        "redacted_preview": f.redacted_preview,
    })).collect();
    // If project_id provided, persist to backend
    if let Some(pid) = args.get("project_id").and_then(|v| v.as_str()) {
        let _ = client.create_scan_result(pid, path, count).await;
        // Ignore errors (scan result persistence is best-effort)
    }
    Ok(json!({
        "findings": findings_json,
        "count": count,
        "scanned_path": path
    }))
}

async fn tool_store_secret(args: &Value, client: &ValtClient) -> Result<Value> {
    let name = args["name"].as_str()
        .ok_or_else(|| crate::error::ValtError::Protocol("name required".into()))?;
    let credential_type = args["credential_type"].as_str()
        .ok_or_else(|| crate::error::ValtError::Protocol("credential_type required".into()))?;
    let value = args["value"].as_str()
        .ok_or_else(|| crate::error::ValtError::Protocol("value required".into()))?;
    let source = args.get("source").and_then(|v| v.as_str()).unwrap_or("scanner");
    let result = client.create_secret(name, credential_type, value, source).await?;
    Ok(json!({
        "secret_id": result.get("id"),
        "name": name,
        "message": "Secret stored successfully"
    }))
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_list_tools_returns_eight() {
        let tools = list_tools();
        assert_eq!(tools.len(), 8);
        let names: Vec<&str> = tools.iter().map(|t| t.name.as_str()).collect();
        assert!(names.contains(&"request_secret_access"));
        assert!(names.contains(&"list_my_secrets"));
        assert!(names.contains(&"authenticate_agent"));
        assert!(names.contains(&"scan_secrets"));
        assert!(names.contains(&"store_secret"));
    }

    #[test]
    fn test_tool_schemas_have_required_fields() {
        for tool in list_tools() {
            assert!(!tool.name.is_empty());
            assert!(!tool.description.is_empty());
            assert_eq!(tool.input_schema.get("type").and_then(|v| v.as_str()), Some("object"));
        }
    }

    #[tokio::test]
    async fn test_call_unknown_tool_errors() {
        use crate::config::ValtConfig;
        let config = ValtConfig::default();
        let result = crate::client::ValtClient::new(config.api_url, "test_token".into());
        if let Ok(client) = result {
            let err = call_tool("nonexistent_tool", &json!({}), &client).await;
            assert!(err.is_err());
            let msg = err.unwrap_err().to_string();
            assert!(msg.contains("Unknown tool"));
        }
    }
}
