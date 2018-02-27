package cmd

import (
	"fmt"
	"os"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/rs/jplot/data"
	"strings"
	"log"
)

var cfgFile string
var NumberPoints int

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "jplot",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.jplot2.yaml)")

	// add common flags
	rootCmd.PersistentFlags().IntVar(&NumberPoints, "points", 100, "Number of values to plot")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".jplot" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".jplot")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

func parseSpec(args []string) []data.GraphSpec {
	specs := make([]data.GraphSpec, 0, len(args))
	for i, v := range args {
		gs := data.GraphSpec{}
		for j, name := range strings.Split(v, "+") {
			var isCounter bool
			var isMarker bool
			n := strings.Split(name, ":")
			for len(n) > 1 {
				switch n[0] {
				case "counter":
					isCounter = true
				case "marker":
					isMarker = true
				default:
					log.Fatalf("Invalid field option: %s", n[0])
				}
				n = n[1:]
			}
			name = n[0]
			if strings.HasPrefix(name, "counter:") {
				isCounter = true
				name = name[8:]
			}
			if strings.HasPrefix(name, "marker:") {
				isMarker = true
				name = name[7:]
			}

			gs.Fields = append(gs.Fields, data.Field{
				ID:      fmt.Sprintf("%d.%d.%s", i, j, name),
				Name:    name,
				Counter: isCounter,
				Marker:  isMarker,
			})
		}
		specs = append(specs, gs)
	}
	return specs
}