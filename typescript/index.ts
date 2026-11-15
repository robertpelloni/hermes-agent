import { Command } from 'commander';

const program = new Command();

program
  .name('code-cli')
  .description('Hermes Agent CLI (TypeScript)')
  .version('1.0.0');

program.command('chat')
  .description('Start a chat session')
  .option('-m, --message <msg>', 'The message to send')
  .action((options) => {
    if (options.message) {
      console.log(`Starting chat with message: ${options.message}`);
    } else {
      console.log('Starting interactive chat session...');
    }
  });

program.command('config')
  .description('Manage configuration')
  .argument('[key]', 'The config key')
  .action((key) => {
    if (key) {
      console.log(`Reading config key: ${key}`);
    } else {
      console.log('Opening interactive config editor...');
    }
  });

program.parse(process.argv);
