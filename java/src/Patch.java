package hermes;

public class Patch {
    public static String applyDiffBlock(String content, String searchBlock, String replaceBlock) throws Exception {
        int index = content.indexOf(searchBlock);
        if (index < 0) {
            throw new Exception("search block not found in file content");
        }

        return content.substring(0, index) + replaceBlock + content.substring(index + searchBlock.length());
    }
}
