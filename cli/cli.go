package cli

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
)

type CLI struct {
	rootCommands map[string]*Command
}

func NewCLI() *CLI {
	return &CLI{
		rootCommands: make(map[string]*Command),
	}
}

func (cli *CLI) AddCommand(cmd *Command) {
	cli.rootCommands[cmd.Name] = cmd
}

func (cli *CLI) Execute() error {
	if len(os.Args) < 2 {
		cli.printUsage()
		return errors.New("no command provided")
	}

	args := os.Args[1:]

	// Global help flag
	if len(args) == 1 && (args[0] == "-h" || args[0] == "--help") {
		cli.printUsage()
		return nil
	}

	ctx := NewCLIContext()
	currentCommands := cli.rootCommands
	var currentCommand *Command

	for len(args) > 0 {
		name := args[0]
		if cmd, exists := currentCommands[name]; exists {
			currentCommand = cmd
			ctx.CommandStack = append(ctx.CommandStack, name)
			args = args[1:]

			// Prepare flag parsing
			flagSet := flag.NewFlagSet(name, flag.ExitOnError)
			flagAliasMap := make(map[string]string)

			for _, fl := range currentCommand.Flags {
				// Set default value in context first
				ctx.SetFlag(fl.Name, fl.Default)
				// Then register with flagSet
				flagSet.String(fl.Name, fl.Default, fl.Usage)
				for _, alias := range fl.Aliases {
					flagAliasMap["--"+alias] = fl.Name
					flagAliasMap["-"+alias] = fl.Name
				}
			}

			// Help flag for the current command
			showHelp := flagSet.Bool("help", false, "Show help for this command")
			flagSet.BoolVar(showHelp, "h", false, "Show help for this command")

			parsedArgs := []string{}
			for i := 0; i < len(args); i++ {
				arg := args[i]
				if strings.HasPrefix(arg, "-") {
					// Check for aliases or flags
					equalIndex := strings.Index(arg, "=")
					if equalIndex != -1 {
						// Format: --alias=value or -a=value
						flagKey := arg[:equalIndex]
						flagValue := arg[equalIndex+1:]
						if primary, exists := flagAliasMap[flagKey]; exists {
							ctx.SetFlag(primary, flagValue)
						} else {
							parsedArgs = append(parsedArgs, arg)
						}
					} else {
						// Format: --alias or -a followed by value
						if primary, exists := flagAliasMap[arg]; exists {
							if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
								ctx.SetFlag(primary, args[i+1])
								i++ // Skip the value
							} else {
								ctx.SetFlag(primary, "")
							}
						} else if strings.HasPrefix(arg, "--") {
							// Handle long-form flags (e.g., --name value)
							flagKey := arg[2:]
							if primary, exists := flagAliasMap["--"+flagKey]; exists {
								if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
									ctx.SetFlag(primary, args[i+1])
									i++
								} else {
									ctx.SetFlag(primary, "")
								}
							} else {
								if arg == "--help" {
									parsedArgs = append(parsedArgs, arg)
								} else {
									parsedArgs = append(parsedArgs, arg+"=")
								}
							}
						} else {
							// Assume short-form flag (e.g., -n value)
							flagKey := arg[1:]
							if primary, exists := flagAliasMap["-"+flagKey]; exists {
								if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
									ctx.SetFlag(primary, args[i+1])
									i++
								} else {
									ctx.SetFlag(primary, "")
								}
							} else {
								parsedArgs = append(parsedArgs, arg)
							}
						}
					}
				} else {
					parsedArgs = append(parsedArgs, arg)
				}
			}

			flagSet.Parse(parsedArgs)
			flagSet.Visit(func(f *flag.Flag) {
				ctx.SetFlag(f.Name, f.Value.String())
			})

			// Ensure all flags have default values if not set or explicitly blank
			for _, fl := range currentCommand.Flags {
				value, exists := ctx.Flags[fl.Name]
				if !exists || value == "" {
					ctx.SetFlag(fl.Name, fl.Default)
				}
			}

			// Update remaining arguments and subcommands
			args = flagSet.Args()
			currentCommands = currentCommand.SubCommands

			if *showHelp {
				cli.printDetailedHelp(ctx.CommandStack, currentCommand)
				return nil
			}
		} else {
			break
		}
	}

	if currentCommand == nil {
		cli.printUsage()
		return fmt.Errorf("unknown command")
	}

	return currentCommand.Handler(ctx)
}
