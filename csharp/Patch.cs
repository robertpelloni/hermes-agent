using System;

namespace Hermes
{
    public static class Patcher
    {
        public static string ApplyDiffBlock(string content, string searchBlock, string replaceBlock)
        {
            if (!content.Contains(searchBlock))
            {
                throw new Exception("Search block not found in file content");
            }

            var index = content.IndexOf(searchBlock);
            return content.Substring(0, index) + replaceBlock + content.Substring(index + searchBlock.Length);
        }
    }
}
