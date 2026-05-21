package cmd

import (
	"fmt"
	"mall-admin/internal/db"
	"os"

	"github.com/spf13/cobra"
	"gorm.io/gorm"
)

const defaultDBPath = "/Users/yangzhongyu/Desktop/code/github/second-hand-mall/mall-server/second-hand-mall.db"
const version = "1.0.0"

var (
	dbPath string
	DB     *gorm.DB
)

var rootCmd = &cobra.Command{
	Use:   "mall-admin",
	Short: "CLI admin tool for the second-hand mall platform",
	Long: `mall-admin provides command-line access to the second-hand mall SQLite database.
Use it to query and manage users and product listings.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Skip DB init for version/help commands
		if cmd.Name() == "version" || cmd.Name() == "help" {
			return nil
		}
		path := dbPath
		if path == "" {
			if env := os.Getenv("MALL_DB_PATH"); env != "" {
				path = env
			} else {
				path = defaultDBPath
			}
		}
		var err error
		DB, err = db.Open(path)
		if err != nil {
			return fmt.Errorf("failed to open database at %q: %w", path, err)
		}
		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	// /Users/yangzhongyu/Desktop/code/github/second-hand-mall/mall-server/second-hand-mall.db
	rootCmd.PersistentFlags().StringVar(&dbPath, "db-path", "", "path to SQLite database file (overrides MALL_DB_PATH env var)")
	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print the version of mall-admin",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("mall-admin v%s\n", version)
		},
	})
}
