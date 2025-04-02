package cli

import (
	"fmt"
	"strings"
)

// printUsage prints the usage information for the CLI tool.
func (cli *CLI) printUsage() {
	fmt.Println("\nCLI Tool Usage:")
	fmt.Println("Usage: [command] [subcommand] [flags]")
	fmt.Println("\nGlobal Flags:")
	fmt.Println("  -h, --help          Show help for the CLI tool")
	fmt.Println("  -v, --version       Show the version of the CLI tool")

	fmt.Println("\nAvailable Commands:")
	for _, cmd := range cli.rootCommands {
		cli.printCommandHelp(cmd, 1)
	}
	fmt.Println()
}

// printCommandHelp prints the help information for a specific command.
func (cli *CLI) printCommandHelp(cmd *Command, indent int) {
	indentation := strings.Repeat("  ", indent)
	fmt.Printf("%s%s: %s\n", indentation, cmd.Name, cmd.Usage)

	if len(cmd.Flags) > 0 {
		fmt.Printf("%sFlags:\n", indentation)
		for _, flag := range cmd.Flags {
			aliases := ""
			if len(flag.Aliases) > 0 {
				aliases = fmt.Sprintf("-%s, ", strings.Join(flag.Aliases, ", -"))
			}
			fmt.Printf("%s  %s--%s value\t%s (default: %s)\n",
				indentation,
				aliases,
				flag.Name,
				flag.Usage,
				flag.Default)
		}
	}

	if len(cmd.SubCommands) > 0 {
		fmt.Printf("%sSubcommands:\n", indentation)
		for _, subCmd := range cmd.SubCommands {
			cli.printCommandHelp(subCmd, indent+1)
		}
	}
}

// printDetailedHelp prints the detailed help information for a specific command.
func (cli *CLI) printDetailedHelp(commandStack []string, cmd *Command) {
	fmt.Printf("\nHelp for Command: %s\n", strings.Join(commandStack, " "))
	fmt.Printf("\n%s: %s\n", cmd.Name, cmd.Usage)
	fmt.Println("\nUsage:")
	fmt.Printf("  %s [flags]\n", strings.Join(commandStack, " "))

	if len(cmd.Flags) > 0 {
		fmt.Println("\nFlags:")
		for _, flag := range cmd.Flags {
			aliases := ""
			if len(flag.Aliases) > 0 {
				aliases = fmt.Sprintf("-%s, ", strings.Join(flag.Aliases, ", -"))
			}
			fmt.Printf("%s--%s value\t%s (default: %s)\n",
				aliases,
				flag.Name,
				flag.Usage,
				flag.Default)
		}
	}

	if len(cmd.SubCommands) > 0 {
		fmt.Println("\nSubcommands:")
		for _, subCmd := range cmd.SubCommands {
			fmt.Printf("  %s: %s\n", subCmd.Name, subCmd.Usage)
		}
	}
	fmt.Println()
}
