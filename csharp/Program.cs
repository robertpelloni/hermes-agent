using System;

namespace Hermes.CSharp
{
    class Program
    {
        static void Main(string[] args)
        {
            if (args.Length == 0)
            {
                Console.WriteLine("Hermes Agent CLI (C#)");
                Console.WriteLine("Usage: dotnet run -- <command> [options]");
                Console.WriteLine("Commands:");
                Console.WriteLine("  chat [--message <msg>]  Start a chat session");
                Console.WriteLine("  config [key]            Manage configuration");
                return;
            }

            string command = args[0];

            switch (command)
            {
                case "chat":
                    string message = null;
                    for (int i = 1; i < args.Length; i++)
                    {
                        if (args[i] == "--message" || args[i] == "-m")
                        {
                            if (i + 1 < args.Length)
                            {
                                message = args[i + 1];
                            }
                            break;
                        }
                    }
                    if (message != null)
                    {
                        Console.WriteLine($"Starting chat with message: {message}");
                    }
                    else
                    {
                        Console.WriteLine("Starting interactive chat session...");
                    }
                    break;
                case "config":
                    if (args.Length > 1)
                    {
                        Console.WriteLine($"Reading config key: {args[1]}");
                    }
                    else
                    {
                        Console.WriteLine("Opening interactive config editor...");
                    }
                    break;
                default:
                    Console.WriteLine($"Unknown command: {command}");
                    break;
            }
        }
    }
}
