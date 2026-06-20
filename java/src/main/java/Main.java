public class Main {
    public static void main(String[] args) {
        if (args.length == 0) {
            System.out.println("Hermes Agent CLI (Java)");
            System.out.println("Usage: java Main <command> [options]");
            System.out.println("Commands:");
            System.out.println("  chat [--message <msg>]  Start a chat session");
            System.out.println("  config [key]            Manage configuration");
            return;
        }

        String command = args[0];

        switch (command) {
            case "chat":
                String message = null;
                for (int i = 1; i < args.length; i++) {
                    if (args[i].equals("--message") || args[i].equals("-m")) {
                        if (i + 1 < args.length) {
                            message = args[i+1];
                        }
                        break;
                    }
                }
                if (message != null) {
                    System.out.println("Starting chat with message: " + message);
                } else {
                    System.out.println("Starting interactive chat session...");
                }
                break;
            case "config":
                if (args.length > 1) {
                    System.out.println("Reading config key: " + args[1]);
                } else {
                    System.out.println("Opening interactive config editor...");
                }
                break;
            default:
                System.out.println("Unknown command: " + command);
                break;
        }
    }
}
