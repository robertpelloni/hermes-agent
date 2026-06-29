use std::error::Error;
use std::fmt;

#[derive(Debug)]
pub struct PatchError(String);

impl fmt::Display for PatchError {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        write!(f, "{}", self.0)
    }
}

impl Error for PatchError {}

pub fn apply_diff_block(content: &str, search_block: &str, replace_block: &str) -> Result<String, PatchError> {
    if !content.contains(search_block) {
        return Err(PatchError("Search block not found in file content".to_string()));
    }

    Ok(content.replacen(search_block, replace_block, 1))
}
