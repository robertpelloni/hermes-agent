export function applyDiffBlock(content: string, searchBlock: string, replaceBlock: string): string {
    const index = content.indexOf(searchBlock);
    if (index < 0) {
        throw new Error("search block not found in file content");
    }

    return content.substring(0, index) + replaceBlock + content.substring(index + searchBlock.length);
}
