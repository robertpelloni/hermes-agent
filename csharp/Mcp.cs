using System;
using System.Collections.Generic;
using System.Diagnostics;
using System.IO;

namespace Hermes
{
    public class McpPlugin
    {
        public string Name { get; set; }
        public Process Process { get; set; }
        public StreamWriter StandardInput { get; set; }
        public StreamReader StandardOutput { get; set; }
    }

    public class McpPluginManager
    {
        private readonly Dictionary<string, McpPlugin> _plugins = new Dictionary<string, McpPlugin>();
        private readonly object _lock = new object();

        public void Load(string name, string command, string args)
        {
            lock (_lock)
            {
                var startInfo = new ProcessStartInfo
                {
                    FileName = command,
                    Arguments = args,
                    RedirectStandardInput = true,
                    RedirectStandardOutput = true,
                    UseShellExecute = false,
                    CreateNoWindow = true
                };

                var process = Process.Start(startInfo);
                if (process == null) throw new Exception("Failed to start process");

                var plugin = new McpPlugin
                {
                    Name = name,
                    Process = process,
                    StandardInput = process.StandardInput,
                    StandardOutput = process.StandardOutput
                };

                _plugins[name] = plugin;
                Console.WriteLine($"[hermes:mcp] Loaded plugin: {name}");
            }
        }

        public void Unload(string name)
        {
            lock (_lock)
            {
                if (_plugins.TryGetValue(name, out var plugin))
                {
                    plugin.Process.Kill();
                    _plugins.Remove(name);
                    Console.WriteLine($"[hermes:mcp] Unloaded plugin: {name}");
                }
            }
        }
    }
}
