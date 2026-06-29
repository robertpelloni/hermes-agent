use tree_sitter::{Parser, Language};
use std::fs;
use std::path::Path;

// This would use tree-sitter-rust in a real project
extern "C" {
    fn tree_sitter_rust() -> Language;
}

pub fn generate_ast_map(base_dir: &str, files: &[String]) -> Result<String, Box<dyn std::error::Error>> {
    let language = unsafe { tree_sitter_rust() };
    let mut parser = Parser::new();
    parser.set_language(&language)?;

    let mut builder = String::from("AST-Based Repository Map:\n");

    for rel_path in files {
        let full_path = Path::new(base_dir).join(rel_path);
        if let Ok(content) = fs::read(&full_path) {
            let tree = parser.parse(&content, None).unwrap();
            let root_node = tree.root_node();

            builder.push_str(&format!("\n## {}\n", rel_path));

            let mut cursor = root_node.walk();
            for child in root_node.children(&mut cursor) {
                if child.kind() == "function_item" || child.kind() == "struct_item" {
                    if let Some(name_node) = child.child_by_field_name("name") {
                        if let Ok(name) = name_node.utf8_text(&content) {
                            builder.push_str(&format!("- {}\n", name));
                        }
                    }
                }
            }
        }
    }

    Ok(builder)
}
