using System;
using System.Collections.Generic;
using System.IO;

namespace Hermes
{
    public class VirtualFileSystem
    {
        private readonly Dictionary<string, byte[]> _files = new Dictionary<string, byte[]>();
        private readonly object _lock = new object();

        public void WriteFile(string path, byte[] content)
        {
            lock (_lock)
            {
                _files[path] = content;
            }
        }

        public byte[] ReadFile(string path)
        {
            lock (_lock)
            {
                if (_files.TryGetValue(path, out var content))
                {
                    return content;
                }
            }

            return File.ReadAllBytes(path);
        }

        public void Flush()
        {
            lock (_lock)
            {
                foreach (var kvp in _files)
                {
                    File.WriteAllBytes(kvp.Key, kvp.Value);
                    Console.WriteLine($"[hermes:vfs] Flushed: {kvp.Key}");
                }
                _files.Clear();
            }
        }
    }
}
