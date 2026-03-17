use thiserror::Error;

#[derive(Debug, Error)]
#[allow(dead_code)]
pub enum ValtError {
    #[error("Config error: {0}")]
    Config(String),
    #[error("HTTP error: {0}")]
    Http(#[from] reqwest::Error),
    #[error("API error {status}: {message}")]
    Api { status: u16, message: String },
    #[error("Crypto error: {0}")]
    Crypto(String),
    #[error("Keychain error: {0}")]
    Keychain(String),
    #[error("Protocol error: {0}")]
    Protocol(String),
    #[error("IO error: {0}")]
    Io(#[from] std::io::Error),
}

pub type Result<T> = std::result::Result<T, ValtError>;
