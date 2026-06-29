using System;

namespace Hermes
{
    public static class Patch
    {
        public static string ApplyDiffBlock(string content, string searchBlock, string replaceBlock)
        {
            int index = content.IndexOf(searchBlock, StringComparison.Ordinal);
            if (index < 0)
            {
                throw new Exception("search block not found in file content");
            }

            return content.Substring(0, index) + replaceBlock + content.Substring(index + searchBlock.Length);
        }
    }
}
