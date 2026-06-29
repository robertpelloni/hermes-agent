package hermes;

public class Patcher {
    public static String applyDiffBlock(String content, String searchBlock, String replaceBlock) throws Exception {
        if (!content.contains(searchBlock)) {
            throw new Exception("Search block not found in file content");
        }
        return content.replaceFirst(java.util.regex.Pattern.quote(searchBlock), java.util.regex.Matcher.quoteReplacement(replaceBlock));
    }
}
