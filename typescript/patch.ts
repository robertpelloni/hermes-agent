export function applyDiffBlock(content: string, searchBlock: string, replaceBlock: string): string {
    if (!content.includes(searchBlock)) {
        throw new Error("Search block not found in file content");
    }

    return content.replace(searchBlock, replaceBlock);
}
