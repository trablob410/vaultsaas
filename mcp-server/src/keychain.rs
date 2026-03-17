use crate::error::{Result, ValtError};

const SERVICE: &str = "valt-mcp-server";
const USER: &str = "auth-token";

const AGENT_TOKEN_USER: &str = "valt-agent";
const AGENT_TOKEN_KEY: &str = "valt-agent-token";

pub fn get_token() -> Option<String> {
    // Check env var first (useful for CI/headless)
    if let Ok(token) = std::env::var("VALT_AUTH_TOKEN") {
        if !token.is_empty() {
            return Some(token);
        }
    }
    // Try OS keychain
    keyring::Entry::new(SERVICE, USER)
        .ok()
        .and_then(|e| e.get_password().ok())
        .filter(|t| !t.is_empty())
}

#[allow(dead_code)]
pub fn store_token(token: &str) -> Result<()> {
    let entry = keyring::Entry::new(SERVICE, USER)
        .map_err(|e| ValtError::Keychain(e.to_string()))?;
    entry.set_password(token)
        .map_err(|e| ValtError::Keychain(e.to_string()))
}

#[allow(dead_code)]
pub fn delete_token() -> Result<()> {
    let entry = keyring::Entry::new(SERVICE, USER)
        .map_err(|e| ValtError::Keychain(e.to_string()))?;
    entry.delete_credential()
        .map_err(|e| ValtError::Keychain(e.to_string()))
}

pub fn get_agent_token() -> Result<Option<String>> {
    // Check env var first (useful for CI/headless)
    if let Ok(token) = std::env::var("VALT_AGENT_TOKEN") {
        if !token.is_empty() {
            return Ok(Some(token));
        }
    }
    // Try OS keychain; not-found is not an error
    let token = keyring::Entry::new(AGENT_TOKEN_USER, AGENT_TOKEN_KEY)
        .ok()
        .and_then(|e| e.get_password().ok())
        .filter(|t| !t.is_empty());
    Ok(token)
}

pub fn set_agent_token(token: &str) -> Result<()> {
    let entry = keyring::Entry::new(AGENT_TOKEN_USER, AGENT_TOKEN_KEY)
        .map_err(|e| ValtError::Keychain(e.to_string()))?;
    entry.set_password(token)
        .map_err(|e| ValtError::Keychain(e.to_string()))
}

#[allow(dead_code)]
pub fn delete_agent_token() -> Result<()> {
    let entry = keyring::Entry::new(AGENT_TOKEN_USER, AGENT_TOKEN_KEY)
        .map_err(|e| ValtError::Keychain(e.to_string()))?;
    entry.delete_credential()
        .map_err(|e| ValtError::Keychain(e.to_string()))
}
